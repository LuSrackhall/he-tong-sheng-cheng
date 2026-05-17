package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ContractHandler struct {
	repo         domain.ContractRepo
	templateRepo domain.TemplateRepo
}

func NewContractHandler(repo domain.ContractRepo, templateRepo domain.TemplateRepo) *ContractHandler {
	return &ContractHandler{repo: repo, templateRepo: templateRepo}
}

type contractReq struct {
	AssetID         uint    `json:"assetId" binding:"required"`
	TenantID        uint    `json:"tenantId" binding:"required"`
	StartDate       string  `json:"startDate" binding:"required"`
	EndDate         string  `json:"endDate" binding:"required"`
	MonthlyRent     float64 `json:"monthlyRent" binding:"required"`
	TotalReceivable float64 `json:"totalReceivable"`
	Deposit         float64 `json:"deposit"`
	TemplateID      *uint   `json:"templateId"`
	Notes           string  `json:"notes,omitempty"`
}

func (h *ContractHandler) List(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	contracts, total, err := h.repo.List(search, status, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list contracts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contracts, "total": total})
}

func (h *ContractHandler) Create(c *gin.Context) {
	var req contractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Asset ID, tenant ID, dates, and monthly rent are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format, use YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format, use YYYY-MM-DD"})
		return
	}
	if !endDate.After(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	if req.TotalReceivable <= 0 {
		// auto-calculate
		wholeMonths := int(endDate.Sub(startDate).Hours()/24) / 30
		remainingDays := int(endDate.Sub(startDate).Hours()/24) % 30
		dailyRate := req.MonthlyRent / 30.0
		req.TotalReceivable = float64(wholeMonths)*req.MonthlyRent + float64(remainingDays)*dailyRate
	}

	contract := &domain.Contract{
		AssetID:         req.AssetID,
		TenantID:        req.TenantID,
		StartDate:       startDate,
		EndDate:         endDate,
		MonthlyRent:     req.MonthlyRent,
		TotalReceivable: req.TotalReceivable,
		TotalReceived:   0,
		Deposit:         req.Deposit,
		Status:          "active",
		TemplateID:      req.TemplateID,
		Notes:           req.Notes,
	}
	if err := h.repo.Create(contract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contract"})
		return
	}

	c.JSON(http.StatusCreated, contract)
}

func (h *ContractHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract ID"})
		return
	}

	contract, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}

	c.JSON(http.StatusOK, contract)
}

func (h *ContractHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract ID"})
		return
	}

	contract, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}

	var req contractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.MonthlyRent > 0 {
		contract.MonthlyRent = req.MonthlyRent
	}
	if req.TotalReceivable > 0 {
		contract.TotalReceivable = req.TotalReceivable
	}
	if req.Deposit > 0 {
		contract.Deposit = req.Deposit
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			contract.StartDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			contract.EndDate = t
		}
	}
	if req.Notes != "" {
		contract.Notes = req.Notes
	}

	if err := h.repo.Update(contract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contract"})
		return
	}

	c.JSON(http.StatusOK, contract)
}

type templateReq struct {
	Name string `json:"name" binding:"required"`
}

func (h *ContractHandler) ListTemplates(c *gin.Context) {
	templates, err := h.templateRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (h *ContractHandler) CreateTemplate(c *gin.Context) {
	var req templateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template name is required"})
		return
	}

	tpl := &domain.Template{
		Name:     req.Name,
		FilePath: "", // filled after file upload
		FieldMap: "{}",
	}
	if err := h.templateRepo.Create(tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}
	c.JSON(http.StatusCreated, tpl)
}

func (h *ContractHandler) UpdateTemplateMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	tpl, err := h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	var req struct {
		FieldMap string `json:"fieldMap" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Field map is required"})
		return
	}

	tpl.FieldMap = req.FieldMap
	if err := h.templateRepo.Update(tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template mapping"})
		return
	}

	c.JSON(http.StatusOK, tpl)
}
