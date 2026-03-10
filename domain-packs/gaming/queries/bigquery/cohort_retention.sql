-- Gaming: Cohort retention query
-- Fetches retention cohort data for embedding.
--
-- Parameters:
--   {{dataset}}       — BigQuery dataset name
--   {{app_id}}        — Application ID filter
--   {{lookback_days}} — How far back to look

SELECT
    app_id, cohort_date, country_code, cohort_size,
    day_1_retention, day_3_retention, day_7_retention,
    day_14_retention, day_30_retention
FROM `{{dataset}}.daily_retention`
WHERE app_id = '{{app_id}}'
    AND cohort_date >= DATE_SUB(CURRENT_DATE(), INTERVAL {{lookback_days}} DAY)
ORDER BY cohort_date DESC
