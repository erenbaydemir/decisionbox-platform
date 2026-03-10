package domainpack

// DiscoveryPack provides domain-specific logic for AI Discovery.
//
// Domain packs optionally implement this interface alongside the base Pack
// interface. The discovery agent checks if a registered pack implements
// DiscoveryPack at runtime using a type assertion:
//
//	pack, _ := domainpack.Get("gaming")
//	dp, ok := pack.(domainpack.DiscoveryPack)
//
// The three-level hierarchy is: Domain > Category > Analysis Areas.
//
//	Domain: gaming
//	├── Category: match3
//	│   ├── Analysis: churn (base)
//	│   ├── Analysis: engagement (base)
//	│   ├── Analysis: monetization (base)
//	│   └── Analysis: levels (match3-specific)
//	└── Category: fps
//	    ├── Analysis: churn (base)
//	    ├── Analysis: engagement (base)
//	    ├── Analysis: monetization (base)
//	    └── Analysis: weapon_balance (fps-specific)
type DiscoveryPack interface {
	// DomainCategories returns the categories within this domain.
	// Each category represents a sub-type of the domain (e.g., game genre).
	//
	//   Gaming:    [match3, fps, strategy, puzzle, ...]
	//   Ecommerce: [b2c, marketplace, subscription, ...]
	DomainCategories() []DomainCategory

	// AnalysisAreas returns the analysis areas for a given category.
	// The result merges base areas (common to the domain) with
	// category-specific areas. Base areas have IsBase=true.
	//
	// If categoryID is empty, returns only base areas.
	AnalysisAreas(categoryID string) []AnalysisArea

	// Prompts returns the merged prompt templates for a given category.
	// Base prompts are combined with category-specific overrides.
	//
	// The Exploration prompt is formed by appending the category's
	// exploration context to the base exploration prompt.
	//
	// Category-specific analysis prompts are added to the base set.
	Prompts(categoryID string) PromptTemplates

	// ProfileSchema returns a JSON Schema (as map) for the project profile
	// for a given category. The UI renders a dynamic form from this schema.
	// Data is stored as map[string]interface{} in MongoDB.
	//
	//   Gaming/match3: genre, mechanics, boosters, IAP packages, KPIs, ...
	//   Ecommerce/b2c: product types, pricing, shipping, AOV targets, ...
	//
	// If categoryID is empty, returns the base domain schema.
	ProfileSchema(categoryID string) map[string]interface{}

}

// DomainCategory is a sub-type within a domain.
//
// Examples:
//
//	Gaming: {ID: "match3", Name: "Match-3", Description: "Puzzle games with match-3 mechanics"}
//	Ecommerce: {ID: "b2c", Name: "B2C Retail", Description: "Direct-to-consumer retail"}
type DomainCategory struct {
	ID          string // unique key within domain: "match3", "fps", "b2c"
	Name        string // display name: "Match-3", "FPS / Shooter", "B2C Retail"
	Description string // short description for UI/docs
}

// AnalysisArea defines a type of insight that the discovery agent looks for.
//
// Base areas are shared across all categories in a domain (e.g., churn is
// relevant for all game genres). Category-specific areas are only relevant
// for that category (e.g., level_difficulty for puzzle games).
type AnalysisArea struct {
	ID          string   // unique key: "churn", "levels", "weapon_balance"
	Name        string   // display name: "Churn Risks", "Level Difficulty"
	Description string   // what this area looks for
	Keywords    []string // keywords to filter exploration queries into this area
	IsBase      bool     // true = shared across all categories in the domain
	Priority    int      // display order (1 = highest priority)
}

// PromptTemplates holds all prompt templates for a domain + category combination.
//
// The Exploration field is the merged result of base + category context.
// The AnalysisAreas map contains prompts for both base and category-specific areas.
type PromptTemplates struct {
	// Exploration is the main system prompt for the autonomous exploration phase.
	// Already merged: base exploration + category-specific context.
	Exploration string

	// Recommendations is the prompt for generating actionable recommendations
	// from discovered insights.
	Recommendations string

	// AnalysisAreas maps analysis area ID to its analysis prompt.
	// Includes both base areas and category-specific areas.
	//   {"churn": "...", "engagement": "...", "levels": "..."}
	AnalysisAreas map[string]string
}

// AsDiscoveryPack attempts to extract DiscoveryPack from a Pack.
// Returns nil, false if the pack does not implement DiscoveryPack.
func AsDiscoveryPack(pack Pack) (DiscoveryPack, bool) {
	dp, ok := pack.(DiscoveryPack)
	return dp, ok
}
