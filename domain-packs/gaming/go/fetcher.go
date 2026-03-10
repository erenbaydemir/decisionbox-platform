package gaming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// queryFileMap maps entity types to SQL file names (without dialect prefix).
var queryFileMap = map[string]string{
	"users":    "user_profiles.sql",
	"entities": "level_performance.sql",
	"sessions": "session_patterns.sql",
	"cohorts":  "cohort_retention.sql",
}

// GamingDataFetcher implements domainpack.DataFetcher for gaming.
// Loads SQL queries from the filesystem at runtime, selecting the
// correct dialect subdirectory based on the warehouse provider.
//
// Query path resolution:
//   {queriesPath}/{warehouseDialect}/{filename}.sql
//
// The queriesPath is set via:
//   1. Constructor argument (NewGamingDataFetcher)
//   2. DOMAIN_PACK_QUERIES_PATH env var
//   3. Default: "./domain-packs/gaming/queries"
type GamingDataFetcher struct {
	queriesPath      string
	warehouseDialect string
}

// NewGamingDataFetcher creates a data fetcher that loads SQL from the given path.
// warehouseDialect selects the SQL subdirectory (e.g., "bigquery", "clickhouse").
func NewGamingDataFetcher(queriesPath, warehouseDialect string) *GamingDataFetcher {
	if queriesPath == "" {
		queriesPath = os.Getenv("DOMAIN_PACK_QUERIES_PATH")
	}
	if queriesPath == "" {
		queriesPath = "./domain-packs/gaming/queries"
	}
	if warehouseDialect == "" {
		warehouseDialect = "bigquery"
	}
	return &GamingDataFetcher{
		queriesPath:      queriesPath,
		warehouseDialect: warehouseDialect,
	}
}

func (f *GamingDataFetcher) SupportedEntityTypes() []string {
	return []string{"users", "entities", "sessions", "cohorts"}
}

// GenerateDatapointID creates a unique datapoint ID for gaming entities.
// Gaming uses: user_id for users, level_number for entities (levels),
// session_id for sessions, cohort_date+country_code for cohorts.
func (f *GamingDataFetcher) GenerateDatapointID(entityType, appID string, data map[string]interface{}) string {
	switch entityType {
	case "users":
		return fmt.Sprintf("app_%s_user_%s", appID, toString(data["user_id"]))
	case "entities":
		return fmt.Sprintf("app_%s_entity_%v", appID, data["level_number"])
	case "sessions":
		return fmt.Sprintf("app_%s_session_%s", appID, toString(data["session_id"]))
	case "cohorts":
		country := toString(data["country_code"])
		if country == "" {
			country = "all"
		}
		return fmt.Sprintf("app_%s_cohort_%s_%s", appID, toString(data["cohort_date"]), country)
	default:
		return fmt.Sprintf("app_%s_%s_%d", appID, entityType, len(data))
	}
}

// collectionMap maps entity types to MongoDB collection names for gaming.
var collectionMap = map[string]string{
	"users":    "embedding_user_profiles",
	"entities": "embedding_entity_profiles",
	"sessions": "embedding_session_profiles",
	"cohorts":  "embedding_cohort_profiles",
}

// CollectionName returns the MongoDB collection name for a gaming entity type.
func (f *GamingDataFetcher) CollectionName(entityType string) string {
	if name, ok := collectionMap[entityType]; ok {
		return name
	}
	return "embedding_" + entityType + "_profiles"
}

// FetchQuery loads the SQL template for the entity type, renders parameters,
// and returns the ready-to-execute query.
func (f *GamingDataFetcher) FetchQuery(entityType, dataset, appID string, lookbackDays int) string {
	filename, ok := queryFileMap[entityType]
	if !ok {
		return ""
	}

	sqlPath := filepath.Join(f.queriesPath, f.warehouseDialect, filename)
	data, err := os.ReadFile(sqlPath)
	if err != nil {
		return ""
	}

	return renderQuery(string(data), dataset, appID, lookbackDays)
}

// renderQuery replaces template parameters in a SQL query.
func renderQuery(query, dataset, appID string, lookbackDays int) string {
	query = strings.ReplaceAll(query, "{{dataset}}", dataset)
	query = strings.ReplaceAll(query, "{{app_id}}", appID)
	query = strings.ReplaceAll(query, "{{lookback_days}}", fmt.Sprintf("%d", lookbackDays))

	// Handle conditional blocks: {{#if lookback_days}}...{{/if}}
	if lookbackDays > 0 {
		query = strings.ReplaceAll(query, "{{#if lookback_days}}", "")
		query = strings.ReplaceAll(query, "{{/if}}", "")
	} else {
		for {
			start := strings.Index(query, "{{#if lookback_days}}")
			if start == -1 {
				break
			}
			end := strings.Index(query[start:], "{{/if}}")
			if end == -1 {
				break
			}
			query = query[:start] + query[start+end+len("{{/if}}"):]
		}
	}

	return query
}
