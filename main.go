package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hotpup/deputy-shift-claimer/internal/gmail"
	"github.com/hotpup/deputy-shift-claimer/internal/parser"
)

func main() {
	// Initialize Gmail client
	gmailClient, err := gmail.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Gmail client: %v", err)
	}

	// Fetch emails from Deputy label
	messages, err := gmailClient.FetchDeputyEmails()
	if err != nil {
		log.Fatalf("Failed to fetch Deputy emails: %v", err)
	}

	// Parse and filter shifts
	config := parser.FilterConfig{
		MinDurationHours: getEnvAsFloat("MIN_DURATION_HOURS", 4.0),
		AllowedRoles: []string{
			"LG: ALL",
			"LG: North",
			"Deck Coordinator",
		},
	}

	for _, msg := range messages {
		shift, err := parser.ParseShift(msg)
		if err != nil {
			log.Printf("Failed to parse shift from message %s: %v", msg.ID, err)
			continue
		}

		if parser.ShouldNotify(shift, config) {
			fmt.Printf("NOTIFICATION: Shift available!\n")
			fmt.Printf("  Role: %s\n", shift.Role)
			fmt.Printf("  Duration: %.2f hours\n", shift.DurationHours)
			fmt.Printf("  Date: %s\n", shift.Date)
			fmt.Printf("  Time: %s - %s\n", shift.StartTime, shift.EndTime)
			fmt.Printf("  Message ID: %s\n", msg.ID)
			fmt.Println()
		}
	}
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	var val float64
	_, err := fmt.Sscanf(valStr, "%f", &val)
	if err != nil {
		return defaultVal
	}
	return val
}
