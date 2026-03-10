package gaming

import (
	"fmt"
	"strings"
)

// PlayerProfile represents player analytics data (only fields available in BigQuery)
type PlayerProfile struct {
	AppID                   string
	UserID                  string
	TotalSessions           int
	AvgSessionDuration      float64 // minutes
	MaxLevelReached         int
	SuccessRate             float64 // 0-1
	IsPayer                 bool
	DaysSinceInstall        int     // user_tenure_days
	DaysSinceLastSession    int     // days_since_last_event
	Platform                string  // platform
	Country                 string  // country_code
	SessionsPerActiveDay    float64 // sessions_per_active_day
	ActiveDays              int     // active_days_in_recency_window
	CurrentLevel            int     // current_level
	LevelCompletionRate     float64 // level_completion_rate
	AvgLevelsPerSession     float64 // avg_levels_per_session
	TotalPurchases          int     // total_purchases
	DaysSinceLastPurchase   int     // days_since_last_purchase

	// From iap_propensity_features_v2
	WillPurchaseNext3Days   bool    // will_purchase_next_3_days (ML prediction)
	TotalLevelsStarted      int     // total_levels_started
	TotalLevelsCompleted    int     // total_levels_completed
	TotalLevelsFailed       int     // total_levels_failed
	TotalBoostersUsed       int     // total_boosters_used (was total_powerups_collected)
	TotalAdsOffered         int     // total_ads_offered
	TotalAdsWatched         int     // total_ads_watched (was total_rewarded_ads_watched)
	LifetimeAdWatchRate     float64 // lifetime_ad_watch_rate
	TotalSpend              float64 // total_spend
	AvgPurchaseAmount       float64 // avg_purchase_amount
	MaxPurchaseAmount       float64 // max_purchase_amount
	DaysSinceFirstPurchase  int     // days_since_first_purchase
	PurchaseDays            int     // purchase_days
	PurchasesPerPurchaseDay float64 // purchases_per_purchase_day
	Recent7dSessions        int     // recent_7d_sessions
	Recent7dActiveDays      int     // recent_7d_active_days
	Recent7dLevelsStarted   int     // recent_7d_levels_started
	Recent7dLevelsCompleted int     // recent_7d_levels_completed
	Recent7dLevelsFailed    int     // recent_7d_levels_failed
	Recent7dSuccessRate     float64 // recent_7d_success_rate
	Recent7dBoostersUsed    int     // recent_7d_boosters_used
	Recent7dAdsOffered      int     // recent_7d_ads_offered
	Recent7dAdsWatched      int     // recent_7d_ads_watched
	Recent7dAdWatchRate     float64 // recent_7d_ad_watch_rate
	Recent7dPurchases       int     // recent_7d_purchases
	Recent14dSessions       int     // recent_14d_sessions
	Recent14dActiveDays     int     // recent_14d_active_days
	Recent14dLevelsStarted  int     // recent_14d_levels_started
	Recent14dPurchases      int     // recent_14d_purchases
	SessionFrequencyTrend   float64 // session_frequency_trend
	SuccessRateTrend        float64 // success_rate_trend
	MaxLevelsInSession      int     // max_levels_in_session
	ActivityRate            float64 // activity_rate
	FailureRate             float64 // failure_rate
	SoftCurrencyBalance     int     // soft_currency_balance (was latest_soft_currency)
	HardCurrencyBalance     int     // hard_currency_balance (was latest_hard_currency)
}

// PlayerConfig contains thresholds and labels for player classification
type PlayerConfig struct {
	// Archetype thresholds
	PowerUserMinSessions    int
	CasualPlayerMaxSessions int

	// Engagement thresholds
	HighEngagementMinSessions float64
	LowEngagementMaxSessions  float64

	// Success rate thresholds
	VeryHighSuccessRate float64
	HighSuccessRate     float64
	LowSuccessRate      float64

	// Session duration thresholds (minutes)
	LongSessionMin  float64
	ShortSessionMax float64

	// Activity thresholds (sessions per active day)
	VeryActiveSessionsPerDay float64
	ActiveSessionsPerDay     float64
	InactiveSessionsPerDay   float64

	// Recency thresholds (days)
	RecentActivityDays int
	DormantDays        int
}

