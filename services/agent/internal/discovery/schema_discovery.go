package discovery

import (
	"context"
	"fmt"
	"time"

	gowarehouse "github.com/decisionbox-io/decisionbox/libs/go-common/warehouse"
	logger "github.com/decisionbox-io/decisionbox/services/agent/internal/log"
	"github.com/decisionbox-io/decisionbox/services/agent/internal/queryexec"
	"github.com/decisionbox-io/decisionbox/services/agent/internal/models"
)

// SchemaDiscovery discovers and analyzes warehouse table schemas.
type SchemaDiscovery struct {
	warehouse gowarehouse.Provider
	executor  *queryexec.QueryExecutor
	projectID string
	dataset   string
	filter    string // e.g., "WHERE app_id = 'xyz'" or ""
}

// SchemaDiscoveryOptions configures schema discovery.
type SchemaDiscoveryOptions struct {
	Warehouse gowarehouse.Provider
	Executor  *queryexec.QueryExecutor
	ProjectID string
	Dataset   string
	Filter    string
}

// NewSchemaDiscovery creates a new schema discovery service.
func NewSchemaDiscovery(opts SchemaDiscoveryOptions) *SchemaDiscovery {
	return &SchemaDiscovery{
		warehouse: opts.Warehouse,
		executor:  opts.Executor,
		projectID: opts.ProjectID,
		dataset:   opts.Dataset,
		filter:    opts.Filter,
	}
}

// DiscoverSchemas discovers all tables and their schemas for the app
func (s *SchemaDiscovery) DiscoverSchemas(ctx context.Context) (map[string]models.TableSchema, error) {
	logger.Info("Discovering BigQuery table schemas")

	// List all tables in the dataset
	tables, err := s.warehouse.ListTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	logger.WithField("table_count", len(tables)).Info("Found tables in dataset")

	schemas := make(map[string]models.TableSchema)

	// For each table, get detailed schema
	for _, tableName := range tables {
		logger.WithField("table", tableName).Debug("Discovering schema for table")

		schema, err := s.DiscoverTableSchema(ctx, tableName)
		if err != nil {
			logger.WithFields(logger.Fields{"error": err.Error(), "table": tableName}).Warn("Failed to discover table schema, skipping")
			continue
		}

		schemas[tableName] = *schema
	}

	logger.WithField("schemas_discovered", len(schemas)).Info("Schema discovery complete")

	return schemas, nil
}

// DiscoverTableSchema discovers the schema for a specific table
func (s *SchemaDiscovery) DiscoverTableSchema(ctx context.Context, tableName string) (*models.TableSchema, error) {
	schema := &models.TableSchema{
		TableName:    tableName,
		Columns:      make([]models.ColumnInfo, 0),
		KeyColumns:   make([]string, 0),
		Metrics:      make([]string, 0),
		Dimensions:   make([]string, 0),
		DiscoveredAt: time.Now(),
	}

	// Get table schema from warehouse
	whSchema, err := s.warehouse.GetTableSchema(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}

	// Extract schema information
	for _, col := range whSchema.Columns {
		colInfo := models.ColumnInfo{
			Name:     col.Name,
			Type:     col.Type,
			Nullable: col.Nullable,
			Category: s.inferColumnCategory(col.Name, col.Type),
		}
		schema.Columns = append(schema.Columns, colInfo)
		s.categorizeColumn(&colInfo, schema)
	}

	// Get row count
	schema.RowCount = whSchema.RowCount

	// Get sample data (limit 10 rows)
	sampleData, err := s.getSampleData(ctx, tableName)
	if err != nil {
		logger.WithField("error", err.Error()).Warn("Failed to get sample data")
	} else {
		schema.SampleData = sampleData
	}

	logger.WithFields(logger.Fields{
		"table":       tableName,
		"columns":     len(schema.Columns),
		"row_count":   schema.RowCount,
		"key_columns": len(schema.KeyColumns),
		"metrics":     len(schema.Metrics),
		"dimensions":  len(schema.Dimensions),
	}).Info("Table schema discovered")

	return schema, nil
}

// inferColumnCategory infers the category of a column based on name and type
func (s *SchemaDiscovery) inferColumnCategory(name string, fieldType string) string {
	nameLower := name

	// Primary key detection
	if nameLower == "id" || nameLower == "user_id" || nameLower == "player_id" ||
		nameLower == "session_id" || nameLower == "event_id" {
		return "primary_key"
	}

	// Time column detection
	if nameLower == "created_at" || nameLower == "updated_at" || nameLower == "timestamp" ||
		nameLower == "start_time" || nameLower == "end_time" || nameLower == "date" ||
		fieldType == "TIMESTAMP" || fieldType == "DATE" || fieldType == "DATETIME" {
		return "time"
	}

	// Metric detection (numeric columns)
	if fieldType == "INT64" || fieldType == "FLOAT64" || fieldType == "NUMERIC" || fieldType == "BIGNUMERIC" {
		// Numeric columns are typically metrics unless they're IDs
		if nameLower == "id" || nameLower == "user_id" || nameLower == "player_id" {
			return "dimension" // IDs are dimensions, not metrics
		}
		return "metric"
	}

	// Everything else is a dimension
	return "dimension"
}

