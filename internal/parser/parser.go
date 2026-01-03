package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hotpup/deputy-shift-claimer/internal/gmail"
)

// Shift represents a Deputy shift
type Shift struct {
	Role          string
	Date          string
	StartTime     string
	EndTime       string
	DurationHours float64
}

// FilterConfig contains configuration for filtering shifts
type FilterConfig struct {
	MinDurationHours float64
	AllowedRoles     []string
}

// ParseShift parses a Deputy shift notification email
func ParseShift(msg *gmail.Message) (*Shift, error) {
	shift := &Shift{}

	// Parse role from subject or body
	shift.Role = extractRole(msg.Subject, msg.Body)
	if shift.Role == "" {
		return nil, fmt.Errorf("could not extract role from message")
	}

	// Parse date
	shift.Date = extractDate(msg.Body)

	// Parse times
	shift.StartTime, shift.EndTime = extractTimes(msg.Body)

	// Calculate duration
	if shift.StartTime != "" && shift.EndTime != "" {
		duration, err := calculateDuration(shift.StartTime, shift.EndTime)
		if err == nil {
			shift.DurationHours = duration
		}
	}

	return shift, nil
}

// extractRole extracts the role from the message
func extractRole(subject, body string) string {
	// Look for common Deputy role patterns
	text := subject + " " + body

	// Check for specific roles
	roles := []string{
		"LG: ALL",
		"LG: North",
		"Deck Coordinator",
		"Lifeguard",
		"Pool Attendant",
		"Swim Instructor",
	}

	for _, role := range roles {
		if strings.Contains(text, role) {
			return role
		}
	}

	// Try to extract role with regex patterns
	// Pattern: "Role: <role name>" or "Position: <role name>"
	rolePatterns := []string{
		`(?i)role:\s*([^\n\r,]+)`,
		`(?i)position:\s*([^\n\r,]+)`,
		`(?i)area:\s*([^\n\r,]+)`,
	}

	for _, pattern := range rolePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractDate extracts the date from the message body
func extractDate(body string) string {
	// Common date patterns in Deputy emails
	datePatterns := []string{
		`\d{1,2}/\d{1,2}/\d{4}`,           // MM/DD/YYYY or DD/MM/YYYY
		`\d{4}-\d{2}-\d{2}`,               // YYYY-MM-DD
		`(?i)(Mon|Tue|Wed|Thu|Fri|Sat|Sun)[a-z]*,?\s+\d{1,2}\s+(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{4}`,
	}

	for _, pattern := range datePatterns {
		re := regexp.MustCompile(pattern)
		match := re.FindString(body)
		if match != "" {
			return match
		}
	}

	return ""
}

// extractTimes extracts start and end times from the message body
func extractTimes(body string) (string, string) {
	// Look for time patterns like "9:00 AM - 5:00 PM" or "09:00-17:00"
	timePatterns := []string{
		`(\d{1,2}:\d{2}\s*(?:AM|PM|am|pm)?)\s*[-–]\s*(\d{1,2}:\d{2}\s*(?:AM|PM|am|pm)?)`,
		`(\d{1,2}:\d{2})\s*[-–]\s*(\d{1,2}:\d{2})`,
	}

	for _, pattern := range timePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(body)
		if len(matches) > 2 {
			return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
		}
	}

	return "", ""
}

// calculateDuration calculates the duration in hours between start and end times
func calculateDuration(startTime, endTime string) (float64, error) {
	// Parse time formats
	formats := []string{
		"3:04 PM",
		"3:04PM",
		"15:04",
		"3:04 pm",
		"3:04pm",
	}

	var start, end time.Time
	var err error

	// Try to parse start time
	for _, format := range formats {
		start, err = time.Parse(format, startTime)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, fmt.Errorf("failed to parse start time: %w", err)
	}

	// Try to parse end time
	for _, format := range formats {
		end, err = time.Parse(format, endTime)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, fmt.Errorf("failed to parse end time: %w", err)
	}

	// If end time is before start time, assume it's the next day
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}

	duration := end.Sub(start)
	hours := duration.Hours()

	return hours, nil
}

// ShouldNotify determines if a shift should trigger a notification
func ShouldNotify(shift *Shift, config FilterConfig) bool {
	// Check if role matches any of the allowed roles (case-insensitive)
	roleMatches := false
	for _, allowedRole := range config.AllowedRoles {
		if strings.EqualFold(shift.Role, allowedRole) {
			roleMatches = true
			break
		}
	}

	// Check if duration meets the minimum
	durationMatches := shift.DurationHours >= config.MinDurationHours

	// Notify if either condition is met
	return roleMatches || durationMatches
}