// DefaultPlayerConfig returns default configuration for player classification
func DefaultPlayerConfig() *PlayerConfig {
	return &PlayerConfig{
		// Archetype
		PowerUserMinSessions:    100,
		CasualPlayerMaxSessions: 10,

		// Engagement
		HighEngagementMinSessions: 50,
		LowEngagementMaxSessions:  5,

		// Success
		VeryHighSuccessRate: 0.8,
		HighSuccessRate:     0.6,
		LowSuccessRate:      0.3,

		// Session duration
		LongSessionMin:  30.0,
		ShortSessionMax: 5.0,

		// Activity (sessions per active day)
		VeryActiveSessionsPerDay: 5.0,
		ActiveSessionsPerDay:     2.0,
		InactiveSessionsPerDay:   0.5,

		// Recency
		RecentActivityDays: 7,
		DormantDays:        30,
	}
}

// BuildPlayerProfile generates semantic text description for a player
func BuildPlayerProfile(player *PlayerProfile, config *PlayerConfig) string {
	if config == nil {
		config = DefaultPlayerConfig()
	}

	var parts []string

	// 1. Archetype (primary classification)
	archetype := classifyPlayerArchetype(player, config)
	parts = append(parts, fmt.Sprintf("Player profile: %s", archetype))

	// 2. Activity level and tenure
	activityLevel := classifyActivityLevel(player, config)
	if player.ActiveDays > 0 {
		parts = append(parts, fmt.Sprintf("%d sessions over %d days (%d active days), %s engagement",
			player.TotalSessions, player.DaysSinceInstall, player.ActiveDays, activityLevel))
	} else {
		parts = append(parts, fmt.Sprintf("%d sessions over %d days, %s engagement",
			player.TotalSessions, player.DaysSinceInstall, activityLevel))
	}

	// 3. Session behavior
	sessionBehavior := classifySessionBehavior(player, config)
	parts = append(parts, sessionBehavior)

	// 4. Performance and progression
	performanceDesc := classifyPerformance(player, config)
	parts = append(parts, performanceDesc)

	// 5. Level progression details
	if player.CurrentLevel > 0 || player.AvgLevelsPerSession > 0 {
		levelDetails := classifyLevelProgression(player)
		parts = append(parts, levelDetails)
	}

	// 6. Monetization and economy
	monetizationDesc := classifyMonetization(player)
	parts = append(parts, monetizationDesc)

	// 7. Engagement features (boosters, ads)
	engagementFeatures := classifyEngagementFeatures(player)
	if engagementFeatures != "" {
		parts = append(parts, engagementFeatures)
	}

	// 8. Economy status (currencies, spending)
	economyStatus := classifyEconomy(player)
	if economyStatus != "" {
		parts = append(parts, economyStatus)
	}

	// 9. Trends (session frequency and success rate)
	trendsDesc := classifyTrends(player)
	if trendsDesc != "" {
		parts = append(parts, trendsDesc)
	}

	// 10. Recent activity (7d/14d)
	recentActivity := classifyRecentActivity(player)
	if recentActivity != "" {
		parts = append(parts, recentActivity)
	}

	// 11. IAP propensity prediction
	if player.WillPurchaseNext3Days {
		parts = append(parts, "predicted to purchase in next 3 days")
	}

	// 12. Recency
	recencyDesc := classifyRecency(player)
	parts = append(parts, recencyDesc)

	// 13. Platform and location context
	parts = append(parts, fmt.Sprintf("%s player from %s", player.Platform, player.Country))

	return strings.Join(parts, ", ")
}

// classifyPlayerArchetype determines the primary player archetype
func classifyPlayerArchetype(player *PlayerProfile, config *PlayerConfig) string {
	// Power user: High session count
	if player.TotalSessions >= config.PowerUserMinSessions {
		if player.IsPayer {
			return "power user spender"
		}
		return "power user"
	}

	// Casual: Low session count
	if player.TotalSessions <= config.CasualPlayerMaxSessions {
		if player.IsPayer {
			return "casual spender"
		}
		return "casual player"
	}

	// Regular player
	if player.IsPayer {
		return "regular spender"
	}
	return "regular player"
}

