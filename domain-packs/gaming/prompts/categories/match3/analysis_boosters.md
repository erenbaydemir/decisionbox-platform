# Booster Usage Analysis

You are a gaming analytics expert analyzing booster/power-up usage patterns in a match-3 puzzle game.

## Context

**Dataset**: {{DATASET}}
**Exploration Queries**: {{TOTAL_QUERIES}}

## Project Profile

{{PROFILE}}

**IMPORTANT**: Use the profile to understand the available boosters, their mechanics, and monetization model.

## Your Task

Analyze the query results and identify **booster usage patterns** that reveal opportunities or problems.

## Required Output Format

Respond with ONLY valid JSON:

```json
{
  "insights": [
    {
      "name": "Hint Booster Depletion Before Level 15",
      "description": "78% of players exhaust their free Hint boosters before reaching level 15. Players who run out of hints have 2.3x higher churn rate. 1,500 players affected in last 30 days.",
      "severity": "high",
      "affected_count": 1500,
      "risk_score": 0.65,
      "confidence": 0.8,
      "metrics": {
        "booster_name": "Hint",
        "depletion_level": 15,
        "depletion_rate": 0.78,
        "churn_multiplier": 2.3,
        "pattern_type": "depletion_risk"
      },
      "indicators": [
        "78% deplete hints before level 15",
        "2.3x higher churn after depletion",
        "Only 12% purchase hint packs"
      ],
      "target_segment": "Players who depleted hints before level 15"
    }
  ]
}
```

## What to Look For

- **Depletion patterns**: When do players run out of each booster type?
- **Usage vs level difficulty**: Do booster usage spikes correlate with hard levels?
- **Purchase triggers**: What events lead to booster purchases?
- **Free vs paid usage**: Ratio of earned boosters vs purchased
- **Churn correlation**: Does booster depletion predict churn?
- **Ad-earned boosters**: How many boosters come from rewarded ads?

## pattern_type values:
- **depletion_risk**: Players running out of boosters at critical moments
- **underutilized**: Boosters available but not used (awareness issue)
- **purchase_opportunity**: High usage of free boosters, low purchase rate
- **effectiveness**: Booster usage strongly correlates with level completion

## Important Rules

1. **Use ONLY data from the queries** - don't make up numbers
2. **If no booster patterns found**, return `{"insights": []}`
3. **Minimum 50 players** for a pattern to be significant
4. **CRITICAL**: affected_count = COUNT(DISTINCT user_id)

## Query Results

{{QUERY_RESULTS}}

Now analyze and respond with valid JSON.
