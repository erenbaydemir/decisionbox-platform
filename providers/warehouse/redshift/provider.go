// Package redshift provides a warehouse.Provider for Amazon Redshift.
// Supports both Serverless (workgroup) and Provisioned (cluster) via the
// Redshift Data API — same API, different identifier parameter.
//
// Configuration:
//
//	WAREHOUSE_PROVIDER=redshift
//	Serverless: workgroup + database + region in project config
//	Provisioned: cluster_id + database + db_user + region in project config
//
// Authentication: AWS credentials (IAM role, env vars, or ~/.aws/credentials).
// Cross-cloud: store AWS credentials in secret provider (warehouse-credentials).
package redshift

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	gowarehouse "github.com/decisionbox-io/decisionbox/libs/go-common/warehouse"
)

func init() {
	gowarehouse.RegisterWithMeta("redshift", func(cfg gowarehouse.ProviderConfig) (gowarehouse.Provider, error) {
		region := cfg["region"]
		if region == "" {
			region = "us-east-1"
		}
		database := cfg["database"]
		if database == "" {
			database = "dev"
		}

		timeoutMin, _ := strconv.Atoi(cfg["timeout_minutes"])
		if timeoutMin == 0 {
			timeoutMin = 5
		}

		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
		)
		if err != nil {
			return nil, fmt.Errorf("redshift: failed to load AWS config: %w", err)
		}

		client := redshiftdata.NewFromConfig(awsCfg)

		return &RedshiftProvider{
			client:    client,
			workgroup: cfg["workgroup"],
			clusterID: cfg["cluster_id"],
			database:  database,
			dbUser:    cfg["db_user"],
			dataset:   cfg["dataset"],
			timeout:   time.Duration(timeoutMin) * time.Minute,
		}, nil
	}, gowarehouse.ProviderMeta{
		Name:        "Amazon Redshift",
		Description: "AWS cloud data warehouse — Serverless or Provisioned",
		ConfigFields: []gowarehouse.ConfigField{
			{Key: "workgroup", Label: "Workgroup Name (Serverless)", Type: "string", Placeholder: "default-workgroup", Description: "For Redshift Serverless. Leave empty for provisioned clusters."},
			{Key: "cluster_id", Label: "Cluster ID (Provisioned)", Type: "string", Placeholder: "my-cluster", Description: "For provisioned clusters. Leave empty for Serverless."},
			{Key: "database", Label: "Database", Required: true, Type: "string", Default: "dev"},
			{Key: "dataset", Label: "Schema", Type: "string", Default: "public", Description: "Redshift schema (equivalent to BigQuery dataset)"},
			{Key: "db_user", Label: "Database User (Provisioned only)", Type: "string", Description: "Required for provisioned clusters. Not needed for Serverless."},
			{Key: "region", Label: "AWS Region", Type: "string", Default: "us-east-1"},
		},
		DefaultPricing: &gowarehouse.WarehousePricing{
			CostModel:           "per_hour",
			CostPerTBScannedUSD: 0, // Redshift pricing is per-RPU-hour, not per-byte
		},
	})
}

// RedshiftProvider implements warehouse.Provider using the Redshift Data API.
type RedshiftProvider struct {
	client    *redshiftdata.Client
	workgroup string // Serverless
	clusterID string // Provisioned
	database  string
	dbUser    string
	dataset   string // schema name (default: "public")
	timeout   time.Duration
}