// classifyActivityLevel determines engagement level
func classifyActivityLevel(player *PlayerProfile, config *PlayerConfig) string {
	// Use sessions per active day for activity classification
	if player.SessionsPerActiveDay >= config.VeryActiveSessionsPerDay {
		return "very high"
	}
	if player.SessionsPerActiveDay >= config.ActiveSessionsPerDay {
		return "high"
	}
	if player.SessionsPerActiveDay >= config.InactiveSessionsPerDay {
		return "moderate"
	}
	if player.SessionsPerActiveDay > 0 {
		return "low"
	}
	return "dormant"
}

// classifySessionBehavior describes session patterns
func classifySessionBehavior(player *PlayerProfile, config *PlayerConfig) string {
	var behavior []string

	// Session duration
	if player.AvgSessionDuration >= config.LongSessionMin {
		behavior = append(behavior, "long sessions")
	} else if player.AvgSessionDuration <= config.ShortSessionMax {
		behavior = append(behavior, "short sessions")
	} else {
		behavior = append(behavior, "medium sessions")
	}

	// Duration value
	behavior = append(behavior, fmt.Sprintf("avg %.1f minutes", player.AvgSessionDuration))

	return strings.Join(behavior, ", ")
}

// classifyPerformance describes player skill and progression
func classifyPerformance(player *PlayerProfile, config *PlayerConfig) string {
	var perf []string

	// Progression
	perf = append(perf, fmt.Sprintf("reached level %d", player.MaxLevelReached))

	// Success rate
	var successLabel string
	if player.SuccessRate >= config.VeryHighSuccessRate {
		successLabel = "very high"
	} else if player.SuccessRate >= config.HighSuccessRate {
		successLabel = "high"
	} else if player.SuccessRate <= config.LowSuccessRate {
		successLabel = "low"
	} else {
		successLabel = "moderate"
	}

	perf = append(perf, fmt.Sprintf("%s success rate %.1f%%", successLabel, player.SuccessRate*100))

	return strings.Join(perf, ", ")
}

// classifyMonetization describes spending behavior
func classifyMonetization(player *PlayerProfile) string {
	if player.IsPayer {
		return "paying user"
	}
	return "non-paying user"
}

// classifyRecency describes last session recency
func classifyRecency(player *PlayerProfile) string {
	days := player.DaysSinceLastSession

	if days == 0 {
		return "active today"
	} else if days == 1 {
		return "active yesterday"
	} else if days <= 3 {
		return fmt.Sprintf("active %d days ago", days)
	} else if days <= 7 {
		return fmt.Sprintf("last seen %d days ago", days)
	} else if days <= 30 {
		return fmt.Sprintf("dormant %d days", days)
	} else {
		return fmt.Sprintf("inactive %d days", days)
	}
}

// classifyLevelProgression describes level progression patterns
func classifyLevelProgression(player *PlayerProfile) string {
	var details []string

	if player.CurrentLevel > 0 {
		details = append(details, fmt.Sprintf("currently on level %d", player.CurrentLevel))
	}

	if player.LevelCompletionRate > 0 {
		details = append(details, fmt.Sprintf("%.0f%% level completion rate", player.LevelCompletionRate*100))
	}

	if player.AvgLevelsPerSession > 0 {
		details = append(details, fmt.Sprintf("%.1f levels per session", player.AvgLevelsPerSession))
	}

	if len(details) > 0 {
		return strings.Join(details, ", ")
	}
	return ""
}

// classifyEngagementFeatures describes engagement with boosters and ads
func classifyEngagementFeatures(player *PlayerProfile) string {
	var features []string

	if player.TotalBoostersUsed > 0 {
		features = append(features, fmt.Sprintf("%d boosters used", player.TotalBoostersUsed))
	}

	if player.TotalAdsWatched > 0 {
		features = append(features, fmt.Sprintf("%d ads watched", player.TotalAdsWatched))
		if player.LifetimeAdWatchRate > 0 {
			features = append(features, fmt.Sprintf("%.0f%% ad watch rate", player.LifetimeAdWatchRate*100))
		}
	} else if player.TotalAdsOffered > 0 {
		features = append(features, fmt.Sprintf("%d ads offered but none watched", player.TotalAdsOffered))
	} else if player.TotalSessions > 5 {
		features = append(features, "no ad engagement")
	}

	if len(features) > 0 {
		return strings.Join(features, ", ")
	}
	return ""
}

