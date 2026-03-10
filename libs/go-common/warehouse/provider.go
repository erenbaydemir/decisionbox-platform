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
