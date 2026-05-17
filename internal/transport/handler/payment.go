package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	repo           domain.PaymentRepo
	contractRepo   domain.ContractRepo
	receiptBookRepo domain.ReceiptBookRepo
	receiptRepo    domain.ReceiptRepo
}

func NewPaymentHandler(repo domain.PaymentRepo, contractRepo domain.ContractRepo, rbRepo domain.ReceiptBookRepo, rRepo domain.ReceiptRepo) *PaymentHandler {
	return &PaymentHandler{repo, contractRepo, rbRepo, rRepo}
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

	contract, err := h.contractRepo.GetByID(uint(contractID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}

	paidAt := time.Now()
	if req.PaidAt != "" {
		if t, err := time.Parse("2006-01-02", req.PaidAt); err == nil {
			paidAt = t
		}
	}

	payment := &domain.Payment{
		ContractID: uint(contractID),
		Amount:     req.Amount,
		PaidAt:     paidAt,
		Notes:      req.Notes,
	}
	if err := h.repo.Create(payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	contract.TotalReceived += req.Amount
	if err := h.contractRepo.Update(contract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contract"})
		return
	}

	shortfall := contract.TotalReceivable - contract.TotalReceived
	if shortfall < 0 {
		shortfall = 0
	}

	c.JSON(http.StatusCreated, gin.H{
		"payment":   payment,
		"shortfall": shortfall,
	})
}
