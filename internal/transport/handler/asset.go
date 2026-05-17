package handler

import (
	"asset-leasing-system/internal/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	repo domain.AssetRepo
}

func NewAssetHandler(repo domain.AssetRepo) *AssetHandler {
	return &AssetHandler{repo: repo}
}

type assetReq struct {
	Name        string `json:"name" binding:"required"`
	AssetType   string `json:"assetType"`
	Description string `json:"description,omitempty"`
	ExtraFields string `json:"extraFields,omitempty"`
}

func (h *AssetHandler) List(c *gin.Context) {
	search := c.Query("search")
	assetType := c.Query("type")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	assets, total, err := h.repo.List(search, assetType, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list assets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": assets, "total": total})
}

func (h *AssetHandler) Create(c *gin.Context) {
	var req assetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	if req.AssetType == "" {
		req.AssetType = "shop"
	}

	asset := &domain.Asset{
		Name:        req.Name,
		AssetType:   req.AssetType,
		Description: req.Description,
		ExtraFields: req.ExtraFields,
		Status:      "idle",
	}
	if err := h.repo.Create(asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create asset"})
		return
	}

	c.JSON(http.StatusCreated, asset)
}

func (h *AssetHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	asset, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

func (h *AssetHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	asset, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	var req assetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Name != "" {
		asset.Name = req.Name
	}
	if req.AssetType != "" {
		asset.AssetType = req.AssetType
	}
	asset.Description = req.Description
	if req.ExtraFields != "" {
		asset.ExtraFields = req.ExtraFields
	}

	if err := h.repo.Update(asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update asset"})
		return
	}

	c.JSON(http.StatusOK, asset)
}
