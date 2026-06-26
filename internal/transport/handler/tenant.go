package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	repo domain.TenantRepo
}

func NewTenantHandler(repo domain.TenantRepo) *TenantHandler {
	return &TenantHandler{repo: repo}
}

// maskIDCard 对身份证号进行脱敏，保留前4位和后4位
func maskIDCard(idCard string) string {
	if len(idCard) <= 8 {
		return idCard
	}
	prefix := idCard[:4]
	suffix := idCard[len(idCard)-4:]
	return prefix + strings.Repeat("*", len(idCard)-8) + suffix
}

type tenantReq struct {
	Name        string `json:"name" binding:"required"`
	Phone       string `json:"phone,omitempty"`
	IDCard      string `json:"idCard,omitempty"`
	IDCardImage string `json:"idCardImage,omitempty"`
	ExtraFields string `json:"extraFields,omitempty"`
}

func (h *TenantHandler) List(c *gin.Context) {
	search := c.Query("search")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tenants, total, err := h.repo.List(search, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tenants"})
		return
	}

	// List 接口对身份证号脱敏
	for i := range tenants {
		tenants[i].IDCard = maskIDCard(tenants[i].IDCard)
	}

	c.JSON(http.StatusOK, gin.H{"data": tenants, "total": total})
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req tenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	tenant := &domain.Tenant{
		Name:        req.Name,
		Phone:       req.Phone,
		IDCard:      req.IDCard,
		IDCardImage: req.IDCardImage,
		ExtraFields: req.ExtraFields,
	}
	if err := h.repo.Create(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

func (h *TenantHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

func (h *TenantHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req tenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Name != "" {
		tenant.Name = req.Name
	}
	tenant.Phone = req.Phone
	tenant.IDCard = req.IDCard
	if req.IDCardImage != "" {
		tenant.IDCardImage = req.IDCardImage
	}
	if req.ExtraFields != "" {
		tenant.ExtraFields = req.ExtraFields
	}

	if err := h.repo.Update(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}
