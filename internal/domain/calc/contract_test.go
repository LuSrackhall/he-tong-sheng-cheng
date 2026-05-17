package calc

import (
	"math"
	"testing"
	"time"
)

const epsilon = 0.01

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestTotalReceivable(t *testing.T) {
	tests := []struct {
		start, end string
		monthlyRent float64
		expected    float64
	}{
		{"2026-01-01", "2026-04-01", 1000, 3000},   // 3 full months = 90 days
		{"2026-01-01", "2026-01-16", 1000, 500},     // 15 days
		{"2026-01-01", "2026-01-31", 1000, 1000},    // 30 days = 1 month
		{"2026-01-01", "2026-01-01", 1000, 0},       // 0 days
	}

	for _, tt := range tests {
		got := TotalReceivable(parseDate(tt.start), parseDate(tt.end), tt.monthlyRent)
		if !floatEqual(got, tt.expected) {
			t.Errorf("TotalReceivable(%s, %s, %.0f) = %.2f, want %.2f", tt.start, tt.end, tt.monthlyRent, got, tt.expected)
		}
	}
}

func TestUsedUpDate(t *testing.T) {
	tests := []struct {
		start          string
		totalReceived  float64
		monthlyRent    float64
		expectedDate   string
	}{
		{"2026-01-01", 3000, 1000, "2026-04-01"},   // 3 full months
		{"2026-01-01", 3500, 1000, "2026-04-16"},   // 3 months + 15 days (ceil(500/33.33))
		{"2026-01-01", 0, 1000, "2026-01-01"},       // nothing paid
		{"2026-01-01", 1000, 1000, "2026-02-01"},    // 1 month
	}

	for _, tt := range tests {
		got := UsedUpDate(parseDate(tt.start), tt.totalReceived, tt.monthlyRent)
		expected := parseDate(tt.expectedDate)
		if !got.Equal(expected) {
			t.Errorf("UsedUpDate(%s, %.0f, %.0f) = %s, want %s", tt.start, tt.totalReceived, tt.monthlyRent, got.Format("2006-01-02"), tt.expectedDate)
		}
	}
}

func TestArrears(t *testing.T) {
	tests := []struct {
		totalReceivable, totalReceived, expected float64
	}{
		{12000, 6000, 6000},
		{12000, 12000, 0},
		{12000, 15000, 0},  // overpaid
	}

	for _, tt := range tests {
		got := Arrears(tt.totalReceivable, tt.totalReceived)
		if got != tt.expected {
			t.Errorf("Arrears(%.0f, %.0f) = %.0f, want %.0f", tt.totalReceivable, tt.totalReceived, got, tt.expected)
		}
	}
}
