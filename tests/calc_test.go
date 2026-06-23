package tests

import (
	"asset-leasing-system/internal/domain/calc"
	"math"
	"testing"
	"time"
)

// ── TotalReceivable 测试（使用真实日历月） ───────────────────────────────────

func TestTotalReceivable_ExactMonths(t *testing.T) {
	// 2026-01-01 到 2026-04-01 = 3个日历月
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	got := calc.TotalReceivable(start, end, 1000)
	want := 3000.0 // 3 * 1000
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(3个整月) = %f, want %f", got, want)
	}
}

func TestTotalReceivable_WithRemainingDays(t *testing.T) {
	// 2026-01-01 到 2026-04-05 = 3个日历月 + 4天
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	got := calc.TotalReceivable(start, end, 1000)
	want := 3000.0 + 4*(1000.0/30.0)
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(3月+4天) = %f, want %f", got, want)
	}
}

func TestTotalReceivable_FullYear(t *testing.T) {
	// 2026-01-01 到 2026-12-31
	// addMonths 循环: 11 whole months (Jan→Dec), cursor=Dec1, remaining=30 days
	// 11 * 1000 + 30 * (1000/30) = 12000
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := calc.TotalReceivable(start, end, 1000)
	want := 12000.0 // 11*1000 + 30*(1000/30)
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(全年) = %f, want %f", got, want)
	}
}

func TestTotalReceivable_ShortPeriod(t *testing.T) {
	// 15天，不到1个日历月
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)
	got := calc.TotalReceivable(start, end, 3000)
	// 0 * 3000 + 15 * (3000/30) = 1500
	want := 15.0 * (3000.0 / 30.0)
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(15天) = %f, want %f", got, want)
	}
}

func TestTotalReceivable_InvalidPeriod(t *testing.T) {
	start := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := calc.TotalReceivable(start, end, 1000)
	if got != 0 {
		t.Errorf("TotalReceivable(无效区间) = %f, want 0", got)
	}
}

// ── UsedUpDate 测试 ──────────────────────────────────────────────────────────

func TestUsedUpDate_ExactMonths(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := calc.UsedUpDate(start, 3000, 1000, endDate)
	// 3个日历月: Jan 1 → Feb 1 → Mar 1 → Apr 1
	want := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("UsedUpDate(3000) = %v, want %v", got, want)
	}
}

func TestUsedUpDate_WithPartialMonth(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	// 付了3500 = 3个月 + 500
	// 500 / (1000/30) = 15天
	got := calc.UsedUpDate(start, 3500, 1000, endDate)
	want := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC) // Apr 1 + 15 days
	if !got.Equal(want) {
		t.Errorf("UsedUpDate(3500) = %v, want %v", got, want)
	}
}

func TestUsedUpDate_ZeroReceived(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := calc.UsedUpDate(start, 0, 1000, endDate)
	if !got.Equal(start) {
		t.Errorf("UsedUpDate(0) = %v, want %v", got, start)
	}
}

func TestUsedUpDate_FullyPaid_CappedAtEndDate(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	totalReceivable := calc.TotalReceivable(start, endDate, 1000)
	got := calc.UsedUpDate(start, totalReceivable, 1000, endDate)
	// 全额付清，usedUpDate 应等于 endDate（被 cap）
	if got.After(endDate) {
		t.Errorf("UsedUpDate(全额) = %v, should not exceed endDate %v", got, endDate)
	}
}

func TestUsedUpDate_CappedAtEndDate(t *testing.T) {
	// 超额付款，usedUpDate 应被限制在 endDate
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	got := calc.UsedUpDate(start, 999999, 1000, endDate)
	if !got.Equal(endDate) {
		t.Errorf("UsedUpDate(超额) = %v, want %v (capped at endDate)", got, endDate)
	}
}

// ── Arrears 测试 ─────────────────────────────────────────────────────────────

func TestArrears_PartialPayment(t *testing.T) {
	got := calc.Arrears(12000, 5000)
	want := 7000.0
	if math.Abs(got-want) > 0.01 {
		t.Errorf("Arrears(12000, 5000) = %f, want %f", got, want)
	}
}

