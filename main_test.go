package main

import (
	"strings"
	"testing"
)

func TestExtractShiftInfo(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		subject       string
		wantRole      string
		wantDuration  float64
		wantStartTime string
		wantEndTime   string
	}{
		{
			name: "Extract role from body",
			body: `
Hello,

A shift is available!
Shift: Bartender
Date: January 5, 2026
`,
			subject:  "Deputy Shift Available",
			wantRole: "Bartender",
		},
		{
			name: "Extract time range",
			body: `
Shift Details:
Position: Server
Time: 9:00 AM - 5:00 PM
`,
			subject:       "New Shift",
			wantRole:      "Server",
			wantDuration:  8.0,
			wantStartTime: "9:00 AM",
			wantEndTime:   "5:00 PM",
		},
		{
			name: "Extract explicit duration",
			body: `
Available shift
Role: Manager
Duration: 10 hours
`,
			subject:      "Shift Available",
			wantRole:     "Manager",
			wantDuration: 10.0,
		},
		{
			name:     "Extract role from subject",
			body:     "A shift is available for pickup.",
			subject:  "Shift: Server - Available Now",
			wantRole: "Server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := extractShiftInfo(tt.body, tt.subject)

			if info == nil {
				t.Fatal("extractShiftInfo returned nil")
			}

			if tt.wantRole != "" && !strings.Contains(info.Role, tt.wantRole) {
				t.Errorf("Role = %v, want %v", info.Role, tt.wantRole)
			}

			if tt.wantDuration > 0 && info.DurationHours != tt.wantDuration {
				t.Errorf("DurationHours = %v, want %v", info.DurationHours, tt.wantDuration)
			}

			if tt.wantStartTime != "" && info.StartTime != tt.wantStartTime {
				t.Errorf("StartTime = %v, want %v", info.StartTime, tt.wantStartTime)
			}

			if tt.wantEndTime != "" && info.EndTime != tt.wantEndTime {
				t.Errorf("EndTime = %v, want %v", info.EndTime, tt.wantEndTime)
			}
		})
	}
}

func TestCalculateDuration(t *testing.T) {
	tests := []struct {
		name      string
		startTime string
		endTime   string
		want      float64
		wantErr   bool
	}{
		{
			name:      "Standard 8-hour shift",
			startTime: "9:00 AM",
			endTime:   "5:00 PM",
			want:      8.0,
		},
		{
			name:      "Overnight shift",
			startTime: "10:00 PM",
			endTime:   "6:00 AM",
			want:      8.0,
		},
		{
			name:      "Short shift",
			startTime: "2:00 PM",
			endTime:   "6:00 PM",
			want:      4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateDuration(tt.startTime, tt.endTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCriteria(t *testing.T) {
	config := &Config{
		TargetShiftDurationHours: 8.0,
		TargetShiftRoles:         []string{"Bartender", "Server"},
	}

	tests := []struct {
		name        string
		shift       ShiftInfo
		wantMatch   bool
		wantContain string
	}{
		{
			name: "Duration and role match",
			shift: ShiftInfo{
				Role:          "Server",
				DurationHours: 8.0,
			},
			wantMatch:   true,
			wantContain: "Duration",
		},
		{
			name: "Role match only",
			shift: ShiftInfo{
				Role:          "Bartender",
				DurationHours: 4.0,
			},
			wantMatch:   true,
			wantContain: "Bartender",
		},
		{
			name: "Duration match only",
			shift: ShiftInfo{
				Role:          "Cook",
				DurationHours: 10.0,
			},
			wantMatch:   true,
			wantContain: "Duration",
		},
		{
			name: "No match",
			shift: ShiftInfo{
				Role:          "Cook",
				DurationHours: 4.0,
			},
			wantMatch: false,
		},
		{
			name: "Partial role match",
			shift: ShiftInfo{
				Role:          "Head Bartender",
				DurationHours: 6.0,
			},
			wantMatch:   true,
			wantContain: "Bartender",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, reason := checkCriteria(tt.shift, config)

			if match != tt.wantMatch {
				t.Errorf("checkCriteria() match = %v, want %v", match, tt.wantMatch)
			}

			if tt.wantMatch && tt.wantContain != "" && !strings.Contains(reason, tt.wantContain) {
				t.Errorf("checkCriteria() reason = %v, want to contain %v", reason, tt.wantContain)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "Short string",
			input:  "Hello",
			maxLen: 10,
			want:   "Hello",
		},
		{
			name:   "Exact length",
			input:  "HelloWorld",
			maxLen: 10,
			want:   "HelloWorld",
		},
		{
			name:   "Long string",
			input:  "This is a very long string",
			maxLen: 10,
			want:   "This is a ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueOrNA(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Non-empty value",
			input: "Server",
			want:  "Server",
		},
		{
			name:  "Empty value",
			input: "",
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueOrNA(tt.input)
			if got != tt.want {
				t.Errorf("valueOrNA() = %v, want %v", got, tt.want)
			}
		})
	}
}
