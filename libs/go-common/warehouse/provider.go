package warehouse

import "context"

// Provider abstracts data warehouse query operations.
// Implement this interface to add support for a new data warehouse
// (e.g., ClickHouse, Redshift, Snowflake, DuckDB).
//
// The BigQuery implementation is provided in warehouse/bigquery/.
//
// Selection via WAREHOUSE_PROVIDER env var (e.g., "bigquery").
type Provider interface {
	// Query executes a SQL query and returns results.
	Query(ctx context.Context, query string, params map[string]interface{}) (*QueryResult, error)

	// ListTables returns all table names in the configured dataset/schema.
	ListTables(ctx context.Context) ([]string, error)

	// GetTableSchema returns schema metadata for a specific table.
	GetTableSchema(ctx context.Context, table string) (*TableSchema, error)

	// GetDataset returns the dataset/schema name being queried.
	GetDataset() string

	// SQLDialect returns the SQL dialect name for this warehouse.
	// Used by the discovery agent to give the LLM context about syntax.
	//   BigQuery: "BigQuery Standard SQL"
	//   PostgreSQL: "PostgreSQL"
	//   ClickHouse: "ClickHouse SQL"
	SQLDialect() string

	// SQLFixPrompt returns a prompt template for fixing SQL errors in this
	// warehouse's dialect. The prompt contains common error patterns and
	// syntax rules specific to this warehouse.
	//
	// Templates use placeholders: {{DATASET}}, {{FILTER}}, {{SCHEMA_INFO}},
	// {{ORIGINAL_SQL}}, {{ERROR_MESSAGE}}, {{CONVERSATION_HISTORY}}.
	//
	// Returns empty string if no warehouse-specific prompt is available.
	SQLFixPrompt() string

	// HealthCheck verifies the warehouse connection is alive.
	HealthCheck(ctx context.Context) error

	// Close releases warehouse resources.
	Close() error
}

// QueryResult holds the result of a warehouse query.
type QueryResult struct {
	Columns []string
	Rows    []map[string]interface{}
}

// TableSchema describes a table's structure in a warehouse-agnostic way.
type TableSchema struct {
	Name     string
	Columns  []ColumnSchema
	RowCount int64
}

// ColumnSchema describes a single column in a warehouse-agnostic way.
type ColumnSchema struct {
	Name     string
	Type     string // Normalized type: "STRING", "INT64", "FLOAT64", "BOOL", "TIMESTAMP", "DATE", "BYTES", "RECORD"
	Nullable bool
}
