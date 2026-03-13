package redshift

import (
	"testing"

	gowarehouse "github.com/decisionbox-io/decisionbox/libs/go-common/warehouse"
)

func TestRedshiftProvider_Registered(t *testing.T) {
	meta, ok := gowarehouse.GetProviderMeta("redshift")
	if !ok {
		t.Fatal("redshift not registered")
	}
	if meta.Name == "" {
		t.Error("missing provider name")
	}
	if meta.Description == "" {
		t.Error("missing description")
	}
}

func TestRedshiftProvider_ConfigFields(t *testing.T) {
	meta, _ := gowarehouse.GetProviderMeta("redshift")

	keys := make(map[string]bool)
	for _, f := range meta.ConfigFields {
		keys[f.Key] = true
	}
	for _, required := range []string{"workgroup", "cluster_id", "database", "dataset", "region"} {
		if !keys[required] {
			t.Errorf("missing config field: %s", required)
		}
	}
}

func TestRedshiftProvider_MissingIdentifier(t *testing.T) {
	p := &RedshiftProvider{
		database: "dev",
	}

	_, err := p.Query(nil, "SELECT 1", nil)
	if err == nil {
		t.Error("expected error when both workgroup and cluster_id are empty")
	}
}

func TestNormalizeRedshiftType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"integer", "INT64"},
		{"int", "INT64"},
		{"int4", "INT64"},
		{"bigint", "INT64"},
		{"int8", "INT64"},
		{"smallint", "INT64"},
		{"real", "FLOAT64"},
		{"double precision", "FLOAT64"},
		{"float8", "FLOAT64"},
		{"numeric(10,2)", "FLOAT64"},
		{"decimal(18,4)", "FLOAT64"},
		{"boolean", "BOOL"},
		{"bool", "BOOL"},
		{"date", "DATE"},
		{"timestamp", "TIMESTAMP"},
		{"timestamp without time zone", "TIMESTAMP"},
		{"timestamptz", "TIMESTAMP"},
		{"bytea", "BYTES"},
		{"varchar", "STRING"},
		{"character varying", "STRING"},
		{"text", "STRING"},
		{"char(10)", "STRING"},
		{"super", "STRING"},
	}
	for _, tt := range tests {
		got := normalizeRedshiftType(tt.input)
		if got != tt.want {
			t.Errorf("normalizeRedshiftType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRedshiftProvider_GetDataset_Default(t *testing.T) {
	p := &RedshiftProvider{}
	if p.GetDataset() != "public" {
		t.Errorf("default dataset = %q, want public", p.GetDataset())
	}
}

func TestRedshiftProvider_GetDataset_Custom(t *testing.T) {
	p := &RedshiftProvider{dataset: "analytics"}
	if p.GetDataset() != "analytics" {
		t.Errorf("dataset = %q, want analytics", p.GetDataset())
	}
}

func TestRedshiftProvider_SQLDialect(t *testing.T) {
	p := &RedshiftProvider{}
	if p.SQLDialect() == "" {
		t.Error("SQLDialect should not be empty")
	}
}

func TestExtractFieldValue_Nil(t *testing.T) {
	// nil field should not panic
	result := extractFieldValue(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}
