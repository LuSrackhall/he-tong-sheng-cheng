package handler

import (
	"asset-leasing-system/internal/domain"
	"asset-leasing-system/internal/domain/calc"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	repo           domain.PaymentRepo
	contractRepo   domain.ContractRepo
	receiptBookRepo domain.ReceiptBookRepo
	receiptRepo    domain.ReceiptRepo
	db             *gorm.DB
}

func NewPaymentHandler(repo domain.PaymentRepo, contractRepo domain.ContractRepo, rbRepo domain.ReceiptBookRepo, rRepo domain.ReceiptRepo, db *gorm.DB) *PaymentHandler {
	return &PaymentHandler{repo, contractRepo, rbRepo, rRepo, db}
}

type paymentReq struct {
	Amount float64 `json:"amount" binding:"required"`
	PaidAt string  `json:"paidAt"`
	Notes  string  `json:"notes,omitempty"`
}

func (h *PaymentHandler) ListByContract(c *gin.Context) {
	contractID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract ID"})
		return
	}

	payments, err := h.repo.ListByContractID(uint(contractID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list payments"})
		return
	}

	c.JSON(http.StatusOK, payments)
}

func (h *PaymentHandler) Create(c *gin.Context) {
	contractID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract ID"})
		return
	}

	var req paymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount is required"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be positive"})
		return
	}

	paidAt := time.Now()
	if req.PaidAt != "" {
		if t, err := time.ParseInLocation("2006-01-02", req.PaidAt, time.Local); err == nil {
			paidAt = t
		}
	}

	payment := &domain.Payment{
		ContractID: uint(contractID),
		Amount:     req.Amount,
		PaidAt:     paidAt,
		Notes:      req.Notes,
	}

	var shortfall float64

	// 事务保护：收款记录 + 合同更新必须原子完成
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		var contract domain.Contract
		if err := tx.First(&contract, contractID).Error; err != nil {
			return err
		}

		contract.TotalReceived += req.Amount
		contract.Status = calc.ContractStatus(contract.EndDate, contract.TotalReceived, contract.TotalReceivable, time.Now())
		if err := tx.Save(&contract).Error; err != nil {
			return err
		}

		shortfall = contract.TotalReceivable - contract.TotalReceived
		if shortfall < 0 {
			shortfall = 0
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "收款失败，请重试"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"payment":   payment,
		"shortfall": shortfall,
	})
}