func (p *RedshiftProvider) Query(ctx context.Context, query string, params map[string]interface{}) (*gowarehouse.QueryResult, error) {
	input := &redshiftdata.ExecuteStatementInput{
		Database: aws.String(p.database),
		Sql:      aws.String(query),
	}

	// Route to Serverless or Provisioned based on config
	if p.workgroup != "" {
		input.WorkgroupName = aws.String(p.workgroup)
	} else if p.clusterID != "" {
		input.ClusterIdentifier = aws.String(p.clusterID)
		if p.dbUser != "" {
			input.DbUser = aws.String(p.dbUser)
		}
	} else {
		return nil, fmt.Errorf("redshift: either workgroup (Serverless) or cluster_id (Provisioned) is required")
	}

	// Execute (async — returns immediately)
	execOutput, err := p.client.ExecuteStatement(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("redshift: ExecuteStatement failed: %w", err)
	}

	stmtID := aws.ToString(execOutput.Id)

	// Poll until complete
	if err := p.waitForCompletion(ctx, stmtID); err != nil {
		return nil, err
	}

	// Get results
	return p.getResults(ctx, stmtID)
}

// waitForCompletion polls DescribeStatement until the query finishes.
func (p *RedshiftProvider) waitForCompletion(ctx context.Context, stmtID string) error {
	deadline := time.Now().Add(p.timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("redshift: query timed out after %s", p.timeout)
		}

		desc, err := p.client.DescribeStatement(ctx, &redshiftdata.DescribeStatementInput{
			Id: aws.String(stmtID),
		})
		if err != nil {
			return fmt.Errorf("redshift: DescribeStatement failed: %w", err)
		}

		switch desc.Status {
		case types.StatusStringFinished:
			return nil
		case types.StatusStringFailed:
			return fmt.Errorf("redshift: query failed: %s", aws.ToString(desc.Error))
		case types.StatusStringAborted:
			return fmt.Errorf("redshift: query aborted")
		}

		// Poll interval: 200ms
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// getResults fetches query results with pagination.
func (p *RedshiftProvider) getResults(ctx context.Context, stmtID string) (*gowarehouse.QueryResult, error) {
	var columns []string
	var rows []map[string]interface{}

	paginator := redshiftdata.NewGetStatementResultPaginator(p.client, &redshiftdata.GetStatementResultInput{
		Id: aws.String(stmtID),
	})

	firstPage := true
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("redshift: GetStatementResult failed: %w", err)
		}

		// Extract column names from first page
		if firstPage {
			columns = make([]string, len(page.ColumnMetadata))
			for i, col := range page.ColumnMetadata {
				columns[i] = aws.ToString(col.Name)
			}
			firstPage = false
		}

		// Convert rows
		for _, record := range page.Records {
			row := make(map[string]interface{})
			for i, field := range record {
				if i < len(columns) {
					row[columns[i]] = extractFieldValue(field)
				}
			}
			rows = append(rows, row)
		}
	}

	return &gowarehouse.QueryResult{
		Columns: columns,
		Rows:    rows,
	}, nil
}

// extractFieldValue converts a Redshift Data API Field to a Go value.
func extractFieldValue(field types.Field) interface{} {
	switch v := field.(type) {
	case *types.FieldMemberStringValue:
		return v.Value
	case *types.FieldMemberLongValue:
		return v.Value
	case *types.FieldMemberDoubleValue:
		return v.Value
	case *types.FieldMemberBooleanValue:
		return v.Value
	case *types.FieldMemberIsNull:
		return nil
	case *types.FieldMemberBlobValue:
		return v.Value
	default:
		return nil
	}
}

func (p *RedshiftProvider) ListTables(ctx context.Context) ([]string, error) {
	return p.ListTablesInDataset(ctx, p.dataset)
}

func (p *RedshiftProvider) ListTablesInDataset(ctx context.Context, dataset string) ([]string, error) {
	if dataset == "" {
		dataset = "public"
	}

	input := &redshiftdata.ListTablesInput{
		Database:      aws.String(p.database),
		SchemaPattern: aws.String(dataset),
	}
	if p.workgroup != "" {
		input.WorkgroupName = aws.String(p.workgroup)
	} else if p.clusterID != "" {
		input.ClusterIdentifier = aws.String(p.clusterID)
		if p.dbUser != "" {
			input.DbUser = aws.String(p.dbUser)
		}
	}

	var tables []string
	paginator := redshiftdata.NewListTablesPaginator(p.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("redshift: ListTables failed: %w", err)
		}
		for _, t := range page.Tables {
			name := aws.ToString(t.Name)
			// Skip system tables
			if !strings.HasPrefix(name, "pg_") && !strings.HasPrefix(name, "stl_") && !strings.HasPrefix(name, "svv_") {
				tables = append(tables, name)
			}
		}
	}

	return tables, nil
}

