package gaming

import (
	"fmt"
	"strings"
	"time"
)

// RulesetProfile represents ruleset analytics data
type RulesetProfile struct {
	AppID           string
	RulesetID       string
	Name            string
	Description     string
	LongDescription string
	TriggerJSON     string // Raw JSON trigger configuration
	ActionsJSON     string // Raw JSON actions configuration
	ABTestJSON      string // Raw JSON A/B test configuration
	Version         int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DaysSinceCreated int
	DaysSinceUpdated int
}

// RulesetConfig contains thresholds for ruleset classification
type RulesetConfig struct {
	// Age thresholds (days)
	NewRulesetMaxDays    int
	RecentUpdateMaxDays  int
	StaleRulesetMinDays  int

	// Version thresholds
	HighlyIteratedMinVersions int
	NewRulesetMaxVersions     int
}

// DefaultRulesetConfig returns default configuration for ruleset classification
func DefaultRulesetConfig() *RulesetConfig {
	return &RulesetConfig{
		// Age thresholds
		NewRulesetMaxDays:   7,   // Created within 7 days
		RecentUpdateMaxDays: 30,  // Updated within 30 days
		StaleRulesetMinDays: 180, // Not updated in 180 days

		// Version thresholds
		HighlyIteratedMinVersions: 10,
		NewRulesetMaxVersions:     2,
	}
}

// BuildRulesetProfile generates semantic text description for a ruleset
func BuildRulesetProfile(ruleset *RulesetProfile, config *RulesetConfig) string {
	if config == nil {
		config = DefaultRulesetConfig()
	}

	var parts []string

	// 1. Ruleset identifier and status
	parts = append(parts, fmt.Sprintf("Ruleset: %s", ruleset.Name))
	if ruleset.IsActive {
		parts = append(parts, "active")
	} else {
		parts = append(parts, "inactive")
	}

	// 2. Version information
	versionDesc := classifyRulesetVersion(ruleset, config)
	parts = append(parts, versionDesc)

	// 3. Description (main semantic content)
	if ruleset.Description != "" {
		parts = append(parts, ruleset.Description)
	}

	// 4. Long description (if available and different from short)
	if ruleset.LongDescription != "" && ruleset.LongDescription != ruleset.Description {
		// Truncate if too long (keep first 200 chars)
		longDesc := ruleset.LongDescription
		if len(longDesc) > 200 {
			longDesc = longDesc[:200] + "..."
		}
		parts = append(parts, longDesc)
	}

	// 5. A/B testing status
	if ruleset.ABTestJSON != "" && ruleset.ABTestJSON != "null" {
		parts = append(parts, "includes A/B testing")
	}

	// 6. Age and update recency
	ageDesc := classifyRulesetAge(ruleset, config)
	parts = append(parts, ageDesc)

	// 7. Trigger and action summary (simple mention, no parsing)
	if ruleset.TriggerJSON != "" && ruleset.TriggerJSON != "null" {
		parts = append(parts, "event-triggered ruleset")
	}
	if ruleset.ActionsJSON != "" && ruleset.ActionsJSON != "null" {
		parts = append(parts, "with automated actions")
	}

	return strings.Join(parts, ", ")
}

// classifyRulesetVersion determines version iteration level
func classifyRulesetVersion(ruleset *RulesetProfile, config *RulesetConfig) string {
	version := ruleset.Version

	if version >= config.HighlyIteratedMinVersions {
		return fmt.Sprintf("version %d (highly iterated)", version)
	} else if version <= config.NewRulesetMaxVersions {
		return fmt.Sprintf("version %d (new)", version)
	}
	return fmt.Sprintf("version %d", version)
}

// classifyRulesetAge describes ruleset age and update recency
func classifyRulesetAge(ruleset *RulesetProfile, config *RulesetConfig) string {
	var parts []string

	// Created age
	if ruleset.DaysSinceCreated <= config.NewRulesetMaxDays {
		parts = append(parts, fmt.Sprintf("created %d days ago (new)", ruleset.DaysSinceCreated))
	} else if ruleset.DaysSinceCreated >= config.StaleRulesetMinDays {
		parts = append(parts, fmt.Sprintf("created %d days ago (legacy)", ruleset.DaysSinceCreated))
	} else {
		parts = append(parts, fmt.Sprintf("created %d days ago", ruleset.DaysSinceCreated))
	}

	// Updated age (if different from created)
	if ruleset.DaysSinceUpdated > 0 && ruleset.DaysSinceUpdated != ruleset.DaysSinceCreated {
		if ruleset.DaysSinceUpdated <= config.RecentUpdateMaxDays {
			parts = append(parts, fmt.Sprintf("updated %d days ago (active)", ruleset.DaysSinceUpdated))
		} else if ruleset.DaysSinceUpdated >= config.StaleRulesetMinDays {
			parts = append(parts, fmt.Sprintf("not updated in %d days (stale)", ruleset.DaysSinceUpdated))
		} else {
			parts = append(parts, fmt.Sprintf("last updated %d days ago", ruleset.DaysSinceUpdated))
		}
	}

	return strings.Join(parts, ", ")
}
