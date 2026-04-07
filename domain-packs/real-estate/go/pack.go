// Package realestate implements the real estate domain pack for DecisionBox.
// It registers itself as "real-estate" via init() so services can select it
// based on the project's domain field.
//
// This pack provides:
//   - AI Discovery: analysis areas, prompts, and profile schemas for real estate CRM
//   - Categories: sales_navigator (Fizbot Sales Navigator)
//
// Usage:
//
//	import _ "github.com/decisionbox-io/decisionbox/domain-packs/real-estate/go"
//	// Then: domainpack.Get("real-estate")
package realestate

import (
	"github.com/decisionbox-io/decisionbox/libs/go-common/domainpack"
)

func init() {
	domainpack.Register("real-estate", NewPack())
}

// RealEstatePack implements domainpack.Pack and domainpack.DiscoveryPack
// for the real estate CRM domain.
type RealEstatePack struct{}

// NewPack creates a new real estate domain pack.
func NewPack() *RealEstatePack {
	return &RealEstatePack{}
}

func (p *RealEstatePack) Name() string { return "real-estate" }
