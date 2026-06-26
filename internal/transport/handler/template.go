package handler

import (
	"asset-leasing-system/internal/docx"
	"asset-leasing-system/internal/domain"
	"asset-leasing-system/internal/pdf"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadTemplate handles POST /api/templates/:id/upload
// Validates that the uploaded docx contains all activeField placeholders.
func (h *ContractHandler) UploadTemplate(c *gin.Context) {
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

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传文件（字段名: file）"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 .docx 格式文件"})
		return
	}

	// 文件大小限制（10MB）
	const maxTemplateSize int64 = 10 << 20
	if file.Size > maxTemplateSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件过大（最大 10MB）"})
		return
	}

	// Read file into memory for validation
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取上传文件失败"})
		return
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取上传文件失败"})
		return
	}

	// Validate placeholders against active fields
	activeFields := parseActiveFields(tpl.ActiveFields)
	if len(activeFields) > 0 {
		missing, err := docx.ValidatePlaceholders(fileData, validateFields(activeFields))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析 Word 文件失败: " + err.Error()})
			return
		}
		if len(missing) > 0 {
			tpl.Validated = false
			h.templateRepo.Update(tpl)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        "Word 文件缺少以下已启用的占位符",
				"missingFields": missing,
			})
			return
		}
	}

	tpl.Validated = true

	uploadDir := "uploads/templates"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	destPath := filepath.Join(uploadDir, fmt.Sprintf("template_%d.docx", id))
	if err := os.WriteFile(destPath, fileData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	tpl.FilePath = destPath
	if err := h.templateRepo.Update(tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新模板失败"})
		return
	}

	c.JSON(http.StatusOK, tpl)
}