func TestArrears_FullyPaid(t *testing.T) {
	got := calc.Arrears(12000, 12000)
	if got != 0 {
		t.Errorf("Arrears(12000, 12000) = %f, want 0", got)
	}
}

func TestArrears_OverPaid(t *testing.T) {
	got := calc.Arrears(12000, 15000)
	if got != 0 {
		t.Errorf("Arrears(12000, 15000) = %f, want 0", got)
	}
}

// ── ContractStatus 测试 ──────────────────────────────────────────────────────

func TestContractStatus_PaidUp(t *testing.T) {
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 12000, 12000, today)
	if got != calc.StatusPaidUp {
		t.Errorf("ContractStatus(全额付清) = %s, want %s", got, calc.StatusPaidUp)
	}
}

func TestContractStatus_Expired(t *testing.T) {
	endDate := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 5000, 12000, today)
	if got != calc.StatusExpired {
		t.Errorf("ContractStatus(已过期) = %s, want %s", got, calc.StatusExpired)
	}
}

func TestContractStatus_Arrears(t *testing.T) {
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 5000, 12000, today)
	if got != calc.StatusArrears {
		t.Errorf("ContractStatus(部分付款) = %s, want %s", got, calc.StatusArrears)
	}
}

func TestContractStatus_Active(t *testing.T) {
	// 修复后：零收款且未过期 → active（执行中）
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 0, 12000, today)
	if got != calc.StatusActive {
		t.Errorf("ContractStatus(0 received) = %s, want %s", got, calc.StatusActive)
	}
}

func TestContractStatus_PaidUpOverridesExpired(t *testing.T) {
	// 修复后：缴清优先于过期 — 已缴清的合同即使过期也不应进入催缴清单
	endDate := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 12000, 12000, today)
	if got != calc.StatusPaidUp {
		t.Errorf("ContractStatus(已过期+全额) = %s, want %s", got, calc.StatusPaidUp)
	}
}

// ── ClassifyArrears 测试 ─────────────────────────────────────────────────────

func TestClassifyArrears_Level3_Overdue(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC) // past
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)  // future
	got := calc.ClassifyArrears(usedUpDate, endDate, 5000, 12000, today)
	if got != calc.Level3Overdue {
		t.Errorf("ClassifyArrears(逾期) = %d, want %d", got, calc.Level3Overdue)
	}
}

func TestClassifyArrears_Level2_Imminent(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC) // 3 days ahead
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := calc.ClassifyArrears(usedUpDate, endDate, 5000, 12000, today)
	if got != calc.Level2Reminder {
		t.Errorf("ClassifyArrears(3天内) = %d, want %d", got, calc.Level2Reminder)
	}
}

func TestClassifyArrears_Level1_Warning(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC) // 10 days ahead
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := calc.ClassifyArrears(usedUpDate, endDate, 5000, 12000, today)
	if got != calc.Level1Warning {
		t.Errorf("ClassifyArrears(10天后) = %d, want %d", got, calc.Level1Warning)
	}
}

func TestClassifyArrears_Level5_Recovery(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC) // past
	got := calc.ClassifyArrears(usedUpDate, endDate, 5000, 12000, today)
	if got != calc.Level5Recovery {
		t.Errorf("ClassifyArrears(已过期未付清) = %d, want %d", got, calc.Level5Recovery)
	}
}

func TestClassifyArrears_Level4_Expiring(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC) // far future
	endDate := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)    // 15 days ahead
	got := calc.ClassifyArrears(usedUpDate, endDate, 5000, 12000, today)
	if got != calc.Level4Expiring {
		t.Errorf("ClassifyArrears(15天后到期) = %d, want %d", got, calc.Level4Expiring)
	}
}

func TestClassifyArrears_FullyPaid(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := calc.ClassifyArrears(usedUpDate, endDate, 12000, 12000, today)
	if got != 0 {
		t.Errorf("ClassifyArrears(全额付清) = %d, want 0", got)
	}
}

func TestClassifyArrears_Priority_Level3OverLevel2(t *testing.T) {
	// usedUpDate in past (L3) overrides L2/L1 even if within 5-day window
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC) // 5 days ago
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := calc.ClassifyArrears(usedUpDate, endDate, 5000, 12000, today)
	if got != calc.Level3Overdue {
		t.Errorf("ClassifyArrears(优先级L3>L2) = %d, want %d", got, calc.Level3Overdue)
	}
}

