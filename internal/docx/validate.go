package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ValidatePlaceholders checks that all expected field placeholders exist in the docx.
// Returns a list of missing fields, or nil if all are present.
func ValidatePlaceholders(templateData []byte, fields []string) ([]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(templateData), int64(len(templateData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open docx: %w", err)
	}

	var allXMLContent strings.Builder
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".xml") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			allXMLContent.Write(content)
		}
	}

	xmlStr := allXMLContent.String()

	// Extract all placeholder keys found in the document (strip XML tags from captured groups)
	foundKeys := make(map[string]bool)
	for _, m := range placeholderRe.FindAllStringSubmatch(xmlStr, -1) {
		key := stripTagsRe.ReplaceAllString(m[1], "")
		if key != "" {
			foundKeys[key] = true
		}
	}

	var missing []string
	for _, field := range fields {
		if !foundKeys[field] {
			missing = append(missing, field)
		}
	}

	return missing, nil
}
