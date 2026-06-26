package domain

// DashboardRepo 仪表盘数据查询接口
type DashboardRepo interface {
	CountActive() (int64, error)
	MonthlyRevenue(year int, month int) (float64, error)
	CountOverdue() (int64, error)
	CountNewThisMonth(year int, month int) (int64, error)
}
