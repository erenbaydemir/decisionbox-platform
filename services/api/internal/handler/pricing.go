package handler

import (
	"context"
	"fmt"
	"net/http"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
	gowarehouse "github.com/decisionbox-io/decisionbox/libs/go-common/warehouse"
	"github.com/decisionbox-io/decisionbox/services/api/internal/database"
	"github.com/decisionbox-io/decisionbox/services/api/internal/models"
)

// SeedPricingFromProviders collects default pricing from all registered providers
// and seeds it to MongoDB if not already present.
func SeedPricingFromProviders(ctx context.Context, repo *database.PricingRepository) {
	existing, _ := repo.Get(ctx)
	if existing != nil {
		return // already seeded
	}

	pricing := &models.Pricing{
		LLM:       make(map[string]map[string]models.TokenPrice),
		Warehouse: make(map[string]models.WarehousePrice),
	}

	// Collect LLM provider pricing
	for _, meta := range gollm.RegisteredProvidersMeta() {
		if len(meta.DefaultPricing) > 0 {
			providerPricing := make(map[string]models.TokenPrice)
			for model, tp := range meta.DefaultPricing {
				providerPricing[model] = models.TokenPrice{
					InputPerMillion:  tp.InputPerMillion,
					OutputPerMillion: tp.OutputPerMillion,
				}
			}
			pricing.LLM[meta.ID] = providerPricing
		}
	}

	// Collect warehouse provider pricing
	for _, meta := range gowarehouse.RegisteredProvidersMeta() {
		if meta.DefaultPricing != nil {
			pricing.Warehouse[meta.ID] = models.WarehousePrice{
				CostModel:           meta.DefaultPricing.CostModel,
				CostPerTBScannedUSD: meta.DefaultPricing.CostPerTBScannedUSD,
			}
		}
	}

	if err := repo.Save(ctx, pricing); err != nil {
		fmt.Printf("Warning: failed to seed pricing: %v\n", err)
	} else {
		fmt.Printf("Pricing seeded from %d LLM + %d warehouse providers\n", len(pricing.LLM), len(pricing.Warehouse))
	}
}

// PricingHandler handles pricing CRUD.
type PricingHandler struct {
	repo *database.PricingRepository
}

func NewPricingHandler(repo *database.PricingRepository) *PricingHandler {
	return &PricingHandler{repo: repo}
}

// Get returns the current pricing data.
// GET /api/v1/pricing
func (h *PricingHandler) Get(w http.ResponseWriter, r *http.Request) {
	pricing, err := h.repo.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get pricing: "+err.Error())
		return
	}
	if pricing == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"llm":       map[string]interface{}{},
			"warehouse": map[string]interface{}{},
		})
		return
	}
	writeJSON(w, http.StatusOK, pricing)
}

// Update saves new pricing data.
// PUT /api/v1/pricing
func (h *PricingHandler) Update(w http.ResponseWriter, r *http.Request) {
	var pricing models.Pricing
	if err := decodeJSON(r, &pricing); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if err := h.repo.Save(r.Context(), &pricing); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save pricing: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, pricing)
}
