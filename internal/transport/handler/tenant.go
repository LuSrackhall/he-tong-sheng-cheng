package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	repo domain.TenantRepo
}

func NewTenantHandler(repo domain.TenantRepo) *TenantHandler {
	return &TenantHandler{repo: repo}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取租户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenants, "total": total})
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req tenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入租户名称"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建租户失败"})
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

func (h *TenantHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户 ID"})
		return
	}

	tenant, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "租户不存在"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

func (h *TenantHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户 ID"})
		return
	}

	tenant, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "租户不存在"})
		return
	}

	var req tenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新租户失败"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}
