package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Render replaces ${key} placeholders in a .docx template and writes the result.
// templateData is the raw bytes of the template .docx file.
// values maps placeholder keys (without ${}) to replacement strings.
func Render(templateData []byte, values map[string]string) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(templateData), int64(len(templateData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open template as zip: %w", err)
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for _, file := range reader.File {
		if err := copyZipFile(file, writer, values); err != nil {
			writer.Close()
			return nil, fmt.Errorf("failed to process file %s: %w", file.Name, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close output zip: %w", err)
	}

	return buf.Bytes(), nil
}

func copyZipFile(file *zip.File, writer *zip.Writer, values map[string]string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	// Only replace placeholders in document.xml (main content) and header/footer XMLs
	if strings.HasSuffix(file.Name, ".xml") {
		content = replacePlaceholders(content, values)
	}

	fw, err := writer.Create(file.Name)
	if err != nil {
		return err
	}
	_, err = fw.Write(content)
	return err
}

func replacePlaceholders(content []byte, values map[string]string) []byte {
	s := string(content)
	for key, val := range values {
		placeholder := "${" + key + "}"
		s = strings.ReplaceAll(s, placeholder, xmlEscape(val))
	}
	return []byte(s)
}

// xmlEscape escapes special XML characters for safe embedding in docx XML.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