func TestClassifyArrears_Priority_Level5OverLevel4(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	usedUpDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC) // 5 days ago
	got := calc.ClassifyArrears(usedUpDate, endDate, 5000, 12000, today)
	if got != calc.Level5Recovery {
		t.Errorf("ClassifyArrears(优先级L5>L4) = %d, want %d", got, calc.Level5Recovery)
	}
}

// ── 月末起租边界测试（addMonths 钳位行为） ───────────────────────────────────

func TestTotalReceivable_MonthEnd_Start(t *testing.T) {
	// 1月31日起租，到4月30日
	// addMonths(Jan31,1) = Feb28 (clamped), addMonths(Feb28,1) = Mar28, addMonths(Mar28,1) = Apr28
	// 3 whole months, remaining: Apr28→Apr30 = 2 days
	start := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	got := calc.TotalReceivable(start, end, 1000)
	want := 3*1000.0 + 2*(1000.0/30.0)
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(月末起租) = %f, want %f", got, want)
	}
}

func TestUsedUpDate_MonthEnd_Start(t *testing.T) {
	start := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	// 付了2个月 = 2000
	// addMonths(Jan31,1) = Feb28 (clamped), addMonths(Feb28,1) = Mar28
	got := calc.UsedUpDate(start, 2000, 1000, endDate)
	want := time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("UsedUpDate(月末起租2个月) = %v, want %v", got, want)
	}
}

func TestUsedUpDate_MonthEnd_ThreeMonths(t *testing.T) {
	start := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	// 付了3个月 = 3000
	// addMonths(Jan31,1) = Feb28, addMonths(Feb28,1) = Mar28, addMonths(Mar28,1) = Apr28
	got := calc.UsedUpDate(start, 3000, 1000, endDate)
	want := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("UsedUpDate(月末起租3个月) = %v, want %v", got, want)
	}
}

func TestUsedUpDate_LeapYear_MonthEnd(t *testing.T) {
	start := time.Date(2028, 1, 31, 0, 0, 0, 0, time.UTC) // 2028 is a leap year
	endDate := time.Date(2028, 12, 31, 0, 0, 0, 0, time.UTC)
	// 付1个月 = 1000
	// addMonths(Jan31,1) = Feb29 (clamped, leap year)
	got := calc.UsedUpDate(start, 1000, 1000, endDate)
	want := time.Date(2028, 2, 29, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("UsedUpDate(闰年月末) = %v, want %v", got, want)
	}
}

func TestTotalReceivable_FebEndOfMonth(t *testing.T) {
	// 2月28日起租到3月31日
	// addMonths(Feb28,1) = Mar28
	// 1 whole month, remaining: Mar28→Mar31 = 3 days
	start := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	got := calc.TotalReceivable(start, end, 1500)
	want := 1500.0 + 3*(1500.0/30.0)
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(2月底起租) = %f, want %f", got, want)
	}
}

// ── 新增: 真实日历月跨月验证 ─────────────────────────────────────────────────

func TestTotalReceivable_CrossMonths_Feb15ToMar20(t *testing.T) {
	// 2026-01-15 到 2026-03-20（跨 2 个完整月 + 5 天）
	// addMonths(Jan15,1)=Feb15, addMonths(Feb15,1)=Mar15
	// 2 whole months, remaining: Mar15→Mar20 = 5 days
	start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	rent := 3000.0
	got := calc.TotalReceivable(start, end, rent)
	want := 2*rent + 5*(rent/30.0) // 6000 + 500 = 6500
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(01-15→03-20) = %f, want %f", got, want)
	}
}

func TestTotalReceivable_MonthEndClamp_Jan31ToFeb28(t *testing.T) {
	// 2026-01-31 到 2026-02-28（月末钳位，恰好 1 个月）
	// addMonths(Jan31,1)=Feb28 (clamped), Feb28 == end → 0 remaining
	// 1 whole month, 0 remaining days
	start := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	rent := 5000.0
	got := calc.TotalReceivable(start, end, rent)
	want := 1 * rent // 5000
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalReceivable(月末钳位Jan31→Feb28) = %f, want %f", got, want)
	}
}

