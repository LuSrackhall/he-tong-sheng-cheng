package handler

import (
	"asset-leasing-system/internal/docx"
	"asset-leasing-system/internal/domain"
	"asset-leasing-system/internal/domain/calc"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	offset, limit := parsePagination(c, 20, 100)

	contracts, total, err := h.repo.List(search, status, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合同列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contracts, "total": total})
}

func (h *ContractHandler) Create(c *gin.Context) {
	var req contractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入资产 ID、租户 ID、租期和月租金"})
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效，请使用 YYYY-MM-DD"})
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", req.EndDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效，请使用 YYYY-MM-DD"})
		return
	}
	if !endDate.After(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期必须晚于开始日期"})
		return
	}

	// 使用 SQL 条件查询检测合同重叠，避免加载全部活跃合同到内存
	overlap, err := h.repo.CheckOverlap(req.AssetID, req.TenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查已有合同失败"})
		return
	}
	if overlap {
		c.JSON(http.StatusConflict, gin.H{"error": "该资产与租户在此时间段已有合同"})
		return
	}

	if req.TotalReceivable <= 0 {
		req.TotalReceivable = calc.TotalReceivable(startDate, endDate, req.MonthlyRent)
	}

	if req.TemplateID != nil && *req.TemplateID != 0 {
		tpl, err := h.templateRepo.GetByID(*req.TemplateID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "指定的模板不存在"})
			return
		}
		if !tpl.Validated {
			c.JSON(http.StatusConflict, gin.H{"error": "所选模板暂不可用：Word 文件校验未通过"})
			return
		}
		requiredFields := []string{"startDate", "endDate", "monthlyRent", "tenantName", "assetName"}
		activeFields := parseActiveFields(tpl.ActiveFields)
		activeSet := make(map[string]bool)
		for k := range activeFields {
			activeSet[k] = true
		}
		var missingRequired []string
		for _, f := range requiredFields {
			if !activeSet[f] {
				missingRequired = append(missingRequired, f)
			}
		}
		if len(missingRequired) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":           fmt.Sprintf("所选模板暂不可用：缺少必填字段映射 %s", strings.Join(missingRequired, "、")),
				"missingRequired": missingRequired,
			})
			return
		}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建合同失败"})
		return
	}

	c.JSON(http.StatusCreated, contract)
}

func (h *ContractHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的合同 ID"})
		return
	}

	contract, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
		return
	}

	c.JSON(http.StatusOK, contract)
}

func (h *ContractHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的合同 ID"})
		return
	}

	contract, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
		return
	}

	var req contractReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
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
		t, err := time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "startDate 日期格式不正确，应为 YYYY-MM-DD"})
			return
		}
		contract.StartDate = t
	}
	if req.EndDate != "" {
		t, err := time.ParseInLocation("2006-01-02", req.EndDate, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "endDate 日期格式不正确，应为 YYYY-MM-DD"})
			return
		}
		contract.EndDate = t
	}
	if req.Notes != "" {
		contract.Notes = req.Notes
	}

	if err := h.repo.Update(contract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新合同失败"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模板列表失败"})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (h *ContractHandler) CreateTemplate(c *gin.Context) {
	var req templateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入模板名称"})
		return
	}

	tpl := &domain.Template{
		Name:     req.Name,
		FilePath: "", // filled after file upload
		FieldMap: `{
  "contractId": "合同编号",
  "startDate": "开始日期",
  "endDate": "结束日期",
  "monthlyRent": "月租金",
  "yearlyRent": "年租金",
  "tenantName": "租户姓名",
  "assetName": "资产名称"
}`,
		ActiveFields: `{
  "startDate": true,
  "endDate": true,
  "monthlyRent": true,
  "yearlyRent": true,
  "tenantName": true,
  "assetName": true
}`,
	}
	if err := h.templateRepo.Create(tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建模板失败"})
		return
	}
	c.JSON(http.StatusCreated, tpl)
}

func (h *ContractHandler) UpdateTemplateMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板 ID"})
		return
	}

	tpl, err := h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}

	var req struct {
		FieldMap     string `json:"fieldMap" binding:"required"`
		ActiveFields string `json:"activeFields"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入字段映射"})
		return
	}

	tpl.FieldMap = req.FieldMap
	if req.ActiveFields != "" {
		tpl.ActiveFields = req.ActiveFields
	}

	// Auto-sync activeFields from fieldMap: add new uncommented keys with default true
	activeMap := parseActiveFields(tpl.ActiveFields)
	if activeMap == nil {
		activeMap = make(map[string]bool)
	}
	uncommentedKeys := parseUncommentedFieldMapKeys(tpl.FieldMap)
	for _, k := range uncommentedKeys {
		if _, exists := activeMap[k]; !exists {
			activeMap[k] = true
		}
	}
	// Serialize synced activeFields
	activeBytes, _ := json.Marshal(activeMap)
	tpl.ActiveFields = string(activeBytes)

	// Enforce required fields: all must be present in activeFields
	requiredFields := []string{"startDate", "endDate", "monthlyRent", "tenantName", "assetName"}
	activeKeys := activeFieldKeys(tpl.ActiveFields)
	activeSet := make(map[string]bool)
	for _, k := range activeKeys {
		activeSet[k] = true
	}
	var missingRequired []string
	for _, f := range requiredFields {
		if !activeSet[f] {
			missingRequired = append(missingRequired, f)
		}
	}
	if len(missingRequired) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           fmt.Sprintf("缺少必填字段映射: %s", strings.Join(missingRequired, "、")),
			"missingRequired": missingRequired,
		})
		return
	}

	if err := h.templateRepo.Update(tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新模板映射失败"})
		return
	}

	// Re-validate existing Word file if present
	if tpl.FilePath != "" {
		fileData, err := os.ReadFile(tpl.FilePath)
		if err != nil {
			tpl.Validated = false
		} else {
			activeMap := parseActiveFields(tpl.ActiveFields)
			// Collect only fields with validate=true
			var validateFields []string
			for k, v := range activeMap {
				if v {
					validateFields = append(validateFields, k)
				}
			}
			if len(validateFields) > 0 {
				missing, err := docx.ValidatePlaceholders(fileData, validateFields)
				if err != nil {
					tpl.Validated = false
				} else {
					tpl.Validated = len(missing) == 0
				}
			} else {
				tpl.Validated = true
			}
		}
		h.templateRepo.Update(tpl)
	}

	c.JSON(http.StatusOK, tpl)
}
