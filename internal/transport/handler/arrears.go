package handler

import (
	"asset-leasing-system/internal/domain"
	"asset-leasing-system/internal/domain/calc"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ArrearsHandler struct {
	contractRepo domain.ContractRepo
}

func NewArrearsHandler(contractRepo domain.ContractRepo) *ArrearsHandler {
	return &ArrearsHandler{contractRepo: contractRepo}
}

type arrearsContract struct {
	ID              uint    `json:"id"`
	Asset           any     `json:"asset"`
	Tenant          any     `json:"tenant"`
	TotalReceived   float64 `json:"totalReceived"`
	TotalReceivable float64 `json:"totalReceivable"`
	UsedUpDate      string  `json:"usedUpDate"`
	EndDate         string  `json:"endDate"`
	ArrearsLevel    int     `json:"arrearsLevel"`
	MonthlyRent     float64 `json:"monthlyRent"`
	Status          string  `json:"status"`
}

func (h *ArrearsHandler) List(c *gin.Context) {
	contracts, err := h.contractRepo.ListActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list contracts"})
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	var result []arrearsContract

	for _, ct := range contracts {
		level := calc.ClassifyArrears(
			calc.UsedUpDate(ct.StartDate, ct.TotalReceived, ct.MonthlyRent),
			ct.EndDate,
			ct.TotalReceived,
			ct.TotalReceivable,
			today,
		)

		if level == 0 {
			continue
		}

		usedUp := calc.UsedUpDate(ct.StartDate, ct.TotalReceived, ct.MonthlyRent)

		result = append(result, arrearsContract{
			ID:              ct.ID,
			Asset:           ct.Asset,
			Tenant:          ct.Tenant,
			TotalReceived:   ct.TotalReceived,
			TotalReceivable: ct.TotalReceivable,
			UsedUpDate:      usedUp.Format("2006-01-02"),
			EndDate:         ct.EndDate.Format("2006-01-02"),
			ArrearsLevel:    level,
			MonthlyRent:     ct.MonthlyRent,
			Status:          ct.Status,
		})
	}

	c.JSON(http.StatusOK, result)
}
