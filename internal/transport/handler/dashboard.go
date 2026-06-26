package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	repo domain.DashboardRepo
}

func NewDashboardHandler(repo domain.DashboardRepo) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())

	activeContracts, err := h.repo.CountActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取活跃合同数失败"})
		return
	}

	monthlyRevenue, err := h.repo.MonthlyRevenue(year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取本月收款金额失败"})
		return
	}

	overdueContracts, err := h.repo.CountOverdue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取逾期合同数失败"})
		return
	}

	newContractsThisMonth, err := h.repo.CountNewThisMonth(year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取本月新增合同数失败"})
		return
	}

	c.JSON(http.StatusOK, domain.DashboardStats{
		ActiveContracts:       activeContracts,
		MonthlyRevenue:        monthlyRevenue,
		OverdueContracts:      overdueContracts,
		NewContractsThisMonth: newContractsThisMonth,
	})
}
