package parser

import (
	"testing"

	"github.com/hotpup/deputy-shift-claimer/internal/gmail"
)

func TestParseShift(t *testing.T) {
	tests := []struct {
		name        string
		message     *gmail.Message
		expectRole  string
		expectError bool
	}{
		{
			name: "LG: ALL role in subject",
			message: &gmail.Message{
				ID:      "msg1",
				Subject: "New Shift Available - LG: ALL",
				Body:    "Shift available on 01/15/2026 from 9:00 AM - 5:00 PM",
			},
			expectRole:  "LG: ALL",
			expectError: false,
		},
		{
			name: "LG: North role in body",
			message: &gmail.Message{
				ID:      "msg2",
				Subject: "New Shift Available",
				Body:    "Role: LG: North\nDate: 01/15/2026\nTime: 9:00 AM - 5:00 PM",
			},
			expectRole:  "LG: North",
			expectError: false,
		},
		{
			name: "Deck Coordinator role",
			message: &gmail.Message{
				ID:      "msg3",
				Subject: "Deck Coordinator Shift",
				Body:    "Available on 01/15/2026 from 10:00 AM - 6:00 PM",
			},
			expectRole:  "Deck Coordinator",
			expectError: false,
		},
		{
			name: "No role found",
			message: &gmail.Message{
				ID:      "msg4",
				Subject: "Some notification",
				Body:    "No role information here",
			},
			expectRole:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shift, err := ParseShift(tt.message)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if shift.Role != tt.expectRole {
				t.Errorf("expected role %q, got %q", tt.expectRole, shift.Role)
			}
		})
	}
}

func TestExtractTimes(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "AM/PM format",
			body:      "Shift from 9:00 AM - 5:00 PM today",
			wantStart: "9:00 AM",
			wantEnd:   "5:00 PM",
		},
		{
			name:      "24-hour format",
			body:      "Time: 09:00-17:00",
			wantStart: "09:00",
			wantEnd:   "17:00",
		},
		{
			name:      "No times",
			body:      "No time information",
			wantStart: "",
			wantEnd:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := extractTimes(tt.body)
			if start != tt.wantStart {
				t.Errorf("extractTimes() start = %v, want %v", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("extractTimes() end = %v, want %v", end, tt.wantEnd)
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
			name:      "8-hour shift",
			startTime: "9:00 AM",
			endTime:   "5:00 PM",
			want:      8.0,
			wantErr:   false,
		},
		{
			name:      "4-hour shift",
			startTime: "10:00 AM",
			endTime:   "2:00 PM",
			want:      4.0,
			wantErr:   false,
		},
		{
			name:      "overnight shift",
			startTime: "10:00 PM",
			endTime:   "6:00 AM",
			want:      8.0,
			wantErr:   false,
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

func TestShouldNotify(t *testing.T) {
	config := FilterConfig{
		MinDurationHours: 4.0,
		AllowedRoles: []string{
			"LG: ALL",
			"LG: North",
			"Deck Coordinator",
		},
	}

	tests := []struct {
		name  string
		shift *Shift
		want  bool
	}{
		{
			name: "matches allowed role",
			shift: &Shift{
				Role:          "LG: ALL",
				DurationHours: 2.0,
			},
			want: true,
		},
		{
			name: "matches duration threshold",
			shift: &Shift{
				Role:          "Other Role",
				DurationHours: 5.0,
			},
			want: true,
		},
		{
			name: "matches both conditions",
			shift: &Shift{
				Role:          "LG: North",
				DurationHours: 6.0,
			},
			want: true,
		},
		{
			name: "matches neither condition",
			shift: &Shift{
				Role:          "Other Role",
				DurationHours: 2.0,
			},
			want: false,
		},
		{
			name: "exact duration threshold",
			shift: &Shift{
				Role:          "Other Role",
				DurationHours: 4.0,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldNotify(tt.shift, config); got != tt.want {
				t.Errorf("ShouldNotify() = %v, want %v", got, tt.want)
			}
		})
	}
}
