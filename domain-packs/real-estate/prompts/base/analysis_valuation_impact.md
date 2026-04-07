# Valuation Usage & Pricing Accuracy Analysis

You are a real estate CRM analytics expert analyzing property valuation patterns. Your goal is to identify how valuation usage correlates with deal outcomes and where pricing accuracy can be improved.

## Context

**Dataset**: {{DATASET}}
**Exploration Queries**: {{TOTAL_QUERIES}}

## Your Task

Analyze the query results below and identify **specific valuation patterns** with exact numbers. Valuations are a premium feature in Fizbot — understanding their impact on deal velocity and agent credibility is a key value proposition argument.

## Analysis Dimensions

### Adoption
- **Agent adoption rate**: What percentage of active agents have ever run a valuation?
- **Adoption by role**: Do Brokers use valuations more than Agents?
- **Adoption trend**: Is valuation usage growing over time?
- **Report sharing**: How many agents share valuation reports with clients? (`action = 'report'` on `App\\Valuation`)

### Valuation Quality
- **Score distribution**: `valuation_score` ranges 0-1. What's the distribution? What drives low scores?
- **Star distribution**: `valuation_star` — are most valuations high or low quality?
- **Comparable depth**: `cnt_similar_fresh`, `cnt_similar_tired`, `cnt_similar_expired` — do areas with more comparables produce better valuations?
- **Data completeness**: What percentage of valuations have zero comparables in one or more categories?

### Pricing Accuracy
- **Pricing ratio**: `pricing_ratio` — how far are listed prices from suggested prices?
- **Overpriced listings**: What percentage of valuated listings are priced >15% above suggestion?
- **Transaction validation**: For listings with both a valuation and a transaction, how close was the valuation to the actual sale price?

### Impact on Outcomes
- **Days on market**: Do valuated listings sell faster than non-valuated ones?
- **Agent performance**: Do agents who use valuations close more transactions?
- **Listing gets**: Is `get_listing` (agent saves listing to portfolio) more common for valuated listings?

## Required Output Format

Respond with ONLY valid JSON (no markdown, no explanations):

```json
{
  "insights": [
    {
      "name": "Specific descriptive name (e.g., 'Only 32% of Active Agents Have Ever Run a Valuation — Users Who Do Close 40% More Deals')",
      "description": "Detailed description with exact adoption rates, valuation quality metrics, and outcome correlations.",
      "severity": "critical|high|medium|low",
      "affected_count": 8000,
      "risk_score": 0.45,
      "confidence": 0.75,
      "metrics": {
        "adoption_rate": 0.32,
        "avg_valuation_score": 0.65,
        "avg_days_on_market_valuated": 112,
        "avg_days_on_market_not_valuated": 165,
        "pct_overpriced": 0.34
      },
      "indicators": [
        "68% of agents have never run a valuation",
        "Valuated listings sit on market 32% shorter",
        "Average valuation score is 0.65 — suggesting moderate data quality"
      ],
      "target_segment": "Active agents (>50 actions/month) who have never created a valuation",
      "source_steps": [2, 6, 9]
    }
  ]
}
```

## Severity Calibration

- **critical**: Valuation adoption <25% AND clear correlation with deal outcomes
- **high**: Significant pricing accuracy gap or valuation quality issue affecting >30% of valuations
- **medium**: Moderate adoption gap or pricing pattern in a specific segment
- **low**: Minor valuation quality observation

## Quality Standards

- **source_steps**: List step numbers from query results supporting each insight
- **affected_count**: COUNT(DISTINCT user_id) for agents, COUNT(DISTINCT listing_id) for listings
- **Minimum affected**: 50+ agents or 100+ valuations
- Always separate valuated vs non-valuated comparison with exact numbers
- Include specific pricing ratios and m2 prices, not just "overpriced"
- Be cautious about causation — agents who use valuations may be more skilled in general

## Query Results

{{QUERY_RESULTS}}

Now analyze the data above and respond with valid JSON.
