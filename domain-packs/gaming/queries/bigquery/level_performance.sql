-- Gaming: Level performance query
-- Fetches level difficulty metrics from level_performance_summary.
--
-- Parameters:
--   {{dataset}}       — BigQuery dataset name
--   {{app_id}}        — Application ID filter
--   {{lookback_days}} — How far back to look (0 = no limit)

SELECT
    app_id, level_number, version,
    unique_players, total_starts, total_successes, total_failures,
    success_rate, avg_attempts_per_player, quit_rate,
    booster_acceptance_rate, updated_at
FROM `{{dataset}}.level_performance_summary`
WHERE app_id = '{{app_id}}'
    AND version IS NOT NULL
    {{#if lookback_days}}AND DATE(updated_at) >= DATE_SUB(CURRENT_DATE(), INTERVAL {{lookback_days}} DAY){{/if}}
ORDER BY level_number ASC
