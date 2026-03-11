package discovery

import (
	"context"
	"fmt"
	"time"

	gowarehouse "github.com/decisionbox-io/decisionbox/libs/go-common/warehouse"
	logger "github.com/decisionbox-io/decisionbox/services/agent/internal/log"
	"github.com/decisionbox-io/decisionbox/services/agent/internal/models"
	"github.com/decisionbox-io/decisionbox/services/agent/internal/queryexec"
)

// SchemaDiscovery discovers and analyzes warehouse table schemas
// across multiple datasets.
type SchemaDiscovery struct {
	warehouse gowarehouse.Provider
	executor  *queryexec.QueryExecutor
	projectID string
	datasets  []string // multiple datasets to discover
	filter    string
}

// SchemaDiscoveryOptions configures schema discovery.
type SchemaDiscoveryOptions struct {
	Warehouse gowarehouse.Provider
	Executor  *queryexec.QueryExecutor
	ProjectID string
	Datasets  []string
	Filter    string
}

// NewSchemaDiscovery creates a new schema discovery service.
func NewSchemaDiscovery(opts SchemaDiscoveryOptions) *SchemaDiscovery {
	return &SchemaDiscovery{
		warehouse: opts.Warehouse,
		executor:  opts.Executor,
		projectID: opts.ProjectID,
		datasets:  opts.Datasets,
		filter:    opts.Filter,
	}
}

// DiscoverSchemas discovers all tables across all configured datasets.
// Table keys are "dataset.table" for multi-dataset, or just "table" for single.
func (s *SchemaDiscovery) DiscoverSchemas(ctx context.Context) (map[string]models.TableSchema, error) {
	logger.WithField("datasets", s.datasets).Info("Discovering warehouse table schemas")

	allSchemas := make(map[string]models.TableSchema)

	for _, dataset := range s.datasets {
		logger.WithField("dataset", dataset).Info("Discovering schemas for dataset")

		// Query tables in this dataset using INFORMATION_SCHEMA
		tablesQuery := fmt.Sprintf(
			"SELECT table_name FROM `%s.INFORMATION_SCHEMA.TABLES` WHERE table_type = 'BASE TABLE'",
			dataset,
		)

		result, err := s.warehouse.Query(ctx, tablesQuery, nil)
		if err != nil {
			// Fallback: try using the provider's ListTables for the default dataset
			logger.WithFields(logger.Fields{"dataset": dataset, "error": err.Error()}).Warn("INFORMATION_SCHEMA query failed, trying ListTables fallback")
			tables, listErr := s.warehouse.ListTables(ctx)
			if listErr != nil {
				logger.WithFields(logger.Fields{"dataset": dataset, "error": listErr.Error()}).Warn("Failed to list tables, skipping dataset")
				continue
			}
			for _, tableName := range tables {
				schema, schemaErr := s.discoverTable(ctx, dataset, tableName)
				if schemaErr != nil {
					continue
				}
				key := fmt.Sprintf("%s.%s", dataset, tableName)
				allSchemas[key] = *schema
			}
			continue
		}

		for _, row := range result.Rows {
			tableName, ok := row["table_name"].(string)
			if !ok {
				continue
			}

			schema, err := s.discoverTable(ctx, dataset, tableName)
			if err != nil {
				logger.WithFields(logger.Fields{"table": tableName, "dataset": dataset, "error": err.Error()}).Warn("Failed to discover table, skipping")
				continue
			}

			// Key includes dataset: "dataset.table"
			key := fmt.Sprintf("%s.%s", dataset, tableName)
			allSchemas[key] = *schema
		}

		logger.WithFields(logger.Fields{
			"dataset": dataset,
			"tables":  len(allSchemas),
		}).Info("Dataset schema discovery complete")
	}

	logger.WithField("total_tables", len(allSchemas)).Info("All schema discovery complete")

	return allSchemas, nil
}

// discoverTable discovers the schema for a specific table in a dataset.
func (s *SchemaDiscovery) discoverTable(ctx context.Context, dataset, tableName string) (*models.TableSchema, error) {
	qualifiedName := fmt.Sprintf("%s.%s", dataset, tableName)

	schema := &models.TableSchema{
		TableName:    qualifiedName,
		Columns:      make([]models.ColumnInfo, 0),
		KeyColumns:   make([]string, 0),
		Metrics:      make([]string, 0),
		Dimensions:   make([]string, 0),
		DiscoveredAt: time.Now(),
	}

	// Get table schema via provider (use just tableName, provider knows the dataset)
	whSchema, err := s.warehouse.GetTableSchema(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("get schema: %w", err)
	}

	for _, col := range whSchema.Columns {
		colInfo := models.ColumnInfo{
			Name:     col.Name,
			Type:     col.Type,
			Nullable: col.Nullable,
			Category: inferColumnCategory(col.Name, col.Type),
		}
		schema.Columns = append(schema.Columns, colInfo)
		categorizeColumn(&colInfo, schema)
	}

	schema.RowCount = whSchema.RowCount

	// Get sample data
	sampleData, err := s.getSampleData(ctx, dataset, tableName)
	if err == nil {
		schema.SampleData = sampleData
	}

	logger.WithFields(logger.Fields{
		"table":   qualifiedName,
		"columns": len(schema.Columns),
		"rows":    schema.RowCount,
	}).Debug("Table schema discovered")

	return schema, nil
}

func (s *SchemaDiscovery) getSampleData(ctx context.Context, dataset, tableName string) ([]map[string]interface{}, error) {
	filterClause := ""
	if s.filter != "" {
		filterClause = s.filter
	}
	query := fmt.Sprintf("SELECT * FROM `%s.%s` %s LIMIT 5", dataset, tableName, filterClause)

	result, err := s.executor.Execute(ctx, query, "sample data for "+dataset+"."+tableName)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func inferColumnCategory(name string, fieldType string) string {
	if name == "id" || name == "user_id" || name == "player_id" ||
		name == "session_id" || name == "event_id" {
		return "primary_key"
	}
	if name == "created_at" || name == "updated_at" || name == "timestamp" ||
		name == "start_time" || name == "end_time" || name == "date" ||
		fieldType == "TIMESTAMP" || fieldType == "DATE" || fieldType == "DATETIME" {
		return "time"
	}
	if fieldType == "INT64" || fieldType == "FLOAT64" || fieldType == "NUMERIC" || fieldType == "BIGNUMERIC" ||
		fieldType == "INTEGER" || fieldType == "FLOAT" {
		if name == "id" || name == "user_id" || name == "player_id" {
			return "dimension"
		}
		return "metric"
	}
	return "dimension"
}

func categorizeColumn(col *models.ColumnInfo, schema *models.TableSchema) {
	switch col.Category {
	case "primary_key":
		schema.KeyColumns = append(schema.KeyColumns, col.Name)
	case "metric":
		schema.Metrics = append(schema.Metrics, col.Name)
	case "dimension", "time":
		schema.Dimensions = append(schema.Dimensions, col.Name)
	}
}
