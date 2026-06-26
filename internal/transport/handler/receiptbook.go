package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReceiptBookHandler struct {
	repo domain.ReceiptBookRepo
}

type ReceiptHandler struct {
	repo domain.ReceiptRepo
}

func NewReceiptBookHandler(repo domain.ReceiptBookRepo) *ReceiptBookHandler {
	return &ReceiptBookHandler{repo: repo}
}

func NewReceiptHandler(repo domain.ReceiptRepo) *ReceiptHandler {
	return &ReceiptHandler{repo: repo}
}

type receiptBookReq struct {
	Prefix     string `json:"prefix" binding:"required"`
	StartNum   int    `json:"startNum"`
	TotalPages int    `json:"totalPages" binding:"required"`
}

func (h *ReceiptBookHandler) List(c *gin.Context) {
	books, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取收据本列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": books, "total": len(books)})
}

func (h *ReceiptBookHandler) Create(c *gin.Context) {
	var req receiptBookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入收据本前缀和总页数"})
		return
	}

	if req.StartNum <= 0 {
		req.StartNum = 1
	}
	if req.TotalPages <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TotalPages must be greater than 0"})
		return
	}

	book := &domain.ReceiptBook{
		Prefix:     req.Prefix,
		StartNum:   req.StartNum,
		CurrentNum: req.StartNum,
		TotalPages: req.TotalPages,
		Status:     "active",
	}
	if err := h.repo.Create(book); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建收据本失败"})
		return
	}

	c.JSON(http.StatusCreated, book)
}

// ListReceipts 列出所有收据（GET /api/receipts）
func (h *ReceiptHandler) ListReceipts(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	receipts, total, err := h.repo.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取收据列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": receipts, "total": total})
}
