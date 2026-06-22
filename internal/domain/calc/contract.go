package calc

import (
	"math"
	"time"
)

const DaysPerMonth = 30.0

// addMonths advances d by n calendar months, clamping end-of-month dates
// (e.g. Jan 31 → Feb 28 instead of Mar 3).
func addMonths(d time.Time, n int) time.Time {
	targetMonth := d.Month() + time.Month(n)
	year := d.Year() + int(targetMonth-1)/12
	month := time.Month((int(targetMonth)-1)%12 + 1)
	day := d.Day()
	lastDay := daysIn(year, month)
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, d.Hour(), d.Minute(), d.Second(), d.Nanosecond(), d.Location())
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// TotalReceivable calculates the total rent receivable for the entire contract period.
// Uses real calendar months (month-end aware), then remaining days at daily rate.
func TotalReceivable(start, end time.Time, monthlyRent float64) float64 {
	if !end.After(start) {
		return 0
	}
	wholeMonths := 0
	cursor := start
	for {
		next := addMonths(cursor, 1)
		if next.After(end) {
			break
		}
		wholeMonths++
		cursor = next
	}
	remainingDays := int(end.Sub(cursor).Hours() / 24)
	dailyRate := monthlyRent / DaysPerMonth
	return float64(wholeMonths)*monthlyRent + float64(remainingDays)*dailyRate
}

// UsedUpDate calculates the date through which the tenant has paid.
// Converts totalReceived into whole calendar months first, then remainder days.
// Each calendar month advances the date by one actual month (respecting varying month lengths).
// The result is capped at endDate.
func UsedUpDate(start time.Time, totalReceived, monthlyRent float64, endDate time.Time) time.Time {
	if monthlyRent <= 0 || totalReceived <= 0 {
		return start
	}
	dailyRate := monthlyRent / DaysPerMonth

	wholeMonths := int(totalReceived / monthlyRent)
	remainder := totalReceived - float64(wholeMonths)*monthlyRent

	date := start
	for i := 0; i < wholeMonths; i++ {
		date = addMonths(date, 1)
	}

	if remainder > 0 {
		extraDays := int(math.Ceil(remainder / dailyRate))
		date = date.AddDate(0, 0, extraDays)
	}

	if date.After(endDate) {
		return endDate
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