// 预览 HTML 包裹模板
const contractPreviewWrapper = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>合同预览</title>
<style>
  @page { size: A4 portrait; margin: 20mm; }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: "Songti SC", "SimSun", "宋体", serif; font-size: 14px; color: #000; line-height: 1.8; }
  .page { max-width: 700px; margin: 0 auto; padding: 20px; }
  .hint-bar { background: #fff3cd; border: 1px solid #ffc107; border-radius: 8px; padding: 12px 16px; margin-bottom: 20px; font-size: 13px; color: #856404; text-align: center; }
  table { border-collapse: collapse; width: 100%%; margin: 8px 0; }
  td, th { border: 1px solid #999; padding: 6px 8px; text-align: left; }
  p { margin: 4px 0; }
  @media print { .no-print { display: none; } }
</style>
</head>
<body>
<div class="page">
  <div class="hint-bar">此为预览，精确格式请下载 Word 文件查看</div>
  %s
</div>
<button class="no-print" onclick="window.print()" style="position:fixed;top:20px;right:20px;padding:12px 24px;background:#007AFF;color:#fff;border:none;border-radius:8px;font-size:16px;cursor:pointer;z-index:1000;">打印 / 保存为 PDF</button>
</body>
</html>`

// validateContractForExport 校验合同是否满足导出条件（模板存在性、校验状态、必填字段）
func (h *ContractHandler) validateContractForExport(contract *domain.Contract) (*domain.Template, []byte, error) {
	if contract.TemplateID == nil {
		return nil, nil, fmt.Errorf("该合同未关联模板，请先在设置中上传并关联模板")
	}

	tpl, err := h.templateRepo.GetByID(*contract.TemplateID)
	if err != nil {
		return nil, nil, fmt.Errorf("模板不存在")
	}

	if tpl.FilePath == "" {
		return nil, nil, fmt.Errorf("模板文件尚未上传")
	}

	if !tpl.Validated {
		return nil, nil, fmt.Errorf("模板暂不可用：Word 文件校验未通过，请重新上传符合要求的 Word 文件")
	}

	// 检查必填字段
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
		return nil, nil, fmt.Errorf("模板暂不可用：缺少必填字段映射 %s，请在字段映射配置中启用", strings.Join(missingRequired, "、"))
	}

	templateData, err := os.ReadFile(tpl.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("无法读取模板文件")
	}

	return tpl, templateData, nil
}

// DownloadContract handles GET /api/contracts/:id/download
// 每次请求动态生成 Word 文件返回
func (h *ContractHandler) DownloadContract(c *gin.Context) {
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

	tpl, templateData, err := h.validateContractForExport(contract)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	values := buildReplaceValues(contract, tpl)
	outputData, err := docx.Render(templateData, values)
	if err != nil {
		log.Printf("生成合同文件失败 contract_id=%d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成合同文件失败，请重试"})
		return
	}

	filename := fmt.Sprintf("contract_%d.docx", id)
	if contract.Tenant != nil && contract.Tenant.Name != "" {
		filename = fmt.Sprintf("contract_%s_%s.docx",
			contract.Tenant.Name,
			time.Now().Format("20060102"))
	}

	// RFC 6266: 非 ASCII 文件名使用 filename*=UTF-8'' 编码
	asciiFilename := fmt.Sprintf("contract_%d.docx", id)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, asciiFilename, url.PathEscape(filename)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", outputData)
}

// PreviewContract handles GET /api/contracts/:id/preview
// 动态生成 Word → 转 HTML → 包裹基础 CSS 后返回
func (h *ContractHandler) PreviewContract(c *gin.Context) {
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

	tpl, templateData, err := h.validateContractForExport(contract)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	values := buildReplaceValues(contract, tpl)
	renderedData, err := docx.Render(templateData, values)
	if err != nil {
		log.Printf("生成合同文件失败 contract_id=%d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成合同文件失败，请重试"})
		return
	}

	bodyHTML, err := docx.ToHTML(renderedData)
	if err != nil {
		log.Printf("合同转 HTML 失败 contract_id=%d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "合同预览生成失败，请重试"})
		return
	}

	html := fmt.Sprintf(contractPreviewWrapper, bodyHTML)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// parseActiveFields parses the JSON active fields.
// Supports both legacy []string format (converted to map with true values)
// and new Record<string, boolean> format.
func parseActiveFields(raw string) map[string]bool {
	if raw == "" || raw == "null" {
		return nil
	}
	raw = strings.TrimSpace(raw)
	if raw[0] == '[' {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err != nil {
			return nil
		}
		result := make(map[string]bool, len(arr))
		for _, k := range arr {
			result[k] = true
		}
		return result
	}
	var obj map[string]bool
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil
	}
	return obj
}

// activeFieldKeys returns the keys from the activeFields map.
func activeFieldKeys(raw string) []string {
	m := parseActiveFields(raw)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// validateFields returns the keys from activeFields that have value true.
func validateFields(activeFields map[string]bool) []string {
	var result []string
	for k, v := range activeFields {
		if v {
			result = append(result, k)
		}
	}
	return result
}

// parseUncommentedFieldMapKeys parses the fieldMap JSON text, stripping
// comment-prefixed lines (//), and returns the keys of the remaining fields.
func parseUncommentedFieldMapKeys(raw string) []string {
	if raw == "" {
		return nil
	}
	fieldMap := make(map[string]string)
	uncommented := strings.Builder{}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			uncommented.WriteString(line + "\n")
		}
	}
	if err := json.Unmarshal([]byte(uncommented.String()), &fieldMap); err != nil {
		return nil
	}
	keys := make([]string, 0, len(fieldMap))
	for k := range fieldMap {
		keys = append(keys, k)
	}
	return keys
}

// DeleteTemplate handles DELETE /api/templates/:id
func (h *ContractHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板 ID"})
		return
	}

	_, err = h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}

	used, err := h.templateRepo.IsUsedByContract(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查模板使用情况失败"})
		return
	}
	if used {
		c.JSON(http.StatusConflict, gin.H{"error": "该模板已被合同引用，无法删除"})
		return
	}

	if err := h.templateRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模板已删除"})
}

// DownloadTemplate handles GET /api/templates/:id/download
func (h *ContractHandler) DownloadTemplate(c *gin.Context) {
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

	if tpl.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板文件尚未上传"})
		return
	}

	if _, err := os.Stat(tpl.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板文件不存在"})
		return
	}

	// 剥离可能导致 HTTP 头注入的字符，使用 RFC 5987 编码
	safeName := strings.NewReplacer("\r", "", "\n", "", `"`, "").Replace(tpl.Name)
	asciiFilename := fmt.Sprintf("template_%s.docx", safeName)
	encodedFilename := fmt.Sprintf("template_%s.docx", url.PathEscape(safeName))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, asciiFilename, encodedFilename))
	c.File(tpl.FilePath)
}