// classifyEconomy describes player's economy status
func classifyEconomy(player *PlayerProfile) string {
	var economy []string

	if player.SoftCurrencyBalance > 0 || player.HardCurrencyBalance > 0 {
		if player.SoftCurrencyBalance > 0 {
			economy = append(economy, fmt.Sprintf("%d soft currency", player.SoftCurrencyBalance))
		}
		if player.HardCurrencyBalance > 0 {
			economy = append(economy, fmt.Sprintf("%d hard currency", player.HardCurrencyBalance))
		}
	}

	if player.TotalPurchases > 0 {
		if player.DaysSinceLastPurchase >= 0 && player.DaysSinceLastPurchase < 9999 {
			economy = append(economy, fmt.Sprintf("%d purchases (last %d days ago)", player.TotalPurchases, player.DaysSinceLastPurchase))
		} else {
			economy = append(economy, fmt.Sprintf("%d total purchases", player.TotalPurchases))
		}
	}

	if player.TotalSpend > 0 {
		economy = append(economy, fmt.Sprintf("$%.2f total spend", player.TotalSpend))
		if player.AvgPurchaseAmount > 0 {
			economy = append(economy, fmt.Sprintf("$%.2f avg purchase", player.AvgPurchaseAmount))
		}
	}

	if len(economy) > 0 {
		return strings.Join(economy, ", ")
	}
	return ""
}

// classifyTrends describes player engagement and performance trends
func classifyTrends(player *PlayerProfile) string {
	var trends []string

	// Session frequency trend
	if player.SessionFrequencyTrend > 0.1 {
		trends = append(trends, "increasing session frequency")
	} else if player.SessionFrequencyTrend < -0.1 {
		trends = append(trends, "decreasing session frequency")
	}

	// Success rate trend
	if player.SuccessRateTrend > 0.1 {
		trends = append(trends, "improving success rate")
	} else if player.SuccessRateTrend < -0.1 {
		trends = append(trends, "declining success rate")
	}

	// Activity rate
	if player.ActivityRate > 0.7 {
		trends = append(trends, "highly active")
	} else if player.ActivityRate < 0.3 && player.ActivityRate > 0 {
		trends = append(trends, "low activity rate")
	}

	if len(trends) > 0 {
		return strings.Join(trends, ", ")
	}
	return ""
}

// classifyRecentActivity describes player's recent 7-day and 14-day activity
func classifyRecentActivity(player *PlayerProfile) string {
	var recent []string

	// Recent 7-day activity
	if player.Recent7dSessions > 0 {
		recent = append(recent, fmt.Sprintf("%d sessions in last 7 days", player.Recent7dSessions))
		if player.Recent7dActiveDays > 0 {
			recent = append(recent, fmt.Sprintf("%d active days", player.Recent7dActiveDays))
		}
	}

	// Recent 7-day levels
	if player.Recent7dLevelsCompleted > 0 {
		recent = append(recent, fmt.Sprintf("%d levels completed recently", player.Recent7dLevelsCompleted))
	}

	// Recent purchases
	if player.Recent7dPurchases > 0 {
		recent = append(recent, fmt.Sprintf("%d purchases in last 7 days", player.Recent7dPurchases))
	} else if player.Recent14dPurchases > 0 {
		recent = append(recent, fmt.Sprintf("%d purchases in last 14 days", player.Recent14dPurchases))
	}

	// Recent ad engagement
	if player.Recent7dAdsWatched > 0 && player.Recent7dAdWatchRate > 0 {
		recent = append(recent, fmt.Sprintf("%.0f%% recent ad watch rate", player.Recent7dAdWatchRate*100))
	}

	if len(recent) > 0 {
		return strings.Join(recent, ", ")
	}
	return ""
}
