// Package gaming implements the gaming domain pack for DecisionBox.
// It registers itself as "gaming" via init() so services can select it
// based on the project's domain field.
//
// This pack provides:
//   - AI Discovery: analysis areas, prompts, and profile schemas for gaming
//   - Categories: match3 (MVP), fps/strategy/puzzle planned
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

// GamingPack implements domainpack.Pack and domainpack.DiscoveryPack
// for the gaming domain.
type GamingPack struct{}

// NewPack creates a new gaming domain pack.
func NewPack() *GamingPack {
	return &GamingPack{}
}

func (p *GamingPack) Name() string { return "gaming" }
