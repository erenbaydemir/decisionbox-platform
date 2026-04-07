## Project Profile

{{PROFILE}}

**IMPORTANT**: Use the project profile above to understand this specific real estate organization — its brand, country, office structure, primary focus, and KPI targets. Tailor all analysis and recommendations to THIS organization's context. When the profile includes targets (e.g., response time target of 30 minutes), use them to calibrate severity — if a metric is below target, that is more severe than a metric that simply declined.

## Real Estate CRM Domain Context

This data comes from **Fizbot**, a real estate CRM and sales management platform. Key domain concepts:

### Lead Types
- **Seller leads** (~96% of volume): Property owners wanting to sell. These come from crawled listing portals (FSBO — For Sale By Owner) and are automatically assigned to agents by Fizbot. The pipeline: prospect → contacted → qualified → meeting → contract.
- **Buyer leads** (~4% of volume): People looking to buy. Created manually by agents or via integrations. The pipeline: prospect → contacted → qualified → meeting → showing → offer.

### Pipeline Stages (stage_id mappings)
- `270010` = Prospect (seller), `270020` = Contacted (seller), `270030` = Qualified (seller), `270040` = Meeting (seller), `270050` = Contract (seller)
- `260010` = Prospect (buyer), `260020` = Contacted (buyer), `260030` = Qualified (buyer), `260040` = Meeting (buyer), `260050` = Showing (buyer), `260060` = Offer (buyer)

### Key ID Mappings
- **category_id**: `30001` = Land, `30002` = Residential, `30003` = Commercial, `30004` = (other type), `30005` = (other type)
- **tenure_id**: `10001` = Sale, `10002` = Rent
- **source_id**: `40001` = Primary portal source, `40002` = Secondary source, `40006` = Third source, `40018` = Fourth source
- **Roles**: Agent (majority), Broker, Team Leader, Assistant, Group Broker, Admin

### Data Patterns
- All tables use **soft deletes** — always filter with `WHERE deleted_at IS NULL`
- User/agent ID field names vary: `user_id`, `assigned_user_id`, `owner_id`, `created_by` — check context
- Offices connect to brands via `brand_id`, brands connect to franchises via `franchise_id`
- Multi-tenant filtering by office: use `assigned_office_id` on leads, `office_id` on transactions

## Previous Discovery Context

{{PREVIOUS_CONTEXT}}
