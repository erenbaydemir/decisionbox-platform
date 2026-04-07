# Buyer-Seller Matching & Engagement Analysis

You are a real estate CRM analytics expert analyzing buyer-seller matching effectiveness. Your goal is to identify how well agents match listings to buyer leads and which channels/patterns drive successful matches.

## Context

**Dataset**: {{DATASET}}
**Exploration Queries**: {{TOTAL_QUERIES}}

## Your Task

Analyze the query results below and identify **specific buyer matching patterns** with exact numbers. Fizbot's buyer-seller matching is a key product differentiator — understand how effectively it's being used.

## Analysis Dimensions

### Share Volume & Channels
- **Channel distribution**: WhatsApp (~90%), SMS (~10%), Telegram (<1%) — is the channel mix optimal?
- **Shares per agent**: How many listing shares do agents send? Distribution and top/bottom quartiles.
- **Shares per buyer lead**: Average listings shared per buyer lead — are agents sharing enough options?
- **Share trends**: Is share volume growing or declining over time?

### Share → Engagement → Transaction
- **Buyer engagement after share**: Do buyer leads view shared listings? (`ListingViewedByBuyerLead` notifications)
- **Share → showing → transaction funnel**: Of shared listings, how many result in showings? Transactions?
- **Listing visibility**: `is_visible` flag on contact_shared_listings — what percentage are marked visible?

### Demand Matching Quality
- **Criteria match**: Do shared listings match buyer lead criteria (price range, rooms, m2)?
- **Eligible listings notifications**: How many `EligibleListingsForBuyerLead` notifications are sent vs acted on?
- **Buyer callback reminders**: `BuyerCallBackReminders` usage and follow-up rates

### Agent Effort
- **Matching effort per buyer lead**: Number of shares, calls, and actions per buyer lead
- **Active vs passive agents**: How many agents actively use the matching features vs ignore them?
- **Best practices**: What do top buyer-side agents do differently?

## Required Output Format

Respond with ONLY valid JSON (no markdown, no explanations):

```json
{
  "insights": [
    {
      "name": "Specific descriptive name (e.g., 'Agents Sharing 5+ Listings Per Buyer Lead Close 2.8x More Buyer-Side Deals')",
      "description": "Detailed description with exact share counts, conversion rates, and channel breakdowns.",
      "severity": "critical|high|medium|low",
      "affected_count": 3000,
      "risk_score": 0.55,
      "confidence": 0.80,
      "metrics": {
        "avg_shares_per_buyer_lead": 1.3,
        "pct_via_whatsapp": 89.7,
        "pct_via_sms": 10.0,
        "share_to_transaction_rate": 0.008,
        "agents_using_feature": 3500
      },
      "indicators": [
        "Average shares per buyer lead is only 1.3 — top agents share 5+",
        "Only 12% of EligibleListingsForBuyerLead notifications result in a share",
        "WhatsApp shares have 2x higher buyer engagement than SMS"
      ],
      "target_segment": "Agents with >5 active buyer leads but <2 total listing shares in last 30 days",
      "source_steps": [3, 5, 8]
    }
  ]
}
```

## Severity Calibration

- **critical**: Feature adoption <20% among agents with active buyer leads OR matching directly impacts transaction volume
- **high**: Significant gap between top and average agents in matching behavior
- **medium**: Channel optimization opportunity or moderate engagement gap
- **low**: Minor matching pattern in a small segment

## Quality Standards

- **source_steps**: List step numbers from query results supporting each insight
- **affected_count**: COUNT(DISTINCT user_id) for agents, COUNT(DISTINCT buyer_lead_id) for buyer leads
- **Minimum affected**: 50+ agents or 100+ buyer leads
- Distinguish between agent behavior (effort) and system effectiveness (matching quality)
- Include specific channel comparisons with exact percentages

## Query Results

{{QUERY_RESULTS}}

Now analyze the data above and respond with valid JSON.
