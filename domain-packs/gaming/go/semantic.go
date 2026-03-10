package gaming

// GamingSemanticGenerator implements domainpack.SemanticGenerator for gaming.
// Uses the full original semantic text generation logic with configurable thresholds.
type GamingSemanticGenerator struct{}

func (g *GamingSemanticGenerator) SupportedEntityTypes() []string {
	return []string{"users", "entities", "sessions", "cohorts", "rulesets"}
}

// GenerateText converts raw gaming data into human-readable text for embedding.
// Dispatches to the original typed BuildXxxProfile functions with full classification logic.
func (g *GamingSemanticGenerator) GenerateText(entityType string, data map[string]interface{}) string {
	switch entityType {
	case "users":
		return BuildPlayerProfile(convertMapToPlayerProfile(data), DefaultPlayerConfig())
	case "entities":
		return BuildLevelProfile(convertMapToLevelProfile(data), DefaultLevelConfig())
	case "sessions":
		return BuildSessionProfile(convertMapToSessionProfile(data), DefaultSessionConfig())
	case "cohorts":
		return BuildCohortProfile(convertMapToCohortProfile(data), DefaultCohortConfig())
	case "rulesets":
		return BuildRulesetProfile(convertMapToRulesetProfile(data), DefaultRulesetConfig())
	default:
		return ""
	}
}

// Conversion functions: map[string]interface{} → typed gaming structs.

func convertMapToPlayerProfile(data map[string]interface{}) *PlayerProfile {
	return &PlayerProfile{
		AppID:                   toString(data["app_id"]),
		UserID:                  toString(data["user_id"]),
		TotalSessions:           toInt(data["total_session_count"]),
		AvgSessionDuration:      toFloat(data["avg_session_duration"]) / 60.0,
		MaxLevelReached:         toInt(data["user_max_level_reached"]),
		SuccessRate:             toFloat(data["user_success_rate"]),
		IsPayer:                 toBool(data["is_payer"]),
		DaysSinceInstall:        toInt(data["user_tenure_days"]),
		DaysSinceLastSession:    toInt(data["days_since_last_event"]),
		Platform:                toStringDefault(data["platform"], "unknown"),
		Country:                 toStringDefault(data["country_code"], "unknown"),
		SessionsPerActiveDay:    toFloat(data["sessions_per_active_day"]),
		ActiveDays:              toInt(data["active_days_in_recency_window"]),
		CurrentLevel:            toInt(data["current_level"]),
		LevelCompletionRate:     toFloat(data["level_completion_rate"]),
		AvgLevelsPerSession:     toFloat(data["avg_levels_per_session"]),
		TotalPurchases:          toInt(data["total_purchases"]),
		DaysSinceLastPurchase:   toInt(data["days_since_last_purchase"]),
		WillPurchaseNext3Days:   toBool(data["will_purchase_next_3_days"]),
		TotalLevelsStarted:      toInt(data["total_levels_started"]),
		TotalLevelsCompleted:    toInt(data["total_levels_completed"]),
		TotalLevelsFailed:       toInt(data["total_levels_failed"]),
		TotalBoostersUsed:       toInt(data["total_boosters_used"]),
		TotalAdsOffered:         toInt(data["total_ads_offered"]),
		TotalAdsWatched:         toInt(data["total_ads_watched"]),
		LifetimeAdWatchRate:     toFloat(data["lifetime_ad_watch_rate"]),
		TotalSpend:              toFloat(data["total_spend"]),
		AvgPurchaseAmount:       toFloat(data["avg_purchase_amount"]),
		MaxPurchaseAmount:       toFloat(data["max_purchase_amount"]),
		DaysSinceFirstPurchase:  toInt(data["days_since_first_purchase"]),
		PurchaseDays:            toInt(data["purchase_days"]),
		PurchasesPerPurchaseDay: toFloat(data["purchases_per_purchase_day"]),
		Recent7dSessions:        toInt(data["recent_7d_sessions"]),
		Recent7dActiveDays:      toInt(data["recent_7d_active_days"]),
		Recent7dLevelsStarted:   toInt(data["recent_7d_levels_started"]),
		Recent7dLevelsCompleted: toInt(data["recent_7d_levels_completed"]),
		Recent7dLevelsFailed:    toInt(data["recent_7d_levels_failed"]),
		Recent7dSuccessRate:     toFloat(data["recent_7d_success_rate"]),
		Recent7dBoostersUsed:    toInt(data["recent_7d_boosters_used"]),
		Recent7dAdsOffered:      toInt(data["recent_7d_ads_offered"]),
		Recent7dAdsWatched:      toInt(data["recent_7d_ads_watched"]),
		Recent7dAdWatchRate:     toFloat(data["recent_7d_ad_watch_rate"]),
		Recent7dPurchases:       toInt(data["recent_7d_purchases"]),
		Recent14dSessions:       toInt(data["recent_14d_sessions"]),
		Recent14dActiveDays:     toInt(data["recent_14d_active_days"]),
		Recent14dLevelsStarted:  toInt(data["recent_14d_levels_started"]),
		Recent14dPurchases:      toInt(data["recent_14d_purchases"]),
		SessionFrequencyTrend:   toFloat(data["session_frequency_trend"]),
		SuccessRateTrend:        toFloat(data["success_rate_trend"]),
		MaxLevelsInSession:      toInt(data["max_levels_in_session"]),
		ActivityRate:            toFloat(data["activity_rate"]),
		FailureRate:             toFloat(data["failure_rate"]),
		SoftCurrencyBalance:     toInt(data["soft_currency_balance"]),
		HardCurrencyBalance:     toInt(data["hard_currency_balance"]),
	}
}

