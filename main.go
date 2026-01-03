package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Config represents the application configuration
type Config struct {
	TargetShiftDurationHours float64  `json:"target_shift_duration_hours"`
	TargetShiftRoles         []string `json:"target_shift_roles"`
	GmailLabel               string   `json:"gmail_label"`
	NotificationMethod       string   `json:"notification_method"`
}

// ShiftInfo represents parsed shift information
type ShiftInfo struct {
	Role          string
	DurationHours float64
	StartTime     string
	EndTime       string
}

const (
	credentialsFile = "credentials.json"
	tokenFile       = "token.json"
	configFile      = "config.json"
)

func main() {
	fmt.Println("Deputy Shift Claimer")
	fmt.Println(strings.Repeat("=", 60))

	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("Target shift duration: >=%.1f hours\n", config.TargetShiftDurationHours)
	fmt.Printf("Target roles: %s\n", strings.Join(config.TargetShiftRoles, ", "))
	fmt.Printf("Gmail label: %s\n", config.GmailLabel)
	fmt.Println(strings.Repeat("=", 60))

	// Authenticate with Gmail
	fmt.Println("\nAuthenticating with Gmail...")
	srv, err := authenticateGmail()
	if err != nil {
		log.Fatalf("Failed to authenticate with Gmail: %v", err)
	}
	fmt.Println("✓ Successfully authenticated")

	// Fetch Deputy emails
	fmt.Printf("\nFetching emails with label '%s'...\n", config.GmailLabel)
	messages, err := getDeputyEmails(srv, config.GmailLabel)
	if err != nil {
		log.Fatalf("Error fetching emails: %v", err)
	}
	fmt.Printf("✓ Found %d email(s)\n", len(messages))

	if len(messages) == 0 {
		return
	}

	// Process each email
	fmt.Println("\nProcessing emails...")
	matchesFound := 0

	for i, message := range messages {
		subject := getEmailHeader(message, "Subject")
		date := getEmailHeader(message, "Date")

		fmt.Printf("[%d/%d] %s...\n", i+1, len(messages), truncate(subject, 50))

		// Decode email body
		body := decodeEmailBody(message.Payload)

		// Extract shift information
		shiftInfo := extractShiftInfo(body, subject)

		if shiftInfo != nil && (shiftInfo.Role != "" || shiftInfo.DurationHours > 0) {
			// Check if it meets criteria
			meetsCriteria, reason := checkCriteria(*shiftInfo, config)

			if meetsCriteria {
				matchesFound++
				notifyMatch(subject, date, *shiftInfo, reason, config.NotificationMethod)
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Processing complete!")
	fmt.Printf("Total emails processed: %d\n", len(messages))
	fmt.Printf("Matching shifts found: %d\n", matchesFound)
	fmt.Println(strings.Repeat("=", 60))
}

// loadConfig loads configuration from JSON file
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config JSON: %w", err)
	}

	return &config, nil
}

// authenticateGmail authenticates with Gmail API
func authenticateGmail() (*gmail.Service, error) {
	ctx := context.Background()

	// Read credentials file
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w\nPlease download OAuth credentials from Google Cloud Console", err)
	}

	// Parse credentials
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	// Get token
	client := getClient(ctx, config)

	// Create Gmail service
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %w", err)
	}

	return srv, nil
}

// getClient gets an OAuth2 HTTP client
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(tokenFile, token)
	}
	return config.Client(ctx, token)
}

// getTokenFromWeb requests a token from the web
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)
	fmt.Print("Enter authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// tokenFromFile retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// getDeputyEmails fetches emails with the specified label
func getDeputyEmails(srv *gmail.Service, labelName string) ([]*gmail.Message, error) {
	user := "me"

	// Get label ID
	labels, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve labels: %w", err)
	}

	var labelID string
	for _, label := range labels.Labels {
		if label.Name == labelName {
			labelID = label.Id
			break
		}
	}

	if labelID == "" {
		return nil, fmt.Errorf("label '%s' not found in Gmail", labelName)
	}

	// Fetch messages with the label
	msgList, err := srv.Users.Messages.List(user).LabelIds(labelID).MaxResults(50).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %w", err)
	}

	if len(msgList.Messages) == 0 {
		return []*gmail.Message{}, nil
	}

	// Fetch full message details
	var messages []*gmail.Message
	for _, msg := range msgList.Messages {
		fullMsg, err := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("Warning: Unable to retrieve message %s: %v", msg.Id, err)
			continue
		}
		messages = append(messages, fullMsg)
	}

	return messages, nil
}

