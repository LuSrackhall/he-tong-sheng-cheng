package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReceiptBookHandler struct {
	repo domain.ReceiptBookRepo
}

func NewReceiptBookHandler(repo domain.ReceiptBookRepo) *ReceiptBookHandler {
	return &ReceiptBookHandler{repo: repo}
}

type receiptBookReq struct {
	Prefix     string `json:"prefix" binding:"required"`
	StartNum   int    `json:"startNum"`
	TotalPages int    `json:"totalPages" binding:"required"`
}

func (h *ReceiptBookHandler) List(c *gin.Context) {
	books, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list receipt books"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": books, "total": len(books)})
}

func (h *ReceiptBookHandler) Create(c *gin.Context) {
	var req receiptBookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix and total pages are required"})
		return
	}

	if req.StartNum <= 0 {
		req.StartNum = 1
	}

	book := &domain.ReceiptBook{
		Prefix:     req.Prefix,
		StartNum:   req.StartNum,
		CurrentNum: req.StartNum,
		TotalPages: req.TotalPages,
		Status:     "active",
	}
	if err := h.repo.Create(book); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create receipt book"})
		return
	}

	c.JSON(http.StatusCreated, book)
}
