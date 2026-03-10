// Package gaming implements the gaming domain pack for DecisionBox.
// It registers itself as "gaming" via init() so services can select it
// based on the app's domain field.
//
// This pack provides:
//   - Semantic text generation for gaming entities (players, levels, sessions, cohorts)
//   - Warehouse queries for gaming-specific feature tables
//   - Entity type definitions for the gaming domain
//
// Usage:
//
//	import _ "github.com/decisionbox-io/decisionbox/domain-packs/gaming/go"
//	// Then: domainpack.Get("gaming")
package gaming

import (
	"github.com/decisionbox-io/decisionbox/libs/go-common/domainpack"
)

func init() {
	domainpack.Register("gaming", NewPack())
}

// GamingPack implements domainpack.Pack for the gaming domain.
type GamingPack struct {
	semantic *GamingSemanticGenerator
	fetcher  *GamingDataFetcher
}

// NewPack creates a new gaming domain pack.
// Reads DOMAIN_PACK_QUERIES_PATH and WAREHOUSE_PROVIDER env vars
// to configure the data fetcher's query loading.
func NewPack() *GamingPack {
	return &GamingPack{
		semantic: &GamingSemanticGenerator{},
		fetcher:  NewGamingDataFetcher("", ""),
	}
}

func (p *GamingPack) Name() string                              { return "gaming" }
func (p *GamingPack) SemanticGenerator() domainpack.SemanticGenerator { return p.semantic }
func (p *GamingPack) DataFetcher() domainpack.DataFetcher            { return p.fetcher }
func (p *GamingPack) EntityTypes() []string {
	return []string{"users", "entities", "sessions", "cohorts"}
}
