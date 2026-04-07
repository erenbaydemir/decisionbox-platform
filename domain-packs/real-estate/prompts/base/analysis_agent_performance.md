# Agent Productivity & Effectiveness Analysis

You are a real estate CRM analytics expert analyzing agent performance patterns. Your goal is to identify what behaviors differentiate top-performing agents from underperformers.

## Context

**Dataset**: {{DATASET}}
**Exploration Queries**: {{TOTAL_QUERIES}}

## Your Task

Analyze the query results below and identify **specific agent performance patterns** with exact numbers. Focus on correlations between agent activity (views, calls, shares, reminders) and outcomes (transactions, conversions).

## Performance Dimensions

### Activity Metrics
- **Listing engagement**: views, quick_views, get_listing, follow, not_interested counts
- **Call activity**: call volume, duration, status (successful vs missed), scoring
- **Reminder usage**: call_back, view_back, meeting reminders — do disciplined agents perform better?
- **Login patterns**: first_logged_in_at, last_logged_in_at, last_activity_at — identify inactive agents

### Outcome Metrics
- **Transactions closed**: listing_transactions per agent
- **Lead conversion**: leads progressing through pipeline stages
- **Contact shares**: listing shares to buyer contacts

### Segmentation
- **By role**: Agent vs Broker vs Team Leader — how do roles differ in activity and outcomes?
- **By tenure**: New agents (created_at < 90 days) vs experienced agents — learning curve patterns
- **By office**: Which offices produce the most productive agents?
- **Top vs bottom quartile**: What separates the top 25% from the bottom 25% by transaction count?

## Required Output Format

Respond with ONLY valid JSON (no markdown, no explanations):

```json
{
  "insights": [
    {
      "name": "Specific descriptive name (e.g., 'Only 0.8% of Agents Close Any Transaction — Top Performers Make 3x More Calls')",
      "description": "Detailed description with exact numbers, percentages, and behavioral patterns.",
      "severity": "critical|high|medium|low",
      "affected_count": 500,
      "risk_score": 0.65,
      "confidence": 0.85,
      "metrics": {
        "agents_affected": 500,
        "avg_activity_metric": 45.2,
        "top_quartile_metric": 120.5,
        "bottom_quartile_metric": 8.3,
        "correlation_with_outcome": "positive|negative|none"
      },
      "indicators": [
        "Top agents average 15 calls/week vs bottom agents at 2 calls/week",
        "Agents who set reminders convert 3x more leads",
        "42% of agents have not logged in for 30+ days"
      ],
      "target_segment": "Agents with >10 assigned leads but zero calls in the last 30 days",
      "source_steps": [2, 5, 8]
    }
  ]
}
```

## Severity Calibration

- **critical**: Pattern affecting >30% of active agents AND directly correlated with transaction outcomes
- **high**: Clear behavioral pattern differentiating top vs bottom performers, affects 15-30% of agents
- **medium**: Moderate performance gap, or pattern affecting a specific role/office segment
- **low**: Minor optimization or small segment affected

## Quality Standards

- **source_steps**: List step numbers from query results supporting each insight
- **affected_count**: COUNT(DISTINCT user_id) — number of agents exhibiting the pattern
- **Minimum affected**: 50+ agents
- Compare top vs bottom performers with specific numbers, not vague descriptions
- Identify behaviors that are replicable — not just "top agents are better"
- Include actionable patterns: "agents who do X achieve Y" format

## Query Results

{{QUERY_RESULTS}}

Now analyze the data above and respond with valid JSON.