// decodeEmailBody decodes the email body from the payload
func decodeEmailBody(payload *gmail.MessagePart) string {
	var body string

	// Check for parts (multipart message)
	if len(payload.Parts) > 0 {
		for _, part := range payload.Parts {
			if part.MimeType == "text/plain" && part.Body.Data != "" {
				decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					body = string(decoded)
					break
				}
			} else if part.MimeType == "text/html" && body == "" && part.Body.Data != "" {
				decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					body = string(decoded)
				}
			}
		}
	} else if payload.Body.Data != "" {
		// Simple message
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			body = string(decoded)
		}
	}

	return body
}

// getEmailHeader extracts a header value from the message
func getEmailHeader(message *gmail.Message, headerName string) string {
	for _, header := range message.Payload.Headers {
		if strings.EqualFold(header.Name, headerName) {
			return header.Value
		}
	}
	return ""
}

// extractShiftInfo extracts shift information from email content
func extractShiftInfo(body, subject string) *ShiftInfo {
	info := &ShiftInfo{}

	// Extract role
	rolePattern := regexp.MustCompile(`(?i)(?:Shift|Position|Role):\s*([A-Za-z\s]+?)(?:\s*-|\s*\n|$)`)
	if matches := rolePattern.FindStringSubmatch(body); len(matches) > 1 {
		info.Role = strings.TrimSpace(matches[1])
	} else if matches := rolePattern.FindStringSubmatch(subject); len(matches) > 1 {
		info.Role = strings.TrimSpace(matches[1])
	}

	// Extract time range
	timePattern := regexp.MustCompile(`(\d{1,2}:\d{2}\s*(?:AM|PM|am|pm)?)\s*[-–—to]+\s*(\d{1,2}:\d{2}\s*(?:AM|PM|am|pm)?)`)
	if matches := timePattern.FindStringSubmatch(body); len(matches) > 2 {
		info.StartTime = strings.TrimSpace(matches[1])
		info.EndTime = strings.TrimSpace(matches[2])

		// Calculate duration
		duration, err := calculateDuration(info.StartTime, info.EndTime)
		if err == nil {
			info.DurationHours = duration
		}
	}

	// Extract explicit duration
	durationPattern := regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(?:hours?|hrs?|h)\b`)
	if matches := durationPattern.FindStringSubmatch(body); len(matches) > 1 && info.DurationHours == 0 {
		duration, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			info.DurationHours = duration
		}
	}

	return info
}

// calculateDuration calculates duration in hours between start and end times
func calculateDuration(startTime, endTime string) (float64, error) {
	// Parse times
	layouts := []string{
		"3:04 PM",
		"3:04PM",
		"15:04",
		"3:04 pm",
		"3:04pm",
	}

	var start, end time.Time
	var err error

	for _, layout := range layouts {
		start, err = time.Parse(layout, startTime)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, fmt.Errorf("unable to parse start time: %w", err)
	}

	for _, layout := range layouts {
		end, err = time.Parse(layout, endTime)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, fmt.Errorf("unable to parse end time: %w", err)
	}

	// If end time is earlier than start time, assume next day
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}

	duration := end.Sub(start)
	return duration.Hours(), nil
}

// checkCriteria checks if shift meets target criteria
func checkCriteria(shift ShiftInfo, config *Config) (bool, string) {
	var reasons []string

	// Check duration
	if shift.DurationHours > 0 && shift.DurationHours >= config.TargetShiftDurationHours {
		reasons = append(reasons, fmt.Sprintf("Duration: %.1fh (target: >=%.1fh)",
			shift.DurationHours, config.TargetShiftDurationHours))
	}

	// Check role
	if shift.Role != "" {
		for _, targetRole := range config.TargetShiftRoles {
			if strings.Contains(strings.ToLower(shift.Role), strings.ToLower(targetRole)) {
				reasons = append(reasons, fmt.Sprintf("Role: %s (matches: %s)", shift.Role, targetRole))
				break
			}
		}
	}

	if len(reasons) > 0 {
		return true, strings.Join(reasons, "; ")
	}

	return false, ""
}

// notifyMatch sends a notification for a matching shift
func notifyMatch(subject, date string, shift ShiftInfo, reason, method string) {
	if method == "console" {
		fmt.Println()
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("🎯 SHIFT MATCH FOUND!")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("Email: %s\n", subject)
		fmt.Printf("Date: %s\n", date)
		fmt.Printf("Shift Role: %s\n", valueOrNA(shift.Role))
		fmt.Printf("Duration: %.1f hours\n", shift.DurationHours)
		fmt.Printf("Start Time: %s\n", valueOrNA(shift.StartTime))
		fmt.Printf("End Time: %s\n", valueOrNA(shift.EndTime))
		fmt.Printf("Match Reason: %s\n", reason)
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()
	}
}

// valueOrNA returns the value or "N/A" if empty
func valueOrNA(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
