package calc

import "time"

const (
	StatusActive  = "active"
	StatusPaidUp  = "paidup"
	StatusArrears = "arrears"
	StatusExpired = "expired"
)

// ContractStatus determines the contract status based on payment state and dates.
func ContractStatus(endDate time.Time, totalReceived, totalReceivable float64, today time.Time) string {
	if totalReceived >= totalReceivable {
		return StatusPaidUp
	}

	if today.After(endDate) {
		return StatusExpired
	}

	if totalReceived > 0 && totalReceived < totalReceivable {
		return StatusArrears
	}

	return StatusActive
}
