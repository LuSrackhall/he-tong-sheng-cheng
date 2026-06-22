package calc

import "testing"

func TestContractStatus(t *testing.T) {
	today := parseDate("2026-06-01")

	tests := []struct {
		name            string
		endDate         string
		totalReceived   float64
		totalReceivable float64
		expected        string
	}{
		{"fully paid", "2026-12-31", 12000, 12000, StatusPaidUp},
		{"active with partial payment", "2026-12-31", 3000, 12000, StatusArrears},
		{"expired with debt", "2026-05-15", 3000, 12000, StatusExpired},
		{"expired and fully paid → paidup", "2026-05-15", 12000, 12000, StatusPaidUp},
		{"no payment yet → active", "2026-12-31", 0, 12000, StatusActive},
		{"at end date with partial payment", "2026-06-01", 3000, 12000, StatusArrears},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContractStatus(parseDate(tt.endDate), tt.totalReceived, tt.totalReceivable, today)
			if got != tt.expected {
				t.Errorf("ContractStatus() = %s, want %s", got, tt.expected)
			}
		})
	}
}
