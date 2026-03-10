-- Gaming: User profiles query
-- Joins user_core_features, user_churn_features, and iap_propensity_features_v2
-- to build a complete player profile for embedding generation.
--
-- Parameters (replaced at runtime):
--   {{dataset}}       — BigQuery dataset name
--   {{app_id}}        — Application ID filter
--   {{lookback_days}} — How far back to look (0 = no limit)

SELECT
    core.app_id, core.user_id, core.last_updated_at,
    core.total_session_count, core.avg_session_duration,
    core.user_max_level_reached, core.user_success_rate,
    core.is_payer, core.user_booster_acceptance_rate,
    churn.country_code, churn.platform,
    churn.days_since_last_event, churn.user_tenure_days,
    churn.sessions_per_active_day, churn.active_days_in_recency_window,
    churn.current_level, churn.level_completion_rate,
    churn.avg_levels_per_session,
    churn.total_purchases, churn.days_since_last_purchase,
    iap.will_purchase_next_3_days,
    iap.total_levels_started, iap.total_levels_completed, iap.total_levels_failed,
    iap.total_boosters_used, iap.total_ads_offered, iap.total_ads_watched,
    iap.lifetime_ad_watch_rate,
    iap.total_spend, iap.avg_purchase_amount, iap.max_purchase_amount,
    iap.days_since_first_purchase, iap.purchase_days, iap.purchases_per_purchase_day,
    iap.recent_7d_sessions, iap.recent_7d_active_days,
    iap.recent_7d_levels_started, iap.recent_7d_levels_completed,
    iap.recent_7d_levels_failed, iap.recent_7d_success_rate,
    iap.recent_7d_boosters_used, iap.recent_7d_ads_offered,
    iap.recent_7d_ads_watched, iap.recent_7d_ad_watch_rate,
    iap.recent_7d_purchases,
    iap.recent_14d_sessions, iap.recent_14d_active_days,
    iap.recent_14d_levels_started, iap.recent_14d_purchases,
    iap.session_frequency_trend, iap.success_rate_trend,
    iap.max_levels_in_session, iap.activity_rate, iap.failure_rate,
    iap.soft_currency_balance, iap.hard_currency_balance
FROM `{{dataset}}.user_core_features` AS core
LEFT JOIN `{{dataset}}.user_churn_features` AS churn
    ON core.app_id = churn.app_id AND core.user_id = churn.user_id
LEFT JOIN `{{dataset}}.iap_propensity_features_v2` AS iap
    ON core.app_id = iap.app_id AND core.user_id = iap.user_id
WHERE core.app_id = '{{app_id}}'
    {{#if lookback_days}}AND DATE(core.last_updated_at) >= DATE_SUB(CURRENT_DATE(), INTERVAL {{lookback_days}} DAY){{/if}}
ORDER BY core.total_session_count DESC
