package handler

import (
	"net/http"

	"github.com/decisionbox-io/decisionbox/libs/go-common/domainpack"
)

// DomainsHandler handles domain pack endpoints.
type DomainsHandler struct{}

func NewDomainsHandler() *DomainsHandler {
	return &DomainsHandler{}
}

// ListDomains returns all registered domain packs with their categories.
// GET /api/v1/domains
func (h *DomainsHandler) ListDomains(w http.ResponseWriter, r *http.Request) {
	type categoryInfo struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type domainInfo struct {
		ID         string         `json:"id"`
		Categories []categoryInfo `json:"categories"`
	}

	var domains []domainInfo
	for _, name := range domainpack.RegisteredPacks() {
		pack, err := domainpack.Get(name)
		if err != nil {
			continue
		}

		info := domainInfo{ID: pack.Name()}

		if dp, ok := domainpack.AsDiscoveryPack(pack); ok {
			for _, cat := range dp.DomainCategories() {
				info.Categories = append(info.Categories, categoryInfo{
					ID:          cat.ID,
					Name:        cat.Name,
					Description: cat.Description,
				})
			}
		}

		domains = append(domains, info)
	}

	writeJSON(w, http.StatusOK, domains)
}

// ListCategories returns categories for a domain.
// GET /api/v1/domains/{domain}/categories
func (h *DomainsHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")

	pack, err := domainpack.Get(domain)
	if err != nil {
		writeError(w, http.StatusNotFound, "domain not found: "+domain)
		return
	}

	dp, ok := domainpack.AsDiscoveryPack(pack)
	if !ok {
		writeError(w, http.StatusBadRequest, "domain does not support discovery")
		return
	}

	writeJSON(w, http.StatusOK, dp.DomainCategories())
}

// GetProfileSchema returns the JSON Schema for a domain/category.
// GET /api/v1/domains/{domain}/categories/{category}/schema
func (h *DomainsHandler) GetProfileSchema(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	category := r.PathValue("category")

	pack, err := domainpack.Get(domain)
	if err != nil {
		writeError(w, http.StatusNotFound, "domain not found: "+domain)
		return
	}

	dp, ok := domainpack.AsDiscoveryPack(pack)
	if !ok {
		writeError(w, http.StatusBadRequest, "domain does not support discovery")
		return
	}

	writeJSON(w, http.StatusOK, dp.ProfileSchema(category))
}

// GetAnalysisAreas returns analysis areas for a domain/category.
// GET /api/v1/domains/{domain}/categories/{category}/areas
func (h *DomainsHandler) GetAnalysisAreas(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	category := r.PathValue("category")

	pack, err := domainpack.Get(domain)
	if err != nil {
		writeError(w, http.StatusNotFound, "domain not found: "+domain)
		return
	}

	dp, ok := domainpack.AsDiscoveryPack(pack)
	if !ok {
		writeError(w, http.StatusBadRequest, "domain does not support discovery")
		return
	}

	writeJSON(w, http.StatusOK, dp.AnalysisAreas(category))
}