func (p *RedshiftProvider) GetTableSchema(ctx context.Context, table string) (*gowarehouse.TableSchema, error) {
	return p.GetTableSchemaInDataset(ctx, p.dataset, table)
}

func (p *RedshiftProvider) GetTableSchemaInDataset(ctx context.Context, dataset, table string) (*gowarehouse.TableSchema, error) {
	if dataset == "" {
		dataset = "public"
	}

	input := &redshiftdata.DescribeTableInput{
		Database:      aws.String(p.database),
		Schema:        aws.String(dataset),
		Table:         aws.String(table),
	}
	if p.workgroup != "" {
		input.WorkgroupName = aws.String(p.workgroup)
	} else if p.clusterID != "" {
		input.ClusterIdentifier = aws.String(p.clusterID)
		if p.dbUser != "" {
			input.DbUser = aws.String(p.dbUser)
		}
	}

	result, err := p.client.DescribeTable(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("redshift: DescribeTable failed: %w", err)
	}

	schema := &gowarehouse.TableSchema{
		Name: table,
	}
	for _, col := range result.ColumnList {
		schema.Columns = append(schema.Columns, gowarehouse.ColumnSchema{
			Name:     aws.ToString(col.Name),
			Type:     normalizeRedshiftType(aws.ToString(col.TypeName)),
			Nullable: col.Nullable != 0,
		})
	}

	// Get row count via query
	countQuery := fmt.Sprintf("SELECT COUNT(*) as cnt FROM %s.%s", dataset, table)
	countResult, err := p.Query(ctx, countQuery, nil)
	if err == nil && len(countResult.Rows) > 0 {
		if cnt, ok := countResult.Rows[0]["cnt"]; ok {
			switch v := cnt.(type) {
			case int64:
				schema.RowCount = v
			case float64:
				schema.RowCount = int64(v)
			}
		}
	}

	return schema, nil
}

func (p *RedshiftProvider) GetDataset() string {
	if p.dataset != "" {
		return p.dataset
	}
	return "public"
}

func (p *RedshiftProvider) SQLDialect() string {
	return "Amazon Redshift SQL (PostgreSQL-compatible)"
}

func (p *RedshiftProvider) SQLFixPrompt() string {
	return "" // TODO: add Redshift-specific SQL fix prompt
}

func (p *RedshiftProvider) ValidateReadOnly(ctx context.Context) error {
	// Redshift Data API with IAM auth is read-only by default
	// unless the IAM policy allows write operations
	return nil
}

func (p *RedshiftProvider) HealthCheck(ctx context.Context) error {
	_, err := p.Query(ctx, "SELECT 1", nil)
	return err
}

func (p *RedshiftProvider) Close() error {
	return nil // Data API is stateless, no connection to close
}

// normalizeRedshiftType maps Redshift types to warehouse-agnostic types.
func normalizeRedshiftType(t string) string {
	t = strings.ToLower(t)
	switch {
	case t == "integer" || t == "int" || t == "int4" || t == "bigint" || t == "int8" || t == "smallint" || t == "int2":
		return "INT64"
	case t == "real" || t == "float4" || t == "double precision" || t == "float8" || t == "float" || strings.HasPrefix(t, "numeric") || strings.HasPrefix(t, "decimal"):
		return "FLOAT64"
	case t == "boolean" || t == "bool":
		return "BOOL"
	case t == "date":
		return "DATE"
	case strings.Contains(t, "timestamp"):
		return "TIMESTAMP"
	case t == "bytea":
		return "BYTES"
	default:
		return "STRING"
	}
}
