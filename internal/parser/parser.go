package main

import (
	"regexp"
	"strings"
	"time"
)

type Shift struct {
	Role      string
	StartTime time.Time
	EndTime   time.Time
}

var (
	timeRegex = regexp.MustCompile(
		`(?i)(\d{1,2}:\d{2}\s*(AM|PM))\s*-\s*(\d{1,2}:\d{2}\s*(AM|PM))`,
	)
	roleRegex = regexp.MustCompile(`(?i)\b(LG|Deck Coordinator)\b`)
	dateRegex = regexp.MustCompile(
		`(?i)(Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday),\s+(\d{1,2}\s+\w+)`,
	)
)

func parseShift(body string) (*Shift, bool) {
	body = strings.ReplaceAll(body, "\n", " ")

	roleMatch := roleRegex.FindString(body)
	timeMatch := timeRegex.FindStringSubmatch(body)
	dateMatch := dateRegex.FindStringSubmatch(body)

	if roleMatch == "" || len(timeMatch) < 5 || len(dateMatch) < 3 {
		return nil, false
	}

	dateStr := dateMatch[2] + " " + time.Now().Format("2006")
	layout := "2 January 2006 3:04 PM"

	start, err1 := time.Parse(layout, dateStr+" "+timeMatch[1])
	end, err2 := time.Parse(layout, dateStr+" "+timeMatch[3])

	if err1 != nil || err2 != nil {
		return nil, false
	}

	return &Shift{
		Role:      roleMatch,
		StartTime: start,
		EndTime:   end,
	}, true
}
