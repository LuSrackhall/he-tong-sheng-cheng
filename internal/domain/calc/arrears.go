package calc

import (
	"math"
	"time"
)

const (
	Level1Warning   = 1 // 应缴预警 — used up date within 25 days ahead
	Level2Reminder  = 2 // 近期应缴提醒 — used up date within 5 days ahead
	Level3Overdue   = 3 // 逾期未缴催收 — used up date in the past, end date still future
	Level4Expiring  = 4 // 到期预警 — end date within 20 days, not fully paid
	Level5Recovery  = 5 // 已到期欠费追缴 — end date in the past, not fully paid
)

var LevelNames = map[int]string{
	Level1Warning:  "应缴预警",
	Level2Reminder: "近期应缴提醒",
	Level3Overdue:  "逾期未缴催收",
	Level4Expiring: "到期预警",
	Level5Recovery: "已到期欠费追缴",
}

var SuggestedActions = map[int]string{
	Level1Warning:  "列入观察，心中有数",
	Level2Reminder: "主动联系，提醒缴纳",
	Level3Overdue:  "上门催收，限期缴纳",
	Level4Expiring: "即将到期，清算账款",
	Level5Recovery: "进入追讨，法律途径",
}

// ClassifyArrears determines the arrears level for a contract.
// Returns the highest priority level (lower number = higher priority).
// Priority order: Level 3 > Level 2 > Level 1 > Level 5 > Level 4
func ClassifyArrears(usedUpDate, endDate time.Time, totalReceived, totalReceivable float64, today time.Time) int {
	notFullyPaid := totalReceived < totalReceivable

	// Level 3: overdue — used up date has passed, end date still in future
	if usedUpDate.Before(today) && endDate.After(today) && notFullyPaid {
		return Level3Overdue
	}

	// Level 2: imminent — used up date within 5 days
	daysToUsedUp := math.Ceil(usedUpDate.Sub(today).Hours() / 24)
	if daysToUsedUp >= 0 && daysToUsedUp <= 5 && notFullyPaid {
		return Level2Reminder
	}

	// Level 1: warning — used up date within 25 days
	if daysToUsedUp >= 0 && daysToUsedUp <= 25 && notFullyPaid {
		return Level1Warning
	}

	// Level 5: post-expiration debt — end date has passed, still owes
	daysPastEnd := math.Ceil(today.Sub(endDate).Hours() / 24)
	if daysPastEnd > 0 && notFullyPaid {
		return Level5Recovery
	}

	// Level 4: expiring — end date within 20 days
	daysToEnd := math.Ceil(endDate.Sub(today).Hours() / 24)
	if daysToEnd >= 0 && daysToEnd <= 20 && notFullyPaid {
		return Level4Expiring
	}

	return 0
}
