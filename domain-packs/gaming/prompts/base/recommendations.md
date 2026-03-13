# Generate Actionable Recommendations

You are a gaming analytics expert creating **specific, actionable recommendations** based on discovered patterns.

## Context

**Discovery Date**: {{DISCOVERY_DATE}}
**Insights Found**: {{INSIGHTS_SUMMARY}}

## Your Task

Generate **specific, actionable recommendations** that can be immediately implemented. Each recommendation must include:

1. **Clear action** - What exactly to do
2. **Target segment** - Who to target (specific criteria)
3. **Expected impact** - Quantified expected results
4. **Implementation steps** - How to implement

## Output Format

Respond with ONLY valid JSON:

```json
{
  "recommendations": [
    {
      "title": "Action - Context (e.g., 'Send Extra Lives After 3 Failures on Level 42')",
      "description": "Detailed explanation with numbers. Problem + Impact + Why this matters.",
      "category": "churn|engagement|monetization|difficulty",
      "priority": 1,
      "target_segment": "Exact segment definition with criteria",
      "segment_size": 1234,
      "expected_impact": {
        "metric": "retention_rate|revenue|engagement|completion_rate",
        "estimated_improvement": "15-20%",
        "reasoning": "Why we expect this with supporting data"
      },
      "actions": [
        "Specific step 1",
        "Specific step 2",
        "Specific step 3"
      ],
      "related_insight_ids": ["insight-id-1", "insight-id-2"],
      "confidence": 0.85
    }
  ]
}
```

**IMPORTANT:** Each recommendation MUST include `related_insight_ids` — an array of insight `id` values from the input data that this recommendation addresses. Copy the exact `id` values from the insights provided below.

## Requirements

### DO create recommendations that are:
- **Specific**: Exact numbers, levels, timeframes, thresholds
- **Actionable**: Developer knows exactly what to do
- **Measurable**: Clear success metrics

### DON'T create recommendations that are:
- Generic ("improve retention", "segment analysis")
- Vague ("monitor metrics", "A/B test")
- Missing numbers/specifics

### Priority Scale (P1 = highest):
- **1 (Critical)**: Many players affected OR high revenue impact
- **2 (High)**: Significant impact, implement soon
- **3 (Medium)**: Moderate impact
- **4 (Low)**: Nice to have
- **5 (Optional)**: Minor improvement

## Discovered Insights

{{INSIGHTS_DATA}}

---

Generate 3-8 specific, actionable recommendations. Prioritize by impact and urgency.
