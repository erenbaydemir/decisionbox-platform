## Match-3 Game Context

This is a **match-3 puzzle game**. Key aspects to explore:

- **Level progression**: Players progress through sequential levels. Look for difficulty spikes, quit rates per level, and progression bottlenecks.
- **Boosters/power-ups**: Items like Hint, Magnet, Extra Life, Hammer that help players pass levels. Analyze usage patterns, purchase vs earned ratios, and correlation with level completion.
- **Lives system**: Players have limited lives. Running out of lives creates churn risk. Analyze life depletion patterns and recovery times.
- **Session patterns**: Match-3 players tend to have short-to-medium sessions. Look for session duration trends and frequency.
- **Monetization**: Typically freemium with IAP (booster packs, no-ads) and rewarded video ads. Analyze conversion funnels and purchase triggers.
- **Lootboxes/rewards**: Post-level reward chests. Analyze reward satisfaction and its impact on retention.

### Match-3 Example Queries

**Level Difficulty**:
```sql
SELECT level_number, quit_rate, success_rate, avg_attempts_per_player, unique_players
FROM `{{DATASET}}.level_performance_weekly_trends`
{{FILTER}}
  AND week_start_date >= DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
HAVING quit_rate > 0.3 OR success_rate < 0.4
ORDER BY quit_rate DESC
LIMIT 20
```

**Booster Usage Patterns**:
```sql
SELECT booster_name, COUNT(DISTINCT user_id) as unique_users,
       COUNT(*) as total_uses, AVG(level_number) as avg_level
FROM `{{DATASET}}.booster_usage`
{{FILTER}}
GROUP BY booster_name
ORDER BY total_uses DESC
```
