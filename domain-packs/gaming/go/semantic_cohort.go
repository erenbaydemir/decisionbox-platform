package gaming

import (
	"fmt"
	"strings"
)

// CohortProfile represents cohort retention analytics data
type CohortProfile struct {
	AppID                 string
	Version               string
	CountryCode           string
	CohortDate            string
	CohortSize            int
	Day1Retention         float64 // 0-1
	Day3Retention         float64 // 0-1
	Day7Retention         float64 // 0-1
	Day14Retention        float64 // 0-1
	Day30Retention        float64 // 0-1
	Day60Retention        float64 // 0-1
	Day90Retention        float64 // 0-1
	Day180Retention       float64 // 0-1
	RetentionQualityScore float64
	RetentionGrade        string
}

// CohortConfig contains thresholds for cohort classification
type CohortConfig struct {
	// Retention thresholds (0-1)
	ExcellentD1Retention  float64
	GoodD1Retention       float64
	PoorD1Retention       float64

	ExcellentD7Retention  float64
	GoodD7Retention       float64
	PoorD7Retention       float64

	ExcellentD30Retention float64
	GoodD30Retention      float64
	PoorD30Retention      float64

	// Drop-off thresholds (percentage points)
	SteepDropOffThreshold    float64 // D1 to D7 drop
	ModerateDropOffThreshold float64

	// Cohort size thresholds
	LargeCohortMin   int
	MediumCohortMin  int
	SmallCohortMax   int
}

// DefaultCohortConfig returns default configuration for cohort classification
func DefaultCohortConfig() *CohortConfig {
	return &CohortConfig{
		// D1 retention thresholds
		ExcellentD1Retention: 0.40,
		GoodD1Retention:      0.25,
		PoorD1Retention:      0.15,

		// D7 retention thresholds
		ExcellentD7Retention: 0.20,
		GoodD7Retention:      0.10,
		PoorD7Retention:      0.05,

		// D30 retention thresholds
		ExcellentD30Retention: 0.10,
		GoodD30Retention:      0.05,
		PoorD30Retention:      0.02,

		// Drop-off thresholds (percentage points)
		SteepDropOffThreshold:    0.30, // >30 percentage points drop
		ModerateDropOffThreshold: 0.15, // 15-30 percentage points drop

		// Cohort size
		LargeCohortMin:  1000,
		MediumCohortMin: 100,
		SmallCohortMax:  50,
	}
}

// BuildCohortProfile generates semantic text description for a cohort
func BuildCohortProfile(cohort *CohortProfile, config *CohortConfig) string {
	if config == nil {
		config = DefaultCohortConfig()
	}

	var parts []string

	// 1. Cohort identifier
	parts = append(parts, fmt.Sprintf("Cohort from %s", cohort.CohortDate))
	if cohort.Version != "" && cohort.Version != "all" {
		parts = append(parts, fmt.Sprintf("version %s", cohort.Version))
	}
	if cohort.CountryCode != "" && cohort.CountryCode != "all" {
		parts = append(parts, fmt.Sprintf("country %s", cohort.CountryCode))
	}

	// 2. Cohort size
	sizeDesc := classifyCohortSize(cohort, config)
	parts = append(parts, sizeDesc)

	// 3. Overall retention quality
	if cohort.RetentionGrade != "" {
		parts = append(parts, fmt.Sprintf("retention grade: %s", cohort.RetentionGrade))
	}
	if cohort.RetentionQualityScore > 0 {
		parts = append(parts, fmt.Sprintf("quality score %.2f", cohort.RetentionQualityScore))
	}

	// 4. Short-term retention (D1-D7)
	shortTermDesc := classifyShortTermRetention(cohort, config)
	parts = append(parts, shortTermDesc)

	// 5. Mid-term retention (D14-D30)
	midTermDesc := classifyMidTermRetention(cohort, config)
	parts = append(parts, midTermDesc)

	// 6. Long-term retention (D60-D180)
	if cohort.Day90Retention > 0 || cohort.Day180Retention > 0 {
		longTermDesc := classifyLongTermRetention(cohort)
		parts = append(parts, longTermDesc)
	}

	// 7. Retention curve shape
	curveDesc := classifyRetentionCurve(cohort, config)
	parts = append(parts, curveDesc)

	// 8. Detailed metrics
	// Note: Retention values are in decimal format (0-1), multiply by 100 for display
	metricsDesc := fmt.Sprintf("D1: %.1f%%, D7: %.1f%%, D30: %.1f%%",
		cohort.Day1Retention*100, cohort.Day7Retention*100, cohort.Day30Retention*100)
	parts = append(parts, metricsDesc)

	return strings.Join(parts, ", ")
}

