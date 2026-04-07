package realestate

import (
	"os"
	"strings"
	"testing"

	"github.com/decisionbox-io/decisionbox/libs/go-common/domainpack"
)

func init() {
	// Tests run from domain-packs/real-estate/go/ — DOMAIN_PACK_PATH is the domain-packs root
	os.Setenv("DOMAIN_PACK_PATH", "../..")
}

func TestRealEstatePackImplementsDiscoveryPack(t *testing.T) {
	pack := NewPack()
	dp, ok := domainpack.AsDiscoveryPack(pack)
	if !ok {
		t.Fatal("RealEstatePack does not implement DiscoveryPack")
	}
	if dp == nil {
		t.Fatal("AsDiscoveryPack returned nil")
	}
}

func TestDomainCategories(t *testing.T) {
	pack := NewPack()
	cats := pack.DomainCategories()

	if len(cats) != 1 {
		t.Fatalf("DomainCategories returned %d categories, want 1", len(cats))
	}

	if cats[0].ID != "sales_navigator" {
		t.Errorf("category ID = %q, want sales_navigator", cats[0].ID)
	}
	if cats[0].Name == "" {
		t.Error("category has empty Name")
	}
	if cats[0].Description == "" {
		t.Error("category has empty Description")
	}
}

func TestAnalysisAreasBase(t *testing.T) {
	pack := NewPack()
	areas := pack.AnalysisAreas("")

	if len(areas) != 6 {
		t.Errorf("base areas = %d, want 6", len(areas))
	}

	ids := make(map[string]bool)
	for _, a := range areas {
		ids[a.ID] = true
		if !a.IsBase {
			t.Errorf("area %q should be IsBase=true", a.ID)
		}
	}

	for _, expected := range []string{"lead_conversion", "agent_performance", "listing_effectiveness", "response_time", "buyer_matching", "valuation_impact"} {
		if !ids[expected] {
			t.Errorf("missing base area: %s", expected)
		}
	}
}

func TestAnalysisAreasSalesNavigator(t *testing.T) {
	pack := NewPack()
	areas := pack.AnalysisAreas("sales_navigator")

	// sales_navigator has no additional category-specific areas (empty areas.json)
	// so it should return only the 6 base areas
	if len(areas) != 6 {
		t.Errorf("sales_navigator areas = %d, want 6 (6 base + 0 category)", len(areas))
	}

	ids := make(map[string]bool)
	for _, a := range areas {
		ids[a.ID] = true
	}

	for _, expected := range []string{"lead_conversion", "agent_performance", "listing_effectiveness", "response_time", "buyer_matching", "valuation_impact"} {
		if !ids[expected] {
			t.Errorf("missing area: %s", expected)
		}
	}
}

func TestAnalysisAreasUnknownCategory(t *testing.T) {
	pack := NewPack()
	areas := pack.AnalysisAreas("nonexistent")

	// Should return only base areas
	if len(areas) != 6 {
		t.Errorf("unknown category areas = %d, want 6 (base only)", len(areas))
	}
}

func TestPromptsBase(t *testing.T) {
	pack := NewPack()
	prompts := pack.Prompts("")

	if prompts.Exploration == "" {
		t.Error("Exploration prompt is empty")
	}
	if !strings.Contains(prompts.Exploration, "Real Estate CRM Analytics Discovery") {
		t.Error("Exploration prompt missing expected header")
	}
	if prompts.Recommendations == "" {
		t.Error("Recommendations prompt is empty")
	}

	// Should have all 6 base analysis areas
	for _, id := range []string{"lead_conversion", "agent_performance", "listing_effectiveness", "response_time", "buyer_matching", "valuation_impact"} {
		if _, ok := prompts.AnalysisAreas[id]; !ok {
			t.Errorf("missing base analysis prompt: %s", id)
		}
	}

	// BaseContext should be loaded
	if prompts.BaseContext == "" {
		t.Error("BaseContext is empty — base_context.md not loaded")
	}
	if !strings.Contains(prompts.BaseContext, "{{PROFILE}}") {
		t.Error("BaseContext missing {{PROFILE}} placeholder")
	}
	if !strings.Contains(prompts.BaseContext, "{{PREVIOUS_CONTEXT}}") {
		t.Error("BaseContext missing {{PREVIOUS_CONTEXT}} placeholder")
	}
}

func TestPromptsBase_NoProfileInExploration(t *testing.T) {
	pack := NewPack()
	prompts := pack.Prompts("")

	// {{PROFILE}} and {{PREVIOUS_CONTEXT}} should NOT be in exploration prompt
	// (they come from base_context which is prepended by the orchestrator)
	if strings.Contains(prompts.Exploration, "{{PROFILE}}") {
		t.Error("exploration prompt should not contain {{PROFILE}} — moved to base_context")
	}
	if strings.Contains(prompts.Exploration, "{{PREVIOUS_CONTEXT}}") {
		t.Error("exploration prompt should not contain {{PREVIOUS_CONTEXT}} — moved to base_context")
	}
}

