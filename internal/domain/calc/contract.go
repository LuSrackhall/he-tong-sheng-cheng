package calc

import (
	"math"
	"time"
)

const DaysPerMonth = 30.0

// TotalReceivable calculates the total rent receivable for the entire contract period.
// formula: wholeMonths * monthlyRent + remainingDays * (monthlyRent / 30)
func TotalReceivable(start, end time.Time, monthlyRent float64) float64 {
	days := int(end.Sub(start).Hours() / 24)
	if days <= 0 {
		return 0
	}
	wholeMonths := days / 30
	remainingDays := days % 30
	dailyRate := monthlyRent / DaysPerMonth
	return float64(wholeMonths)*monthlyRent + float64(remainingDays)*dailyRate
}

// UsedUpDate calculates the date through which the tenant has paid.
// Converts totalReceived into whole calendar months first, then remainder days.
// Each calendar month advances the date by one actual month (respecting varying month lengths).
func UsedUpDate(start time.Time, totalReceived, monthlyRent float64) time.Time {
	if monthlyRent <= 0 || totalReceived <= 0 {
		return start
	}
	dailyRate := monthlyRent / DaysPerMonth

	wholeMonths := int(totalReceived / monthlyRent)
	remainder := totalReceived - float64(wholeMonths)*monthlyRent

	date := start
	for i := 0; i < wholeMonths; i++ {
		date = date.AddDate(0, 1, 0)
	}

	if remainder > 0 {
		extraDays := int(math.Ceil(remainder / dailyRate))
		date = date.AddDate(0, 0, extraDays)
	}

	return date
}

// Arrears returns the amount still owed.
func Arrears(totalReceivable, totalReceived float64) float64 {
	if totalReceived >= totalReceivable {
		return 0
	}
	return totalReceivable - totalReceived
}
