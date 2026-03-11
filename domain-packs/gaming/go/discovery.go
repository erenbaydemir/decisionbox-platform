package gaming

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decisionbox-io/decisionbox/libs/go-common/domainpack"
)

// Default paths relative to the working directory.
// Can be overridden via DOMAIN_PACK_PATH env var.
var (
	promptsPath  = "domain-packs/gaming/prompts"
	profilesPath = "domain-packs/gaming/profiles"
)

func init() {
	if p := os.Getenv("DOMAIN_PACK_PATH"); p != "" {
		promptsPath = filepath.Join(p, "prompts")
		profilesPath = filepath.Join(p, "profiles")
	}
}

// Compile-time check: GamingPack implements DiscoveryPack.
var _ domainpack.DiscoveryPack = (*GamingPack)(nil)

// DomainCategories returns the game genre categories.
func (p *GamingPack) DomainCategories() []domainpack.DomainCategory {
	return []domainpack.DomainCategory{
		{
			ID:          "match3",
			Name:        "Match-3",
			Description: "Puzzle games with match-3 mechanics (e.g., Candy Crush, Manor Cafe)",
		},
	}
}

// baseAnalysisAreas are shared across all gaming categories.
var baseAnalysisAreas = []domainpack.AnalysisArea{
	{
		ID:          "churn",
		Name:        "Churn Risks",
		Description: "Players at risk of leaving the game",
		Keywords:    []string{"churn", "retention", "cohort", "day_", "d1_", "d7_", "d30_", "inactive", "lapsed"},
		IsBase:      true,
		Priority:    1,
	},
	{
		ID:          "engagement",
		Name:        "Engagement Patterns",
		Description: "Player behavior and session trends",
		Keywords:    []string{"session", "engagement", "duration", "frequency", "active", "dau", "mau", "playtime"},
		IsBase:      true,
		Priority:    2,
	},
	{
		ID:          "monetization",
		Name:        "Monetization Opportunities",
		Description: "Revenue optimization and conversion opportunities",
		Keywords:    []string{"purchase", "iap", "revenue", "payer", "currency", "spend", "arpu", "ltv", "conversion"},
		IsBase:      true,
		Priority:    3,
	},
}

// categoryAnalysisAreas maps category ID to category-specific analysis areas.
var categoryAnalysisAreas = map[string][]domainpack.AnalysisArea{
	"match3": {
		{
			ID:          "levels",
			Name:        "Level Difficulty",
			Description: "Difficulty spikes and frustration points in level progression",
			Keywords:    []string{"level", "quit", "success", "difficulty", "fail", "attempt", "stage", "star"},
			IsBase:      false,
			Priority:    4,
		},
		{
			ID:          "boosters",
			Name:        "Booster Usage",
			Description: "Power-up usage patterns, depletion risks, and purchase opportunities",
			Keywords:    []string{"booster", "hint", "magnet", "power", "extra_life", "hammer", "consumable"},
			IsBase:      false,
			Priority:    5,
		},
	},
}

// AnalysisAreas returns base + category-specific analysis areas.
func (p *GamingPack) AnalysisAreas(categoryID string) []domainpack.AnalysisArea {
	areas := make([]domainpack.AnalysisArea, len(baseAnalysisAreas))
	copy(areas, baseAnalysisAreas)

	if specific, ok := categoryAnalysisAreas[categoryID]; ok {
		areas = append(areas, specific...)
	}

	return areas
}

// Prompts returns merged prompt templates for a given category.
// Reads from filesystem (not embedded) so prompts are language-agnostic.
func (p *GamingPack) Prompts(categoryID string) domainpack.PromptTemplates {
	templates := domainpack.PromptTemplates{
		AnalysisAreas: make(map[string]string),
	}

	// Load base exploration prompt
	templates.Exploration = readPromptFile(filepath.Join(promptsPath, "base", "exploration.md"))

	// Merge category-specific exploration context
	if categoryID != "" {
		contextPath := filepath.Join(promptsPath, "categories", categoryID, "exploration_context.md")
		if context := readPromptFile(contextPath); context != "" {
			templates.Exploration = templates.Exploration + "\n\n" + context
		}
	}

	// Load base recommendations prompt
	templates.Recommendations = readPromptFile(filepath.Join(promptsPath, "base", "recommendations.md"))

	// Load base analysis prompts
	for _, area := range baseAnalysisAreas {
		path := filepath.Join(promptsPath, "base", fmt.Sprintf("analysis_%s.md", area.ID))
		if content := readPromptFile(path); content != "" {
			templates.AnalysisAreas[area.ID] = content
		}
	}

	// Load category-specific analysis prompts
	if categoryID != "" {
		if specific, ok := categoryAnalysisAreas[categoryID]; ok {
			for _, area := range specific {
				path := filepath.Join(promptsPath, "categories", categoryID, fmt.Sprintf("analysis_%s.md", area.ID))
				if content := readPromptFile(path); content != "" {
					templates.AnalysisAreas[area.ID] = content
				}
			}
		}
	}

	return templates
}

// ProfileSchema returns the merged JSON Schema for a given category.
func (p *GamingPack) ProfileSchema(categoryID string) map[string]interface{} {
	// Parse base schema
	baseData, err := os.ReadFile(filepath.Join(profilesPath, "schema.json"))
	if err != nil {
		return map[string]interface{}{"error": "base schema not found: " + err.Error()}
	}

	var base map[string]interface{}
	if err := json.Unmarshal(baseData, &base); err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	if categoryID == "" {
		return base
	}

	// Load category-specific schema and merge properties
	catPath := filepath.Join(profilesPath, "categories", categoryID+".json")
	catData, err := os.ReadFile(catPath)
	if err != nil {
		return base
	}

	var catSchema map[string]interface{}
	if err := json.Unmarshal(catData, &catSchema); err != nil {
		return base
	}

	baseProps, _ := base["properties"].(map[string]interface{})
	catProps, _ := catSchema["properties"].(map[string]interface{})
	if baseProps != nil && catProps != nil {
		for k, v := range catProps {
			baseProps[k] = v
		}
	}

	return base
}

func readPromptFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