// categorizeColumn categorizes a column into metrics, dimensions, or key columns
func (s *SchemaDiscovery) categorizeColumn(col *models.ColumnInfo, schema *models.TableSchema) {
	switch col.Category {
	case "primary_key":
		schema.KeyColumns = append(schema.KeyColumns, col.Name)
	case "metric":
		schema.Metrics = append(schema.Metrics, col.Name)
	case "dimension":
		schema.Dimensions = append(schema.Dimensions, col.Name)
	case "time":
		// Time columns are both dimensions and tracked separately
		schema.Dimensions = append(schema.Dimensions, col.Name)
	}
}

// getSampleData gets sample rows from the table
func (s *SchemaDiscovery) getSampleData(ctx context.Context, tableName string) ([]map[string]interface{}, error) {
	filterClause := ""
	if s.filter != "" {
		filterClause = s.filter
	}
	query := fmt.Sprintf("SELECT * FROM `%s.%s` %s LIMIT 10", s.dataset, tableName, filterClause)

	// Use executor to run query (with error handling)
	result, err := s.executor.Execute(ctx, query, "Get sample data for "+tableName)
	if err != nil {
		// If this fails, it's not critical - we can still discover schema
		return nil, err
	}

	return result.Data, nil
}

// GetSchemaAsJSON returns the schemas as a JSON string for Claude
func (s *SchemaDiscovery) GetSchemaAsJSON(schemas map[string]models.TableSchema) (string, error) {
	// Build a simplified version for Claude
	simplified := make(map[string]interface{})

	for tableName, schema := range schemas {
		tableInfo := map[string]interface{}{
			"table_name":  tableName,
			"row_count":   schema.RowCount,
			"key_columns": schema.KeyColumns,
			"metrics":     schema.Metrics,
			"dimensions":  schema.Dimensions,
			"columns":     make([]map[string]string, 0),
		}

		for _, col := range schema.Columns {
			tableInfo["columns"] = append(tableInfo["columns"].([]map[string]string), map[string]string{
				"name":     col.Name,
				"type":     col.Type,
				"category": col.Category,
			})
		}

		// Add sample data if available (just first 3 rows)
		if len(schema.SampleData) > 0 {
			sampleCount := 3
			if len(schema.SampleData) < sampleCount {
				sampleCount = len(schema.SampleData)
			}
			tableInfo["sample_data"] = schema.SampleData[:sampleCount]
		}

		simplified[tableName] = tableInfo
	}

	// Convert to JSON
	// Note: In production, use json.MarshalIndent for better formatting
	// For now, we'll return a formatted string representation
	return fmt.Sprintf("%+v", simplified), nil
}

// InspectTable provides a detailed inspection of a table for Claude
func (s *SchemaDiscovery) InspectTable(ctx context.Context, tableName string) (string, error) {
	schema, err := s.DiscoverTableSchema(ctx, tableName)
	if err != nil {
		return "", err
	}

	// Build inspection report
	report := fmt.Sprintf("Table: %s\n", tableName)
	report += fmt.Sprintf("Rows: %d\n", schema.RowCount)
	report += fmt.Sprintf("Columns: %d\n\n", len(schema.Columns))

	report += "Key Columns:\n"
	for _, col := range schema.KeyColumns {
		report += fmt.Sprintf("  - %s\n", col)
	}

	report += fmt.Sprintf("\nMetrics (%d):\n", len(schema.Metrics))
	for _, col := range schema.Metrics {
		report += fmt.Sprintf("  - %s\n", col)
	}

	report += fmt.Sprintf("\nDimensions (%d):\n", len(schema.Dimensions))
	for _, col := range schema.Dimensions {
		report += fmt.Sprintf("  - %s\n", col)
	}

	report += "\nAll Columns:\n"
	for _, col := range schema.Columns {
		report += fmt.Sprintf("  - %s (%s) [%s]\n", col.Name, col.Type, col.Category)
	}

	if len(schema.SampleData) > 0 {
		report += fmt.Sprintf("\nSample Data (%d rows):\n", len(schema.SampleData))
		for i, row := range schema.SampleData {
			if i >= 3 {
				break // Show only first 3 rows
			}
			report += fmt.Sprintf("  Row %d: %+v\n", i+1, row)
		}
	}

	return report, nil
}