// ── 新增: ContractStatus 优先级验证 ───────────────────────────────────────────

func TestContractStatus_ExpiredButFullyPaid(t *testing.T) {
	// 已过期且已缴清 → paidup（paidup 优先于 expired）
	endDate := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 12000, 12000, today)
	if got != calc.StatusPaidUp {
		t.Errorf("ContractStatus(已过期+已缴清) = %s, want %s", got, calc.StatusPaidUp)
	}
}

func TestContractStatus_NoPaymentNotExpired(t *testing.T) {
	// 未收款未过期 → active
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 0, 12000, today)
	if got != calc.StatusActive {
		t.Errorf("ContractStatus(未收款未过期) = %s, want %s", got, calc.StatusActive)
	}
}

func TestContractStatus_PartialPaymentNotExpired(t *testing.T) {
	// 部分收款未过期 → arrears
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	got := calc.ContractStatus(endDate, 5000, 12000, today)
	if got != calc.StatusArrears {
		t.Errorf("ContractStatus(部分收款未过期) = %s, want %s", got, calc.StatusArrears)
	}
}

func TestContractStatus_PriorityOrder(t *testing.T) {
	// 综合优先级: paidup > expired > arrears > active
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name            string
		endDate         time.Time
		totalReceived   float64
		totalReceivable float64
		want            string
	}{
		{
			name:            "paidup优先: 过期+缴清→paidup",
			endDate:         time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   12000,
			totalReceivable: 12000,
			want:            calc.StatusPaidUp,
		},
		{
			name:            "expired: 过期+未缴清→expired",
			endDate:         time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   5000,
			totalReceivable: 12000,
			want:            calc.StatusExpired,
		},
		{
			name:            "arrears: 未过期+部分收款→arrears",
			endDate:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   3000,
			totalReceivable: 12000,
			want:            calc.StatusArrears,
		},
		{
			name:            "active: 未过期+零收款→active",
			endDate:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   0,
			totalReceivable: 12000,
			want:            calc.StatusActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.ContractStatus(tt.endDate, tt.totalReceived, tt.totalReceivable, today)
			if got != tt.want {
				t.Errorf("ContractStatus() = %s, want %s", got, tt.want)
			}
		})
	}
}

// ── 新增: ListUnpaid 概念验证 ─────────────────────────────────────────────────
// ListUnpaid 返回 status != "paidup" 的合同。
// 此处验证 ContractStatus 的输出与 ListUnpaid 过滤逻辑的一致性。

func TestListUnpaidConcept_AllStatusesExceptPaidUp(t *testing.T) {
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		endDate         time.Time
		totalReceived   float64
		totalReceivable float64
	shouldBeUnpaid  bool // ListUnpaid 应返回此合同
	}{
		{
			name:            "active → 应在未缴清列表中",
			endDate:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   0,
			totalReceivable: 12000,
			shouldBeUnpaid:  true,
		},
		{
			name:            "arrears → 应在未缴清列表中",
			endDate:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   5000,
			totalReceivable: 12000,
			shouldBeUnpaid:  true,
		},
		{
			name:            "expired → 应在未缴清列表中",
			endDate:         time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   5000,
			totalReceivable: 12000,
			shouldBeUnpaid:  true,
		},
		{
			name:            "paidup → 不应在未缴清列表中",
			endDate:         time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   12000,
			totalReceivable: 12000,
			shouldBeUnpaid:  false,
		},
		{
			name:            "expired+paidup → 不应在未缴清列表中",
			endDate:         time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			totalReceived:   12000,
			totalReceivable: 12000,
			shouldBeUnpaid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := calc.ContractStatus(tt.endDate, tt.totalReceived, tt.totalReceivable, today)
			isUnpaid := status != calc.StatusPaidUp
			if isUnpaid != tt.shouldBeUnpaid {
				t.Errorf("status=%s, isUnpaid=%v, want shouldBeUnpaid=%v", status, isUnpaid, tt.shouldBeUnpaid)
			}
		})
	}
}
