package handler

import (
	"asset-leasing-system/internal/docx"
	"asset-leasing-system/internal/domain"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadTemplate handles POST /api/templates/:id/upload
// Accepts multipart form file, saves to uploads/templates/, updates template FilePath.
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

	uploadDir := "uploads/templates"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	destPath := filepath.Join(uploadDir, fmt.Sprintf("template_%d.docx", id))
	if err := c.SaveUploadedFile(file, destPath); err != nil {
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
// Fills the contract's associated template with data and saves the output docx.
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
// Serves the previously exported docx file.
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

	// Get contract for filename
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

// buildReplaceValues collects all placeholder values from contract, asset, tenant data
// combined with the template's fieldMap.
func buildReplaceValues(contract *domain.Contract, tpl *domain.Template) map[string]string {
	values := make(map[string]string)

	// Parse fieldMap to know which fields are mapped (for display labels)
	fieldMap := make(map[string]string)
	if tpl.FieldMap != "" {
		_ = json.Unmarshal([]byte(tpl.FieldMap), &fieldMap)
	}

	// Contract fields
	values["contractId"] = fmt.Sprintf("%d", contract.ID)
	values["startDate"] = contract.StartDate.Format("2006-01-02")
	values["endDate"] = contract.EndDate.Format("2006-01-02")
	values["monthlyRent"] = fmt.Sprintf("%.2f", contract.MonthlyRent)
	values["totalReceivable"] = fmt.Sprintf("%.2f", contract.TotalReceivable)
	values["totalReceived"] = fmt.Sprintf("%.2f", contract.TotalReceived)
	values["deposit"] = fmt.Sprintf("%.2f", contract.Deposit)
	values["notes"] = contract.Notes
	values["status"] = contract.Status

	// Asset fields
	if contract.Asset != nil {
		values["assetName"] = contract.Asset.Name
		values["assetType"] = contract.Asset.AssetType
		values["assetDescription"] = contract.Asset.Description
	} else {
		values["assetName"] = ""
		values["assetType"] = ""
		values["assetDescription"] = ""
	}

	// Tenant fields
	if contract.Tenant != nil {
		values["tenantName"] = contract.Tenant.Name
		values["tenantIDCard"] = contract.Tenant.IDCard
		values["tenantPhone"] = contract.Tenant.Phone
	} else {
		values["tenantName"] = ""
		values["tenantIDCard"] = ""
		values["tenantPhone"] = ""
	}

	// Date field
	values["today"] = time.Now().Format("2006-01-02")

	// Also add the fieldMap labels themselves as values so users can map custom keys
	for key := range fieldMap {
		if _, exists := values[key]; !exists {
			values[key] = ""
		}
	}

	return values
}
