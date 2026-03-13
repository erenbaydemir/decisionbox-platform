package models

import "testing"

func TestGetDatasets_MultipleDatasets(t *testing.T) {
	w := WarehouseConfig{
		Datasets: []string{"events_prod", "features_prod"},
	}
	ds := w.GetDatasets()
	if len(ds) != 2 {
		t.Errorf("len = %d, want 2", len(ds))
	}
	if ds[0] != "events_prod" {
		t.Errorf("ds[0] = %q", ds[0])
	}
}

func TestGetDatasets_Empty(t *testing.T) {
	w := WarehouseConfig{}
	ds := w.GetDatasets()
	if ds != nil {
		t.Errorf("should return nil for empty config, got %v", ds)
	}
}

func TestGetDatasets_SingleInArray(t *testing.T) {
	w := WarehouseConfig{
		Datasets: []string{"only_one"},
	}
	ds := w.GetDatasets()
	if len(ds) != 1 || ds[0] != "only_one" {
		t.Errorf("ds = %v", ds)
	}
}