func TestPromptsBase_NoProfileInAnalysis(t *testing.T) {
	pack := NewPack()
	prompts := pack.Prompts("sales_navigator")

	for id, content := range prompts.AnalysisAreas {
		if strings.Contains(content, "{{PROFILE}}") {
			t.Errorf("analysis prompt %q should not contain {{PROFILE}} — moved to base_context", id)
		}
		if strings.Contains(content, "{{PREVIOUS_CONTEXT}}") {
			t.Errorf("analysis prompt %q should not contain {{PREVIOUS_CONTEXT}} — moved to base_context", id)
		}
	}
}

func TestPromptsSalesNavigatorMerge(t *testing.T) {
	pack := NewPack()
	prompts := pack.Prompts("sales_navigator")

	// Exploration should contain both base and sales_navigator context
	if !strings.Contains(prompts.Exploration, "Real Estate CRM Analytics Discovery") {
		t.Error("merged exploration missing base content")
	}
	if !strings.Contains(prompts.Exploration, "Fizbot Sales Navigator Context") {
		t.Error("merged exploration missing sales_navigator context")
	}

	// Should have all 6 analysis area prompts
	for _, id := range []string{"lead_conversion", "agent_performance", "listing_effectiveness", "response_time", "buyer_matching", "valuation_impact"} {
		content, ok := prompts.AnalysisAreas[id]
		if !ok {
			t.Errorf("missing analysis prompt: %s", id)
			continue
		}
		if content == "" {
			t.Errorf("empty analysis prompt: %s", id)
		}
	}
}

func TestProfileSchemaBase(t *testing.T) {
	pack := NewPack()
	schema := pack.ProfileSchema("")

	if _, ok := schema["error"]; ok {
		t.Fatalf("ProfileSchema returned error: %v", schema["error"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("schema has no properties")
	}

	for _, expected := range []string{"business_info", "business_model", "kpis"} {
		if _, ok := props[expected]; !ok {
			t.Errorf("base schema missing property: %s", expected)
		}
	}
}

func TestProfileSchemaSalesNavigatorMerge(t *testing.T) {
	pack := NewPack()
	schema := pack.ProfileSchema("sales_navigator")

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("schema has no properties")
	}

	// Should have base properties
	for _, expected := range []string{"business_info", "business_model", "kpis"} {
		if _, ok := props[expected]; !ok {
			t.Errorf("merged schema missing base property: %s", expected)
		}
	}

	// Should have sales_navigator-specific properties
	for _, expected := range []string{"platform_usage"} {
		if _, ok := props[expected]; !ok {
			t.Errorf("merged schema missing sales_navigator property: %s", expected)
		}
	}
}

func TestAnalysisAreaKeywords(t *testing.T) {
	pack := NewPack()
	areas := pack.AnalysisAreas("sales_navigator")

	for _, a := range areas {
		if len(a.Keywords) == 0 {
			t.Errorf("area %q has no keywords", a.ID)
		}
		if a.Name == "" {
			t.Errorf("area %q has empty Name", a.ID)
		}
		if a.Priority == 0 {
			t.Errorf("area %q has zero Priority", a.ID)
		}
	}
}

func TestAnalysisPromptsContainQueryResults(t *testing.T) {
	pack := NewPack()
	prompts := pack.Prompts("sales_navigator")

	for id, content := range prompts.AnalysisAreas {
		if !strings.Contains(content, "{{QUERY_RESULTS}}") {
			t.Errorf("analysis prompt %q missing {{QUERY_RESULTS}} variable", id)
		}
	}
}

func TestExplorationContainsRequiredVars(t *testing.T) {
	pack := NewPack()
	prompts := pack.Prompts("")

	requiredVars := []string{"{{DATASET}}", "{{SCHEMA_INFO}}", "{{ANALYSIS_AREAS}}"}
	for _, v := range requiredVars {
		if !strings.Contains(prompts.Exploration, v) {
			t.Errorf("exploration prompt missing %s variable", v)
		}
	}
}

func TestRecommendationsContainsRequiredVars(t *testing.T) {
	pack := NewPack()
	prompts := pack.Prompts("")

	requiredVars := []string{"{{INSIGHTS_DATA}}", "{{INSIGHTS_SUMMARY}}", "{{DISCOVERY_DATE}}"}
	for _, v := range requiredVars {
		if !strings.Contains(prompts.Recommendations, v) {
			t.Errorf("recommendations prompt missing %s variable", v)
		}
	}
}
