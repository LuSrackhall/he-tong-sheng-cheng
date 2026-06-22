package calc

import "time"

const (
	StatusActive  = "active"
	StatusPaidUp  = "paidup"
	StatusArrears = "arrears"
	StatusExpired = "expired"
)

// ContractStatus determines the contract status based on payment state and dates.
// Priority: paidup > expired > arrears > active
func ContractStatus(endDate time.Time, totalReceived, totalReceivable float64, today time.Time) string {
	// 已缴清优先（不论是否过期，缴清合同不进催缴清单）
	if totalReceived >= totalReceivable {
		return StatusPaidUp
	}

	// 已过期（未缴清）
	if today.After(endDate) {
		return StatusExpired
	}

	// 有收款但未缴清
	if totalReceived > 0 {
		return StatusArrears
	}

	// 未收款，执行中
	return StatusActive
}
