package gaming

import (
	"testing"

	"github.com/decisionbox-io/decisionbox/libs/go-common/domainpack"
)

func TestGamingPack_Registered(t *testing.T) {
	packs := domainpack.RegisteredPacks()
	found := false
	for _, p := range packs {
		if p == "gaming" {
			found = true
		}
	}
	if !found {
		t.Fatal("gaming not registered")
	}
}

func TestGamingPack_Get(t *testing.T) {
	pack, err := domainpack.Get("gaming")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if pack.Name() != "gaming" {
		t.Errorf("Name() = %q", pack.Name())
	}
}

func TestGamingPack_EntityTypes(t *testing.T) {
	pack := NewPack()
	types := pack.EntityTypes()
	expected := map[string]bool{"users": true, "entities": true, "sessions": true, "cohorts": true}
	for _, et := range types {
		if !expected[et] {
			t.Errorf("Unexpected entity type: %q", et)
		}
	}
	if len(types) != 4 {
		t.Errorf("Expected 4 entity types, got %d", len(types))
	}
}

func TestGamingPack_SemanticGenerator(t *testing.T) {
	pack := NewPack()
	gen := pack.SemanticGenerator()
	if gen == nil {
		t.Fatal("SemanticGenerator() returned nil")
	}
}

func TestGamingPack_DataFetcher(t *testing.T) {
	pack := NewPack()
	fetcher := pack.DataFetcher()
	if fetcher == nil {
		t.Fatal("DataFetcher() returned nil")
	}
}

func TestSemanticGenerator_Users(t *testing.T) {
	gen := &GamingSemanticGenerator{}
	text := gen.GenerateText("users", map[string]interface{}{
		"total_session_count":    100,
		"user_max_level_reached": 45,
		"user_success_rate":      0.72,
		"is_payer":               true,
		"total_spend":            49.99,
		"platform":               "iOS",
		"country_code":           "US",
		"user_tenure_days":       int64(90),
		"days_since_last_event":  int64(2),
	})
	if text == "" {
		t.Fatal("Empty text for user profile")
	}
	// Should contain classification terms from the original logic
	t.Logf("Generated: %s", text)
}

func TestSemanticGenerator_Users_Empty(t *testing.T) {
	gen := &GamingSemanticGenerator{}
	text := gen.GenerateText("users", map[string]interface{}{})
	if text == "" {
		t.Fatal("Should produce text even with empty data")
	}
}

func TestSemanticGenerator_Entities(t *testing.T) {
	gen := &GamingSemanticGenerator{}
	text := gen.GenerateText("entities", map[string]interface{}{
		"level_number":           10,
		"unique_players":         500,
		"total_starts":           2000,
		"total_successes":        1400,
		"total_failures":         600,
		"success_rate":           0.7,
		"quit_rate":              0.15,
		"avg_attempts_per_player": 2.5,
	})
	if text == "" {
		t.Fatal("Empty text for entity profile")
	}
}

func TestSemanticGenerator_Sessions(t *testing.T) {
	gen := &GamingSemanticGenerator{}
	text := gen.GenerateText("sessions", map[string]interface{}{
		"session_id": "s123",
		"user_id":    "u456",
		"duration":   float64(1800),
		"platform":   "Android",
	})
	if text == "" {
		t.Fatal("Empty text for session profile")
	}
}

func TestSemanticGenerator_Cohorts(t *testing.T) {
	gen := &GamingSemanticGenerator{}
	text := gen.GenerateText("cohorts", map[string]interface{}{
		"cohort_date":     "2026-01-15",
		"cohort_size":     1000,
		"day_1_retention":  0.45,
		"day_7_retention":  0.25,
		"day_30_retention": 0.10,
		"country_code":    "US",
	})
	if text == "" {
		t.Fatal("Empty text for cohort profile")
	}
}

func TestSemanticGenerator_Unsupported(t *testing.T) {
	gen := &GamingSemanticGenerator{}
	text := gen.GenerateText("unknown_type", map[string]interface{}{"x": 1})
	if text != "" {
		t.Errorf("Expected empty text for unsupported type, got: %q", text)
	}
}

func TestDataFetcher_DatapointID_Users(t *testing.T) {
	f := NewGamingDataFetcher("/tmp", "bigquery")
	id := f.GenerateDatapointID("users", "app1", map[string]interface{}{"user_id": "u123"})
	if id != "app_app1_user_u123" {
		t.Errorf("got %q", id)
	}
}

func TestDataFetcher_DatapointID_Entities(t *testing.T) {
	f := NewGamingDataFetcher("/tmp", "bigquery")
	id := f.GenerateDatapointID("entities", "app1", map[string]interface{}{"level_number": 5})
	if id != "app_app1_entity_5" {
		t.Errorf("got %q", id)
	}
}

func TestDataFetcher_DatapointID_Sessions(t *testing.T) {
	f := NewGamingDataFetcher("/tmp", "bigquery")
	id := f.GenerateDatapointID("sessions", "app1", map[string]interface{}{"session_id": "s789"})
	if id != "app_app1_session_s789" {
		t.Errorf("got %q", id)
	}
}

func TestDataFetcher_DatapointID_Cohorts(t *testing.T) {
	f := NewGamingDataFetcher("/tmp", "bigquery")
	id := f.GenerateDatapointID("cohorts", "app1", map[string]interface{}{
		"cohort_date":  "2026-01-15",
		"country_code": "US",
	})
	if id != "app_app1_cohort_2026-01-15_US" {
		t.Errorf("got %q", id)
	}
}

func TestDataFetcher_CollectionName(t *testing.T) {
	f := NewGamingDataFetcher("/tmp", "bigquery")
	tests := map[string]string{
		"users":    "embedding_user_profiles",
		"entities": "embedding_entity_profiles",
		"sessions": "embedding_session_profiles",
		"cohorts":  "embedding_cohort_profiles",
		"custom":   "embedding_custom_profiles",
	}
	for entityType, expected := range tests {
		if got := f.CollectionName(entityType); got != expected {
			t.Errorf("CollectionName(%q) = %q, want %q", entityType, got, expected)
		}
	}
}

func TestDataFetcher_FetchQuery_NoFile(t *testing.T) {
	f := NewGamingDataFetcher("/nonexistent/path", "bigquery")
	query := f.FetchQuery("users", "ds", "app1", 30)
	if query != "" {
		t.Errorf("Expected empty query for missing file, got: %q", query)
	}
}

func TestDataFetcher_SupportedEntityTypes(t *testing.T) {
	f := NewGamingDataFetcher("/tmp", "bigquery")
	types := f.SupportedEntityTypes()
	if len(types) != 4 {
		t.Errorf("Expected 4 types, got %d", len(types))
	}
}
