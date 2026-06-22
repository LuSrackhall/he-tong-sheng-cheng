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
		name         string
		start, end   string
		monthlyRent  float64
		expected     float64
	}{
		{"3 full months", "2026-01-01", "2026-04-01", 1000, 3000},
		{"15 days", "2026-01-01", "2026-01-16", 1000, 500},
		{"30 days", "2026-01-01", "2026-01-31", 1000, 1000},
		{"0 days", "2026-01-01", "2026-01-01", 1000, 0},
		{"full year", "2026-01-01", "2026-12-31", 1000, 12000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TotalReceivable(parseDate(tt.start), parseDate(tt.end), tt.monthlyRent)
			if !floatEqual(got, tt.expected) {
				t.Errorf("TotalReceivable(%s, %s, %.0f) = %.2f, want %.2f", tt.start, tt.end, tt.monthlyRent, got, tt.expected)
			}
		})
	}
}

func TestTotalReceivable_MonthEndStart(t *testing.T) {
	// 1月31日起租到4月30日
	// addMonths(Jan31,1)=Feb28, addMonths(Feb28,1)=Mar28, addMonths(Mar28,1)=Apr28
	// 3 whole months + 2 days remaining
	start := parseDate("2026-01-31")
	end := parseDate("2026-04-30")
	got := TotalReceivable(start, end, 1000)
	want := 3*1000.0 + 2*(1000.0/30.0)
	if !floatEqual(got, want) {
		t.Errorf("TotalReceivable(月末起租) = %.2f, want %.2f", got, want)
	}
}

func TestUsedUpDate(t *testing.T) {
	endDate := parseDate("2026-12-31")
	tests := []struct {
		name          string
		start         string
		totalReceived float64
		monthlyRent   float64
		expectedDate  string
	}{
		{"3 full months", "2026-01-01", 3000, 1000, "2026-04-01"},
		{"3 months + 15 days", "2026-01-01", 3500, 1000, "2026-04-16"},
		{"nothing paid", "2026-01-01", 0, 1000, "2026-01-01"},
		{"1 month", "2026-01-01", 1000, 1000, "2026-02-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UsedUpDate(parseDate(tt.start), tt.totalReceived, tt.monthlyRent, endDate)
			expected := parseDate(tt.expectedDate)
			if !got.Equal(expected) {
				t.Errorf("UsedUpDate(%s, %.0f, %.0f) = %s, want %s", tt.start, tt.totalReceived, tt.monthlyRent, got.Format("2006-01-02"), tt.expectedDate)
			}
		})
	}
}

func TestUsedUpDate_CappedAtEndDate(t *testing.T) {
	start := parseDate("2026-01-01")
	endDate := parseDate("2026-06-30")
	got := UsedUpDate(start, 999999, 1000, endDate)
	if !got.Equal(endDate) {
		t.Errorf("UsedUpDate(超额) = %s, want %s (capped at endDate)", got.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}
}

func TestArrears(t *testing.T) {
	tests := []struct {
		totalReceivable, totalReceived, expected float64
	}{
		{12000, 6000, 6000},
		{12000, 12000, 0},
		{12000, 15000, 0},
	}

	for _, tt := range tests {
		got := Arrears(tt.totalReceivable, tt.totalReceived)
		if got != tt.expected {
			t.Errorf("Arrears(%.0f, %.0f) = %.0f, want %.0f", tt.totalReceivable, tt.totalReceived, got, tt.expected)
		}
	}
}
