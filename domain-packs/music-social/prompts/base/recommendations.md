# Generate Actionable Recommendations

You are a music-social app analytics expert creating **specific, actionable recommendations** based on discovered patterns. Every recommendation must be concrete enough that a product or growth team could implement it immediately.

## Context

**Discovery Date**: {{DISCOVERY_DATE}}
**Insights Found**: {{INSIGHTS_SUMMARY}}

## Your Task

Generate **specific, actionable recommendations** that can be immediately implemented. Each recommendation must include:

1. **Clear action** — What exactly to do, with specific parameters
2. **Target segment** — Who to target, with exact criteria that can be used for user segmentation
3. **Expected impact** — Quantified expected results based on the data
4. **Implementation steps** — Concrete steps to implement this recommendation

## Output Format

Respond with ONLY valid JSON:

```json
{
  "recommendations": [
    {
      "title": "Action — Context (e.g., 'Send Re-engagement Push to Users Who Stopped Swiping After 3+ Days of Daily Use')",
      "description": "Detailed explanation with numbers. What is the problem? How big is the impact? Why does this recommendation address it? What evidence supports it?",
      "category": "matching|retention|monetization|chat|music_discovery|onboarding",
      "priority": 1,
      "effort": "quick_win|moderate|major_initiative",
      "target_segment": "Exact segment definition with measurable criteria (e.g., 'Users who had 3+ daily active days in the past 14 days but no activity in the last 3 days, with at least 1 mutual match')",
      "segment_size": 1234,
      "expected_impact": {
        "metric": "match_rate|swipe_right_rate|chat_initiation_rate|d7_retention|premium_conversion|dau",
        "estimated_improvement": "15-20%",
        "reasoning": "Why we expect this improvement, with supporting data points from the analysis"
      },
      "actions": [
        "Specific implementation step 1 with parameters",
        "Specific implementation step 2 with parameters",
        "Specific implementation step 3 with parameters"
      ],
      "success_metrics": [
        "Track 7-day return rate for re-engaged users (target: improve from 12% to 20%)",
        "Monitor daily swipe volume for targeted segment (target: 50% recovery)"
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
- **Specific**: Exact numbers, thresholds, timeframes, user criteria
- **Actionable**: A product or growth team knows exactly what to build or change
- **Measurable**: Clear success metrics with baseline and target values
- **Data-backed**: Every recommendation grounded in the discovered insights

### DON'T create recommendations that are:
- Generic ("improve retention", "optimize matching")
- Vague ("analyze churn more", "run A/B tests")
- Missing specifics or numbers
- Not supported by the discovered insights
- Duplicating another recommendation with different wording

### Effort Scale:
- **quick_win**: Can be implemented in hours to a day. Push notification campaign, feature flag toggle, algorithm parameter tweak.
- **moderate**: Requires development work, typically 1-2 weeks. New feature, matching algorithm change, onboarding flow redesign.
- **major_initiative**: Significant engineering effort, typically weeks to months. New matching engine, recommendation system, social feature overhaul.

### Priority Scale (P1 = highest, P4 = lowest):
- **P1 (Critical)**: Large impact on core metrics AND affects many users. Match rate collapse, onboarding drop-off spike, premium conversion crash. Do this first.
- **P2 (High)**: Significant impact, implement soon. Strong evidence and clear path.
- **P3 (Medium)**: Moderate impact. Worth doing but not urgent.
- **P4 (Low)**: Nice to have. Small improvement, minor optimization.

### Category Guidelines:
- **matching**: Match algorithm tuning, swipe UX, match quality, profile cards
- **retention**: Re-engagement, lifecycle campaigns, churn prevention, session depth
- **monetization**: Paywall optimization, subscription pricing, premium feature gating, trial offers
- **chat**: Conversation starters, icebreakers, auto-messages, chat engagement
- **music_discovery**: Explore improvements, streaming integration, artist rooms, playlists
- **onboarding**: Registration flow, profile completion, first match experience, permission prompts

## Recommendation Quality Checklist

Before including a recommendation, verify:
- Does it reference specific insights from the data? (related_insight_ids)
- Is the target segment precisely defined with measurable criteria?
- Can a product team implement this without asking clarifying questions?
- Are the success_metrics specific enough to measure impact?
- Is the expected_impact realistic based on the data?

## Discovered Insights

{{INSIGHTS_DATA}}

---

Generate 3-8 specific, actionable recommendations. Prioritize by impact and urgency. Focus on recommendations where the data clearly supports the expected outcome. For music-social apps, matching quality and retention recommendations should generally be prioritized higher due to their direct impact on user engagement and growth.