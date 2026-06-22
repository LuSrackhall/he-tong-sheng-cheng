package calc

import "testing"

func TestClassifyArrears(t *testing.T) {
	today := parseDate("2026-06-01")

	tests := []struct {
		name            string
		usedUpDate      string
		endDate         string
		totalReceived   float64
		totalReceivable float64
		expectedLevel   int
	}{
		{
			name:            "Level 1: warning 20 days ahead",
			usedUpDate:      "2026-06-21",
			endDate:         "2026-12-31",
			totalReceived:   6000,
			totalReceivable: 12000,
			expectedLevel:   Level1Warning,
		},
		{
			name:            "Level 1: warning 28 days ahead (within 30-day window)",
			usedUpDate:      "2026-06-29",
			endDate:         "2026-12-31",
			totalReceived:   6000,
			totalReceivable: 12000,
			expectedLevel:   Level1Warning,
		},
		{
			name:            "Level 2: imminent reminder 3 days ahead",
			usedUpDate:      "2026-06-04",
			endDate:         "2026-12-31",
			totalReceived:   5000,
			totalReceivable: 12000,
			expectedLevel:   Level2Reminder,
		},
		{
			name:            "Level 3: overdue",
			usedUpDate:      "2026-05-15",
			endDate:         "2026-12-31",
			totalReceived:   3000,
			totalReceivable: 12000,
			expectedLevel:   Level3Overdue,
		},
		{
			name:            "Level 3: today equals endDate (boundary)",
			usedUpDate:      "2026-05-15",
			endDate:         "2026-06-01",
			totalReceived:   3000,
			totalReceivable: 12000,
			expectedLevel:   Level3Overdue,
		},
		{
			name:            "Level 4: expiring in 10 days",
			usedUpDate:      "2026-12-31",
			endDate:         "2026-06-11",
			totalReceived:   3000,
			totalReceivable: 12000,
			expectedLevel:   Level4Expiring,
		},
		{
			name:            "Level 5: post-expiration debt",
			usedUpDate:      "2026-03-01",
			endDate:         "2026-05-15",
			totalReceived:   3000,
			totalReceivable: 12000,
			expectedLevel:   Level5Recovery,
		},
		{
			name:            "Level 3 takes priority over Level 4",
			usedUpDate:      "2026-05-15",
			endDate:         "2026-06-15",
			totalReceived:   3000,
			totalReceivable: 12000,
			expectedLevel:   Level3Overdue,
		},
		{
			name:            "Fully paid — no level",
			usedUpDate:      "2026-05-15",
			endDate:         "2026-12-31",
			totalReceived:   12000,
			totalReceivable: 12000,
			expectedLevel:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyArrears(parseDate(tt.usedUpDate), parseDate(tt.endDate), tt.totalReceived, tt.totalReceivable, today)
			if got != tt.expectedLevel {
				t.Errorf("ClassifyArrears() = %d, want %d", got, tt.expectedLevel)
			}
		})
	}
}

func TestSuggestedActionsExist(t *testing.T) {
	for level := 1; level <= 5; level++ {
		if name, ok := LevelNames[level]; !ok || name == "" {
			t.Errorf("Level %d missing name", level)
		}
		if action, ok := SuggestedActions[level]; !ok || action == "" {
			t.Errorf("Level %d missing suggested action", level)
		}
	}
}
