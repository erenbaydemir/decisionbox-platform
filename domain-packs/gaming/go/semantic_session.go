package gaming

import (
	"fmt"
	"strings"
	"time"
)

// SessionProfile represents session analytics data
type SessionProfile struct {
	AppID          string
	SessionID      string
	UserID         string
	StartTime      time.Time
	DurationMinutes float64 // converted from seconds
	Status         string
	HasVideo       bool
	DeviceModel    string
	AppVersion     string
	Platform       string
	EndReason      string
	CountryCode    string
	LocationSource string
}

// SessionConfig contains thresholds for session classification
type SessionConfig struct {
	// Duration thresholds (minutes)
	VeryShortSessionMax  float64
	ShortSessionMax      float64
	MediumSessionMax     float64
	LongSessionMin       float64
	VeryLongSessionMin   float64

	// Time of day thresholds (hour of day)
	MorningStart   int
	AfternoonStart int
	EveningStart   int
	NightStart     int
}

// DefaultSessionConfig returns default configuration for session classification
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		// Duration thresholds
		VeryShortSessionMax: 1.0,   // < 1 min
		ShortSessionMax:     5.0,   // 1-5 min
		MediumSessionMax:    15.0,  // 5-15 min
		LongSessionMin:      15.0,  // 15-30 min
		VeryLongSessionMin:  30.0,  // > 30 min

		// Time of day
		MorningStart:   6,  // 6 AM
		AfternoonStart: 12, // 12 PM
		EveningStart:   18, // 6 PM
		NightStart:     22, // 10 PM
	}
}

// BuildSessionProfile generates semantic text description for a session
func BuildSessionProfile(session *SessionProfile, config *SessionConfig) string {
	if config == nil {
		config = DefaultSessionConfig()
	}

	var parts []string

	// 1. Session identifier and status
	parts = append(parts, fmt.Sprintf("Session %s", session.SessionID))
	if session.Status != "" {
		parts = append(parts, fmt.Sprintf("status: %s", session.Status))
	}

	// 2. Duration classification
	if session.DurationMinutes > 0 {
		durationClass := classifySessionDuration(session, config)
		parts = append(parts, durationClass)
	}

	// 3. Time of day
	timeOfDay := classifyTimeOfDay(session, config)
	parts = append(parts, timeOfDay)

	// 4. Platform and device
	platformDesc := classifyPlatformDevice(session)
	parts = append(parts, platformDesc)

	// 5. App version
	if session.AppVersion != "" && session.AppVersion != "unknown" {
		parts = append(parts, fmt.Sprintf("app version %s", session.AppVersion))
	}

	// 6. Location
	if session.CountryCode != "" && session.CountryCode != "unknown" {
		countryName := getCountryName(session.CountryCode)
		parts = append(parts, fmt.Sprintf("from %s", countryName))
	}

	// 7. Video engagement
	if session.HasVideo {
		parts = append(parts, "includes video content")
	}

	// 8. End reason
	if session.EndReason != "" && session.EndReason != "unknown" {
		parts = append(parts, fmt.Sprintf("ended due to: %s", session.EndReason))
	}

	// 9. User identifier (for similarity search)
	parts = append(parts, fmt.Sprintf("user %s", session.UserID))

	return strings.Join(parts, ", ")
}

// classifySessionDuration determines session length category
func classifySessionDuration(session *SessionProfile, config *SessionConfig) string {
	duration := session.DurationMinutes

	if duration < config.VeryShortSessionMax {
		return fmt.Sprintf("very short session (%.1f min)", duration)
	} else if duration < config.ShortSessionMax {
		return fmt.Sprintf("short session (%.1f min)", duration)
	} else if duration < config.MediumSessionMax {
		return fmt.Sprintf("medium session (%.1f min)", duration)
	} else if duration < config.VeryLongSessionMin {
		return fmt.Sprintf("long session (%.1f min)", duration)
	}
	return fmt.Sprintf("very long session (%.1f min)", duration)
}

// classifyTimeOfDay determines when the session occurred
func classifyTimeOfDay(session *SessionProfile, config *SessionConfig) string {
	hour := session.StartTime.Hour()
	weekday := session.StartTime.Weekday()

	var timeLabel string
	if hour >= config.MorningStart && hour < config.AfternoonStart {
		timeLabel = "morning"
	} else if hour >= config.AfternoonStart && hour < config.EveningStart {
		timeLabel = "afternoon"
	} else if hour >= config.EveningStart && hour < config.NightStart {
		timeLabel = "evening"
	} else {
		timeLabel = "night"
	}

	// Add weekend/weekday context
	dayType := "weekday"
	if weekday == time.Saturday || weekday == time.Sunday {
		dayType = "weekend"
	}

	return fmt.Sprintf("%s %s session", dayType, timeLabel)
}

// classifyPlatformDevice describes platform and device information
func classifyPlatformDevice(session *SessionProfile) string {
	var parts []string

	if session.Platform != "" && session.Platform != "unknown" {
		parts = append(parts, session.Platform)
	}

	if session.DeviceModel != "" && session.DeviceModel != "unknown" {
		parts = append(parts, session.DeviceModel)
	}

	if len(parts) == 0 {
		return "unknown platform"
	}

	return strings.Join(parts, " ")
}

// getCountryName converts ISO country code to country name
func getCountryName(code string) string {
	countryNames := map[string]string{
		"US": "United States",
		"GB": "United Kingdom",
		"CA": "Canada",
		"AU": "Australia",
		"DE": "Germany",
		"FR": "France",
		"IT": "Italy",
		"ES": "Spain",
		"NL": "Netherlands",
		"BR": "Brazil",
		"MX": "Mexico",
		"AR": "Argentina",
		"CO": "Colombia",
		"CL": "Chile",
		"PE": "Peru",
		"IN": "India",
		"CN": "China",
		"JP": "Japan",
		"KR": "South Korea",
		"RU": "Russia",
		"TR": "Turkey",
		"SA": "Saudi Arabia",
		"AE": "United Arab Emirates",
		"EG": "Egypt",
		"ZA": "South Africa",
		"NG": "Nigeria",
		"KE": "Kenya",
		"PL": "Poland",
		"SE": "Sweden",
		"NO": "Norway",
		"DK": "Denmark",
		"FI": "Finland",
		"PT": "Portugal",
		"GR": "Greece",
		"BE": "Belgium",
		"AT": "Austria",
		"CH": "Switzerland",
		"IE": "Ireland",
		"NZ": "New Zealand",
		"SG": "Singapore",
		"MY": "Malaysia",
		"TH": "Thailand",
		"ID": "Indonesia",
		"PH": "Philippines",
		"VN": "Vietnam",
		"PK": "Pakistan",
		"BD": "Bangladesh",
		"IL": "Israel",
		"UA": "Ukraine",
		"RO": "Romania",
		"CZ": "Czech Republic",
		"HU": "Hungary",
	}

	if name, exists := countryNames[code]; exists {
		return name
	}

	// If country code not in map, return the code itself
	return code
}
