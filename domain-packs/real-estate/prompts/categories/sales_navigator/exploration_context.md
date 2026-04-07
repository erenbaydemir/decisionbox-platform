## Fizbot Sales Navigator Context

This is data from **Fizbot Sales Navigator**, a CRM platform used by real estate agencies and brokerages. Key aspects to explore:

- **Seller-dominant pipeline**: ~96% of leads are seller leads (FSBO — For Sale By Owner), crawled from real estate portals and auto-assigned to agents. Only ~4% are buyer leads, manually created. Analyze these pipelines separately — they have completely different dynamics.
- **Lead sources**: Leads come from multiple sources (source_id). The largest source (40001) accounts for ~70% of leads. Analyze conversion rates by source to identify which portals deliver the highest-quality leads.
- **Office hierarchy**: Offices belong to brands (brand_id), brands belong to franchises (franchise_id). When comparing performance, segment by brand/franchise to identify organizational best practices. Key brands: RE/MAX, Coldwell Banker, Turyap, plus thousands of independent offices.
- **Multi-country**: Fizbot operates in Turkey (primary), Portugal, Italy, Spain, and Romania. Look for country-level patterns in the brand/office data.
- **Agent roles**: 93% are Agents, 4% Brokers, 2% Team Leaders, 1% Assistants. Brokers and Team Leaders manage offices — their actions affect many agents.
- **Valuation tool**: Fizbot's valuation feature compares listings against similar properties (fresh, tired, expired, transaction comps). Valuation adoption is a premium feature — track its usage and impact on deal outcomes.
- **Communication channels**: Agents share listings with buyer contacts primarily via WhatsApp (~90%), with SMS (~10%) as secondary. Telegram is negligible. Channel effectiveness may vary by market.
- **Subscription model**: Offices subscribe to Fizbot with seat-based pricing. 681 currently active subscriptions out of 11.7K total — subscription churn is a business concern for Fizbot itself.

### Fizbot-Specific Example Queries

**Lead Conversion by Source**:
```sql
SELECT
  source_id,
  COUNT(*) as total_leads,
  COUNTIF(contacted_at IS NOT NULL) as contacted,
  ROUND(SAFE_DIVIDE(COUNTIF(contacted_at IS NOT NULL), COUNT(*)) * 100, 2) as contact_rate_pct,
  COUNTIF(contract_at IS NOT NULL) as contracts
FROM `{{DATASET}}.leads`
WHERE deleted_at IS NULL
  AND lead_type = 'seller'
  AND created_at >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)
{{FILTER}}
GROUP BY source_id
ORDER BY total_leads DESC
LIMIT 10
```

**Office Performance Comparison**:
```sql
SELECT
  o.brand_id,
  b.name as brand_name,
  COUNT(DISTINCT l.assigned_office_id) as offices,
  COUNT(DISTINCT l.assigned_user_id) as agents,
  COUNT(*) as total_leads,
  COUNTIF(l.contacted_at IS NOT NULL) as contacted,
  ROUND(SAFE_DIVIDE(COUNTIF(l.contacted_at IS NOT NULL), COUNT(*)) * 100, 2) as contact_rate_pct
FROM `{{DATASET}}.leads` l
JOIN `{{DATASET}}.offices` o ON l.assigned_office_id = o.id
JOIN `{{DATASET}}.brands` b ON o.brand_id = b.id
WHERE l.deleted_at IS NULL
  AND l.created_at >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)
{{FILTER}}
GROUP BY o.brand_id, b.name
HAVING COUNT(*) > 100
ORDER BY contact_rate_pct DESC
LIMIT 20
```

**Valuation Adoption by Agent Tenure**:
```sql
SELECT
  CASE
    WHEN DATE_DIFF(CURRENT_DATE(), DATE(u.created_at), DAY) < 90 THEN 'new_0_90d'
    WHEN DATE_DIFF(CURRENT_DATE(), DATE(u.created_at), DAY) < 365 THEN 'mid_90_365d'
    ELSE 'veteran_365d+'
  END as agent_tenure,
  COUNT(DISTINCT u.id) as total_agents,
  COUNT(DISTINCT v.user_id) as agents_with_valuations,
  ROUND(SAFE_DIVIDE(COUNT(DISTINCT v.user_id), COUNT(DISTINCT u.id)) * 100, 2) as adoption_pct
FROM `{{DATASET}}.users` u
LEFT JOIN `{{DATASET}}.valuations` v ON v.user_id = u.id AND v.deleted_at IS NULL
WHERE u.deleted_at IS NULL
{{FILTER}}
GROUP BY agent_tenure
ORDER BY adoption_pct DESC
```

**Contact Share Effectiveness**:
```sql
SELECT
  cs.channel,
  COUNT(DISTINCT cs.id) as total_shares,
  COUNT(DISTINCT cs.user_id) as agents,
  COUNT(DISTINCT csl.listing_id) as unique_listings_shared,
  AVG(listings_per_share.cnt) as avg_listings_per_share
FROM `{{DATASET}}.contact_shares` cs
LEFT JOIN (
  SELECT contact_share_id, COUNT(*) as cnt
  FROM `{{DATASET}}.contact_shared_listings`
  WHERE deleted_at IS NULL
  GROUP BY contact_share_id
) listings_per_share ON listings_per_share.contact_share_id = cs.id
LEFT JOIN `{{DATASET}}.contact_shared_listings` csl ON csl.contact_share_id = cs.id AND csl.deleted_at IS NULL
WHERE cs.deleted_at IS NULL
  AND cs.created_at >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)
{{FILTER}}
GROUP BY cs.channel
ORDER BY total_shares DESC
```
