package models

// CostEstimate is the result of a cost estimation.
type CostEstimate struct {
	LLM       LLMCostEstimate       `json:"llm"`
	Warehouse WarehouseCostEstimate `json:"warehouse"`
	TotalUSD  float64               `json:"total_cost_usd"`
	Breakdown CostBreakdown         `json:"breakdown"`
}

type LLMCostEstimate struct {
	Provider              string  `json:"provider"`
	Model                 string  `json:"model"`
	EstimatedInputTokens  int     `json:"estimated_input_tokens"`
	EstimatedOutputTokens int     `json:"estimated_output_tokens"`
	CostUSD               float64 `json:"cost_usd"`
}

type WarehouseCostEstimate struct {
	Provider              string  `json:"provider"`
	EstimatedQueries      int     `json:"estimated_queries"`
	EstimatedBytesScanned int64   `json:"estimated_bytes_scanned"`
	CostUSD               float64 `json:"cost_usd"`
}

type CostBreakdown struct {
	Exploration     float64 `json:"exploration"`
	Analysis        float64 `json:"analysis"`
	Validation      float64 `json:"validation"`
	Recommendations float64 `json:"recommendations"`
}
