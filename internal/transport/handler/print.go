package handler

import (
	"asset-leasing-system/internal/domain"
	"asset-leasing-system/internal/pdf"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PrintHandler struct {
	receiptRepo     domain.ReceiptRepo
	receiptBookRepo domain.ReceiptBookRepo
	paymentRepo     domain.PaymentRepo
	contractRepo    domain.ContractRepo
	tenantRepo      domain.TenantRepo
	assetRepo       domain.AssetRepo
	db              *gorm.DB
}

func NewPrintHandler(
	receiptRepo domain.ReceiptRepo,
	receiptBookRepo domain.ReceiptBookRepo,
	paymentRepo domain.PaymentRepo,
	contractRepo domain.ContractRepo,
	tenantRepo domain.TenantRepo,
	assetRepo domain.AssetRepo,
	db *gorm.DB,
) *PrintHandler {
	return &PrintHandler{receiptRepo, receiptBookRepo, paymentRepo, contractRepo, tenantRepo, assetRepo, db}
}

// PrintReceipt 生成三联收据 HTML（POST /api/print/receipt/:id，id = paymentID）
func (h *PrintHandler) PrintReceipt(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的收款记录 ID"})
		return
	}

	// 查找收款记录
	payment, err := h.paymentRepo.GetByID(uint(paymentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "收款记录不存在"})
		return
	}

	// 查找或创建收据记录
	receipt, err := h.receiptRepo.GetByPaymentID(uint(paymentID))
	if err != nil {
		// 收据不存在，自动创建（补打场景），在事务中保证序号分配和收据创建的原子性
		var createErr error
		err = h.db.Transaction(func(tx *gorm.DB) error {
			book, bookErr := h.receiptBookRepo.GetActive()
			if bookErr != nil {
				return fmt.Errorf("没有可用的收据本，请先创建收据本")
			}

			seq, seqErr := h.receiptBookRepo.AllocateSequence(book.ID)
			if seqErr != nil || seq == 0 {
				return fmt.Errorf("收据本已用完，请创建新的收据本")
			}

			receipt = &domain.Receipt{
				ReceiptBookID: book.ID,
				PaymentID:     uint(paymentID),
				SequenceNum:   seq,
				Amount:        payment.Amount,
				PrintedAt:     time.Now(),
			}
			if err := h.receiptRepo.Create(receipt); err != nil {
				return fmt.Errorf("创建收据记录失败")
			}
			return nil
		})
		if err != nil {
			createErr = err
			c.JSON(http.StatusBadRequest, gin.H{"error": createErr.Error()})
			return
		}
	}

	// 获取收据本信息
	book, err := h.receiptBookRepo.GetByID(receipt.ReceiptBookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取收据本信息失败"})
		return
	}

	// 获取合同、租户、资产信息
	contract, err := h.contractRepo.GetByID(payment.ContractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合同信息失败"})
		return
	}

	var tenantName, assetName string
	if contract.Tenant != nil {
		tenantName = contract.Tenant.Name
	} else if t, tErr := h.tenantRepo.GetByID(contract.TenantID); tErr == nil {
		tenantName = t.Name
	}

	if contract.Asset != nil {
		assetName = contract.Asset.Name
	} else if a, aErr := h.assetRepo.GetByID(contract.AssetID); aErr == nil {
		assetName = a.Name
	}

	// 生成收据号
	receiptNo := fmt.Sprintf("%s-%04d", book.Prefix, receipt.SequenceNum)

	// 生成 HTML
	html, err := pdf.GenerateReceiptHTML(pdf.ReceiptData{
		ReceiptNo:   receiptNo,
		Amount:      payment.Amount,
		AmountCN:    pdf.ConvertToCNAmount(payment.Amount),
		TenantName:  tenantName,
		AssetName:   assetName,
		Purpose:     fmt.Sprintf("租金（%s ~ %s）", contract.StartDate.Format("2006-01-02"), contract.EndDate.Format("2006-01-02")),
		PaidDate:    payment.PaidAt.Format("2006年01月02日"),
		Operator:    "", // 可从 JWT 中获取
		CompanyName: "租赁管理处",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成收据失败"})
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
