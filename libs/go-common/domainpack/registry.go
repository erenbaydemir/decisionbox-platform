package domainpack

import (
	"fmt"
	"sync"
)

var (
	packsMu sync.RWMutex
	packs   = make(map[string]Pack)
)

// Register makes a domain pack available by name.
// Domain pack modules call this in their init() function:
//
//	func init() {
//	    domainpack.Register("gaming", NewGamingPack())
//	}
//
// Services then select the pack based on the app's domain field.
func Register(name string, pack Pack) {
	packsMu.Lock()
	defer packsMu.Unlock()
	if pack == nil {
		panic("domainpack: Register pack is nil for " + name)
	}
	if _, exists := packs[name]; exists {
		panic("domainpack: Register called twice for " + name)
	}
	packs[name] = pack
}

// Get returns a registered domain pack by name.
// Returns an error if the pack is not registered.
func Get(name string) (Pack, error) {
	packsMu.RLock()
	pack, exists := packs[name]
	packsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("domainpack: unknown domain %q (registered: %v)", name, RegisteredPacks())
	}
	return pack, nil
}

// RegisteredPacks returns the names of all registered domain packs.
func RegisteredPacks() []string {
	packsMu.RLock()
	defer packsMu.RUnlock()
	names := make([]string, 0, len(packs))
	for k := range packs {
		names = append(names, k)
	}
	return names
}
