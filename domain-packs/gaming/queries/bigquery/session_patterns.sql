-- Gaming: Session patterns query
-- Fetches recent session data for embedding.
--
-- Parameters:
--   {{dataset}}       — BigQuery dataset name
--   {{app_id}}        — Application ID filter
--   {{lookback_days}} — How far back to look

SELECT
    app_id, user_id, session_id, start_time,
    duration, status, platform, country_code
FROM `{{dataset}}.sessions`
WHERE app_id = '{{app_id}}'
    AND start_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL {{lookback_days}} DAY)
ORDER BY start_time DESC
LIMIT 10000
