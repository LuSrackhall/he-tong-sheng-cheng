package handler

import (
	"asset-leasing-system/internal/docx"
	"asset-leasing-system/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	tpl, err := h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required (field name: file)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .docx files are allowed"})
		return
	}

	// Read file into memory for validation
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	// Validate placeholders against active fields
	activeFields := parseActiveFields(tpl.ActiveFields)
	if len(activeFields) > 0 {
		missing, err := docx.ValidatePlaceholders(fileData, validateFields(activeFields))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse docx: " + err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	destPath := filepath.Join(uploadDir, fmt.Sprintf("template_%d.docx", id))
	if err := os.WriteFile(destPath, fileData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	tpl.FilePath = destPath
	if err := h.templateRepo.Update(tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}

	c.JSON(http.StatusOK, tpl)
}

// ExportContract handles POST /api/contracts/:id/export
func (h *ContractHandler) ExportContract(c *gin.Context) {
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

	if contract.TemplateID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contract has no template assigned"})
		return
	}

	tpl, err := h.templateRepo.GetByID(*contract.TemplateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	if tpl.FilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template file not uploaded yet"})
		return
	}

	if !tpl.Validated {
		c.JSON(http.StatusConflict, gin.H{"error": "模板暂不可用：Word 文件校验未通过，请重新上传符合要求的 Word 文件"})
		return
	}

	// Check required active fields
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
			"error":           fmt.Sprintf("模板暂不可用：缺少必填字段映射 %s，请在字段映射配置中启用", strings.Join(missingRequired, "、")),
			"missingRequired": missingRequired,
		})
		return
	}

	templateData, err := os.ReadFile(tpl.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read template file"})
		return
	}

	values := buildReplaceValues(contract, tpl)

	outputData, err := docx.Render(templateData, values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render document: " + err.Error()})
		return
	}

	exportDir := "uploads/exports"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create export directory"})
		return
	}

	exportPath := filepath.Join(exportDir, fmt.Sprintf("contract_%d.docx", id))
	if err := os.WriteFile(exportPath, outputData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save exported file"})
		return
	}

	downloadURL := fmt.Sprintf("/api/contracts/%d/download", id)
	c.JSON(http.StatusOK, gin.H{
		"message":     "Contract exported successfully",
		"downloadUrl": downloadURL,
		"filePath":    exportPath,
	})
}

// DownloadContract handles GET /api/contracts/:id/download
func (h *ContractHandler) DownloadContract(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract ID"})
		return
	}

	exportPath := filepath.Join("uploads/exports", fmt.Sprintf("contract_%d.docx", id))

	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exported file not found. Please export the contract first."})
		return
	}

	contract, err := h.repo.GetByID(uint(id))
	filename := fmt.Sprintf("contract_%d.docx", id)
	if err == nil && contract.Tenant != nil {
		filename = fmt.Sprintf("contract_%s_%s.docx",
			contract.Tenant.Name,
			time.Now().Format("20060102"))
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.File(exportPath)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	_, err = h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	used, err := h.templateRepo.IsUsedByContract(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check template usage"})
		return
	}
	if used {
		c.JSON(http.StatusConflict, gin.H{"error": "该模板已被合同引用，无法删除"})
		return
	}

	if err := h.templateRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模板已删除"})
}

// buildReplaceValues collects all placeholder values from contract, asset, tenant data.
func buildReplaceValues(contract *domain.Contract, tpl *domain.Template) map[string]string {
	values := make(map[string]string)

	// Parse fieldMap, stripping comment-prefixed lines
	fieldMap := make(map[string]string)
	if tpl.FieldMap != "" {
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

	values["today"] = time.Now().Format("2006-01-02")

	// Add all keys from fieldMap as values so custom fields get replaced
	for key := range fieldMap {
		if _, exists := values[key]; !exists {
			values[key] = ""
		}
	}

	return values
}