// PreviewTemplate handles GET /api/templates/:id/preview
func (h *ContractHandler) PreviewTemplate(c *gin.Context) {
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

	// 构建字段列表
	fields := []pdf.TemplatePreviewField{
		{Key: "contractId", Label: "合同编号", Required: false},
		{Key: "startDate", Label: "开始日期", Required: true},
		{Key: "endDate", Label: "结束日期", Required: true},
		{Key: "monthlyRent", Label: "月租金", Required: true},
		{Key: "yearlyRent", Label: "年租金", Required: false},
		{Key: "totalReceivable", Label: "应收总额", Required: false},
		{Key: "totalReceived", Label: "已收金额", Required: false},
		{Key: "deposit", Label: "押金", Required: false},
		{Key: "notes", Label: "备注", Required: false},
		{Key: "status", Label: "合同状态", Required: false},
		{Key: "assetName", Label: "资产名称", Required: true},
		{Key: "assetType", Label: "资产类型", Required: false},
		{Key: "assetDescription", Label: "资产描述", Required: false},
		{Key: "tenantName", Label: "租户姓名", Required: true},
		{Key: "tenantIDCard", Label: "身份证号", Required: false},
		{Key: "tenantPhone", Label: "联系电话", Required: false},
		{Key: "signingDate", Label: "签订日期", Required: false},
		{Key: "today", Label: "当天日期", Required: false},
	}

	// 添加自定义字段
	if tpl.FieldMap != "" {
		fieldMap := make(map[string]string)
		_ = json.Unmarshal([]byte(tpl.FieldMap), &fieldMap)
		builtinKeys := map[string]bool{
			"contractId": true, "startDate": true, "endDate": true,
			"monthlyRent": true, "yearlyRent": true, "totalReceivable": true,
			"totalReceived": true, "deposit": true, "notes": true, "status": true,
			"assetName": true, "assetType": true, "assetDescription": true,
			"tenantName": true, "tenantIDCard": true, "tenantPhone": true,
			"signingDate": true, "today": true,
		}
		for key, label := range fieldMap {
			if !builtinKeys[key] {
				fields = append(fields, pdf.TemplatePreviewField{Key: key, Label: label, Required: false})
			}
		}
	}

	html, err := pdf.GenerateTemplatePreviewHTML(pdf.TemplatePreviewData{
		Name:   tpl.Name,
		Fields: fields,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成模板预览失败"})
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// buildReplaceValues collects all placeholder values from contract, asset, tenant data.
func buildReplaceValues(contract *domain.Contract, tpl *domain.Template) map[string]string {
	values := make(map[string]string)

	// Parse fieldMap, stripping comment-prefixed lines
	fieldMap := make(map[string]string)
	if tpl != nil && tpl.FieldMap != "" {
		uncommented := strings.Builder{}
		for _, line := range strings.Split(tpl.FieldMap, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
				uncommented.WriteString(line + "\n")
			}
		}
		_ = json.Unmarshal([]byte(uncommented.String()), &fieldMap)
	}

	// Contract fields
	values["contractId"] = fmt.Sprintf("%d", contract.ID)
	values["startDate"] = contract.StartDate.Format("2006-01-02")
	values["endDate"] = contract.EndDate.Format("2006-01-02")
	values["monthlyRent"] = fmt.Sprintf("%.2f", contract.MonthlyRent)
	values["yearlyRent"] = fmt.Sprintf("%.2f", contract.MonthlyRent*12)
	values["totalReceivable"] = fmt.Sprintf("%.2f", contract.TotalReceivable)
	values["totalReceived"] = fmt.Sprintf("%.2f", contract.TotalReceived)
	values["deposit"] = fmt.Sprintf("%.2f", contract.Deposit)
	values["notes"] = contract.Notes
	values["status"] = contract.Status

	if contract.Asset != nil {
		values["assetName"] = contract.Asset.Name
		values["assetType"] = contract.Asset.AssetType
		values["assetDescription"] = contract.Asset.Description
	} else {
		values["assetName"] = ""
		values["assetType"] = ""
		values["assetDescription"] = ""
	}

	if contract.Tenant != nil {
		values["tenantName"] = contract.Tenant.Name
		values["tenantIDCard"] = contract.Tenant.IDCard
		values["tenantPhone"] = contract.Tenant.Phone
	} else {
		values["tenantName"] = ""
		values["tenantIDCard"] = ""
		values["tenantPhone"] = ""
	}

	values["signingDate"] = contract.CreatedAt.Format("2006-01-02")
	values["today"] = time.Now().Format("2006-01-02")

	// Add all keys from fieldMap as values so custom fields get replaced
	for key := range fieldMap {
		if _, exists := values[key]; !exists {
			values[key] = ""
		}
	}

	return values
}
