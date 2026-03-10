package gaming

import (
	"fmt"
	"strings"
)

// LevelProfile represents level performance analytics data
type LevelProfile struct {
	AppID                      string
	Version                    string
	LevelNumber                int
	UniquePlayers              int
	TotalStarts                int
	TotalSuccesses             int
	TotalFailures              int
	SuccessRate                float64 // 0-1
	AvgAttemptsPerPlayer       float64
	QuitRate                   float64 // 0-1
	LevelBoosterAcceptanceRate float64 // 0-1
}

// LevelConfig contains thresholds for level classification
type LevelConfig struct {
	// Difficulty thresholds
	VeryEasySuccessRate  float64
	EasySuccessRate      float64
	HardSuccessRate      float64
	VeryHardSuccessRate  float64

	// Engagement thresholds
	HighQuitRate         float64
	ModerateQuitRate     float64

	// Attempts thresholds
	ManyAttemptsPerPlayer float64
	FewAttemptsPerPlayer  float64
}

// DefaultLevelConfig returns default configuration for level classification
func DefaultLevelConfig() *LevelConfig {
	return &LevelConfig{
		// Difficulty
		VeryEasySuccessRate: 0.9,
		EasySuccessRate:     0.7,
		HardSuccessRate:     0.4,
		VeryHardSuccessRate: 0.2,

		// Quit rate
		HighQuitRate:     0.3,
		ModerateQuitRate: 0.15,

		// Attempts
		ManyAttemptsPerPlayer: 5.0,
		FewAttemptsPerPlayer:  2.0,
	}
}

// BuildLevelProfile generates semantic text description for a level
func BuildLevelProfile(level *LevelProfile, config *LevelConfig) string {
	if config == nil {
		config = DefaultLevelConfig()
	}

	var parts []string

	// 1. Level identifier
	parts = append(parts, fmt.Sprintf("Level %d", level.LevelNumber))
	if level.Version != "" {
		parts = append(parts, fmt.Sprintf("version %s", level.Version))
	}

	// 2. Difficulty classification
	difficulty := classifyLevelDifficulty(level, config)
	parts = append(parts, difficulty)

	// 3. Player engagement
	engagement := classifyLevelEngagement(level, config)
	parts = append(parts, engagement)

	// 4. Attempt patterns
	attemptPattern := classifyAttemptPattern(level, config)
	parts = append(parts, attemptPattern)

	// 5. Quit behavior
	quitBehavior := classifyQuitBehavior(level, config)
	parts = append(parts, quitBehavior)

	// 6. Booster usage
	if level.LevelBoosterAcceptanceRate > 0 {
		parts = append(parts, fmt.Sprintf("%.0f%% booster acceptance", level.LevelBoosterAcceptanceRate*100))
	}

	// 7. Performance stats
	parts = append(parts, fmt.Sprintf("%d players, %d starts, %.0f%% success rate",
		level.UniquePlayers, level.TotalStarts, level.SuccessRate*100))

	return strings.Join(parts, ", ")
}

// classifyLevelDifficulty determines level difficulty based on success rate
func classifyLevelDifficulty(level *LevelProfile, config *LevelConfig) string {
	if level.SuccessRate >= config.VeryEasySuccessRate {
		return "very easy difficulty"
	} else if level.SuccessRate >= config.EasySuccessRate {
		return "easy difficulty"
	} else if level.SuccessRate >= config.HardSuccessRate {
		return "moderate difficulty"
	} else if level.SuccessRate >= config.VeryHardSuccessRate {
		return "hard difficulty"
	}
	return "very hard difficulty"
}

// classifyLevelEngagement describes player engagement with the level
func classifyLevelEngagement(level *LevelProfile, config *LevelConfig) string {
	playersPerStart := float64(level.UniquePlayers) / float64(level.TotalStarts)

	if playersPerStart > 0.8 {
		return "high abandonment (many try once and quit)"
	} else if playersPerStart > 0.5 {
		return "moderate persistence"
	}
	return "high retry engagement"
}

// classifyAttemptPattern describes attempt patterns
func classifyAttemptPattern(level *LevelProfile, config *LevelConfig) string {
	if level.AvgAttemptsPerPlayer >= config.ManyAttemptsPerPlayer {
		return fmt.Sprintf("%.1f avg attempts (sticky level)", level.AvgAttemptsPerPlayer)
	} else if level.AvgAttemptsPerPlayer <= config.FewAttemptsPerPlayer {
		return fmt.Sprintf("%.1f avg attempts (quick pass or quit)", level.AvgAttemptsPerPlayer)
	}
	return fmt.Sprintf("%.1f avg attempts per player", level.AvgAttemptsPerPlayer)
}

// classifyQuitBehavior describes quit behavior
func classifyQuitBehavior(level *LevelProfile, config *LevelConfig) string {
	if level.QuitRate >= config.HighQuitRate {
		return fmt.Sprintf("high quit rate %.0f%% (frustration point)", level.QuitRate*100)
	} else if level.QuitRate >= config.ModerateQuitRate {
		return fmt.Sprintf("moderate quit rate %.0f%%", level.QuitRate*100)
	}
	return fmt.Sprintf("low quit rate %.0f%%", level.QuitRate*100)
}