func convertMapToLevelProfile(data map[string]interface{}) *LevelProfile {
	return &LevelProfile{
		AppID:                      toString(data["app_id"]),
		LevelNumber:               toInt(data["level_number"]),
		UniquePlayers:              toInt(data["unique_players"]),
		TotalStarts:               toInt(data["total_starts"]),
		TotalSuccesses:            toInt(data["total_successes"]),
		TotalFailures:             toInt(data["total_failures"]),
		SuccessRate:               toFloat(data["success_rate"]),
		AvgAttemptsPerPlayer:      toFloat(data["avg_attempts_per_player"]),
		QuitRate:                  toFloat(data["quit_rate"]),
		LevelBoosterAcceptanceRate: toFloat(data["booster_acceptance_rate"]),
	}
}

func convertMapToSessionProfile(data map[string]interface{}) *SessionProfile {
	return &SessionProfile{
		AppID:           toString(data["app_id"]),
		SessionID:       toString(data["session_id"]),
		UserID:          toString(data["user_id"]),
		DurationMinutes: toFloat(data["duration"]) / 60.0,
		Status:          toStringDefault(data["status"], "unknown"),
		Platform:        toStringDefault(data["platform"], "unknown"),
		CountryCode:     toStringDefault(data["country_code"], "unknown"),
	}
}

func convertMapToCohortProfile(data map[string]interface{}) *CohortProfile {
	return &CohortProfile{
		AppID:          toString(data["app_id"]),
		CohortDate:     toString(data["cohort_date"]),
		CohortSize:     toInt(data["cohort_size"]),
		Day1Retention:  toFloat(data["day_1_retention"]),
		Day3Retention:  toFloat(data["day_3_retention"]),
		Day7Retention:  toFloat(data["day_7_retention"]),
		Day14Retention: toFloat(data["day_14_retention"]),
		Day30Retention: toFloat(data["day_30_retention"]),
		CountryCode:    toStringDefault(data["country_code"], "all"),
	}
}

func convertMapToRulesetProfile(data map[string]interface{}) *RulesetProfile {
	return &RulesetProfile{
		AppID:           toString(data["app_id"]),
		RulesetID:       toString(data["ruleset_id"]),
		Name:            toString(data["name"]),
		Description:     toString(data["description"]),
		LongDescription: toString(data["long_description"]),
		TriggerJSON:     toString(data["trigger_json"]),
		ActionsJSON:     toString(data["actions_json"]),
		Version:         toInt(data["version"]),
		IsActive:        toBool(data["is_active"]),
	}
}

func toStringDefault(v interface{}, def string) string {
	s := toString(v)
	if s == "" || s == "<nil>" {
		return def
	}
	return s
}
