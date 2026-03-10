// Package domainpack defines interfaces for domain-specific functionality.
//
// DecisionBox's core platform is domain-agnostic. Domain packs layer
// industry-specific logic on top: discovery prompts, analysis categories,
// profile schemas, and (future) semantic text generation.
//
// Each domain pack (gaming, e-commerce, social, etc.) implements these
// interfaces and registers itself via Register(). Services select the
// pack based on the project's "domain" field.
//
// Example — adding an e-commerce domain pack:
//
//  1. Create domain-packs/ecommerce/go/ with implementations
//  2. Call domainpack.Register("ecommerce", NewEcommercePack())
//  3. Service: import _ ".../domain-packs/ecommerce/go"
//  4. Set domain: "ecommerce" on the project
package domainpack

// Pack represents a domain pack. Every domain pack must implement this
// base interface. Additional capabilities (like DiscoveryPack) are
// checked via type assertion at runtime.
type Pack interface {
	// Name returns the domain name (e.g., "gaming", "ecommerce").
	Name() string
}