// classifyCohortSize determines cohort size category
func classifyCohortSize(cohort *CohortProfile, config *CohortConfig) string {
	size := cohort.CohortSize

	if size >= config.LargeCohortMin {
		return fmt.Sprintf("large cohort (%d users)", size)
	} else if size >= config.MediumCohortMin {
		return fmt.Sprintf("medium cohort (%d users)", size)
	} else if size <= config.SmallCohortMax {
		return fmt.Sprintf("small cohort (%d users)", size)
	}
	return fmt.Sprintf("%d users", size)
}

// classifyShortTermRetention describes D1-D7 retention performance
func classifyShortTermRetention(cohort *CohortProfile, config *CohortConfig) string {
	d1 := cohort.Day1Retention
	d7 := cohort.Day7Retention

	var d1Label, d7Label string

	// D1 classification
	if d1 >= config.ExcellentD1Retention {
		d1Label = "excellent"
	} else if d1 >= config.GoodD1Retention {
		d1Label = "good"
	} else if d1 <= config.PoorD1Retention {
		d1Label = "poor"
	} else {
		d1Label = "moderate"
	}

	// D7 classification
	if d7 >= config.ExcellentD7Retention {
		d7Label = "excellent"
	} else if d7 >= config.GoodD7Retention {
		d7Label = "good"
	} else if d7 <= config.PoorD7Retention {
		d7Label = "poor"
	} else {
		d7Label = "moderate"
	}

	return fmt.Sprintf("%s D1 retention, %s D7 retention", d1Label, d7Label)
}

// classifyMidTermRetention describes D14-D30 retention performance
func classifyMidTermRetention(cohort *CohortProfile, config *CohortConfig) string {
	d30 := cohort.Day30Retention

	var d30Label string
	if d30 >= config.ExcellentD30Retention {
		d30Label = "excellent"
	} else if d30 >= config.GoodD30Retention {
		d30Label = "good"
	} else if d30 <= config.PoorD30Retention {
		d30Label = "poor"
	} else {
		d30Label = "moderate"
	}

	return fmt.Sprintf("%s D30 retention", d30Label)
}

// classifyLongTermRetention describes D60-D180 retention performance
func classifyLongTermRetention(cohort *CohortProfile) string {
	var parts []string

	if cohort.Day60Retention > 0 {
		parts = append(parts, fmt.Sprintf("D60: %.1f%%", cohort.Day60Retention*100))
	}
	if cohort.Day90Retention > 0 {
		parts = append(parts, fmt.Sprintf("D90: %.1f%%", cohort.Day90Retention*100))
	}
	if cohort.Day180Retention > 0 {
		parts = append(parts, fmt.Sprintf("D180: %.1f%%", cohort.Day180Retention*100))
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("long-term: %s", strings.Join(parts, ", "))
}

// classifyRetentionCurve describes the shape of the retention curve
func classifyRetentionCurve(cohort *CohortProfile, config *CohortConfig) string {
	// Calculate drop-offs
	d1ToD7Drop := cohort.Day1Retention - cohort.Day7Retention
	d7ToD30Drop := cohort.Day7Retention - cohort.Day30Retention

	var curveShape string

	// Early drop-off pattern
	if d1ToD7Drop >= config.SteepDropOffThreshold {
		curveShape = "steep early drop-off"
	} else if d1ToD7Drop >= config.ModerateDropOffThreshold {
		curveShape = "moderate early drop-off"
	} else {
		curveShape = "gradual early decline"
	}

	// Mid-term stability
	if d7ToD30Drop < 0.05 {
		curveShape += ", stable mid-term"
	} else if d7ToD30Drop >= 0.10 {
		curveShape += ", continued mid-term decline"
	}

	return curveShape
}
