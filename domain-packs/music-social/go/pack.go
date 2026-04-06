// Package musicsocial implements the music-social domain pack for DecisionBox.
// It registers itself as "music-social" via init() so services can select it
// based on the project's domain field.
//
// This pack provides:
//   - AI Discovery: analysis areas, prompts, and profile schemas for music-based social apps
//   - Categories: music_matching (music-taste-based social matching and dating)
//
// Usage:
//
//	import _ "github.com/decisionbox-io/decisionbox/domain-packs/music-social/go"
//	// Then: domainpack.Get("music-social")
package musicsocial

import (
	"github.com/decisionbox-io/decisionbox/libs/go-common/domainpack"
)

func init() {
	domainpack.Register("music-social", NewPack())
}

// MusicSocialPack implements domainpack.Pack and domainpack.DiscoveryPack
// for the music-social domain.
type MusicSocialPack struct{}

// NewPack creates a new music-social domain pack.
func NewPack() *MusicSocialPack {
	return &MusicSocialPack{}
}

func (p *MusicSocialPack) Name() string { return "music-social" }
