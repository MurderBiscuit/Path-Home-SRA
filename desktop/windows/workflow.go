package main

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Shared workflow validation: fail closed when a form could create a false legal or money conclusion.
func validateMoneyField(raw string) error {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "$") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "$"))
	}
	if strings.Contains(raw, "$") {
		return errors.New("enter a valid dollar amount")
	}
	if raw == "" {
		return nil
	} // Unknown amount is not a free fee.
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 10000000 {
		return errors.New("enter a nonnegative dollar amount")
	}
	return nil
}
func dateOrderValid(start, end string) bool {
	if start != "" {
		if _, err := time.Parse("2006-01-02", start); err != nil {
			return false
		}
	}
	if end != "" {
		if _, err := time.Parse("2006-01-02", end); err != nil {
			return false
		}
	}
	return start == "" || end == "" || end >= start
}
func knownJurisdiction(code string, known []struct {
	Code string `json:"jurisdiction_code"`
	Name string `json:"jurisdiction_name"`
}) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return true
	}
	if code == "FED" {
		return false
	}
	for _, v := range known {
		if v.Code == code {
			return true
		}
	}
	return false
}

// Prefer the specific issue over ambiguous words like "move" so a change of address is not a travel rule.
func questionTopic(q string) string {
	q = strings.ToLower(q)
	topics := make([]string, 0, 4)
	containsAny := func(words ...string) bool {
		for _, v := range words {
			if strings.Contains(q, v) {
				return true
			}
		}
		return false
	}
	if containsAny("vote", "voting", "election") {
		topics = append(topics, "voting")
	}
	if containsAny("address", "housing", "residence", "reside", "live with", "move in", "moving in") {
		topics = append(topics, "housing")
	}
	if containsAny("report", "check in", "officer", "appointment") {
		topics = append(topics, "reporting")
	}
	if containsAny("travel", "county", "out of state", "leave the state", "move to another state") {
		topics = append(topics, "travel")
	}
	if regexp.MustCompile(`\b(work|working|job|jobs|employ|employed|employment)\b`).MatchString(q) {
		topics = append(topics, "employment")
	}
	if len(topics) > 1 {
		return "multiple"
	}
	if len(topics) == 1 {
		return topics[0]
	}
	return ""
}

func federalSupervision(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "federal" || v == "federal supervised release"
}
