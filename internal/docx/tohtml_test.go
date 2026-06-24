package docx

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

// createTestDocx 创建包含指定 document.xml 内容的测试 docx 字节数据
func createTestDocx(t *testing.T, documentXML string) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	f, err := w.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(documentXML)); err != nil {
		t.Fatal(err)
	}

	// docx 至少需要 [Content_Types].xml
	ct, err := w.Create("[Content_Types].xml")
	if err != nil {
		t.Fatal(err)
	}
	ct.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`))

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

const docxNS = `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`

func TestToHTML_PlainText(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
    <w:p><w:r><w:t>第一段</w:t></w:r></w:p>
    <w:p><w:r><w:t>第二段</w:t></w:r></w:p>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if !strings.Contains(html, "<p>第一段</p>") {
		t.Errorf("expected first paragraph, got: %s", html)
	}
	if !strings.Contains(html, "<p>第二段</p>") {
		t.Errorf("expected second paragraph, got: %s", html)
	}
}

func TestToHTML_BoldAndItalic(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
    <w:p>
      <w:r><w:rPr><w:b/></w:rPr><w:t>粗体</w:t></w:r>
      <w:r><w:rPr><w:i/></w:rPr><w:t>斜体</w:t></w:r>
      <w:r><w:rPr><w:b/><w:i/></w:rPr><w:t>粗斜</w:t></w:r>
      <w:r><w:t>普通</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if !strings.Contains(html, "<strong>粗体</strong>") {
		t.Errorf("expected bold, got: %s", html)
	}
	if !strings.Contains(html, "<em>斜体</em>") {
		t.Errorf("expected italic, got: %s", html)
	}
	if !strings.Contains(html, "<strong><em>粗斜</em></strong>") && !strings.Contains(html, "<em><strong>粗斜</strong></em>") {
		t.Errorf("expected bold+italic, got: %s", html)
	}
	if !strings.Contains(html, "普通") {
		t.Errorf("expected plain text, got: %s", html)
	}
}

func TestToHTML_Table(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
    <w:tbl>
      <w:tr>
        <w:tc><w:p><w:r><w:t>左上</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:t>右上</w:t></w:r></w:p></w:tc>
      </w:tr>
      <w:tr>
        <w:tc><w:p><w:r><w:t>左下</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:t>右下</w:t></w:r></w:p></w:tc>
      </w:tr>
    </w:tbl>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if !strings.Contains(html, "<table>") {
		t.Errorf("expected table tag, got: %s", html)
	}
	if !strings.Contains(html, "<td><p>左上</p></td>") {
		t.Errorf("expected cell content, got: %s", html)
	}
	if !strings.Contains(html, "<tr>") {
		t.Errorf("expected tr tags, got: %s", html)
	}
}

func TestToHTML_TableWithFormatting(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
    <w:tbl>
      <w:tr>
        <w:tc><w:p><w:r><w:rPr><w:b/></w:rPr><w:t>粗体单元格</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:rPr><w:i/></w:rPr><w:t>斜体单元格</w:t></w:r></w:p></w:tc>
      </w:tr>
    </w:tbl>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if !strings.Contains(html, "<strong>粗体单元格</strong>") {
		t.Errorf("expected bold in table cell, got: %s", html)
	}
	if !strings.Contains(html, "<em>斜体单元格</em>") {
		t.Errorf("expected italic in table cell, got: %s", html)
	}
}

func TestToHTML_EmptyDocument(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if html != "" {
		t.Errorf("expected empty string for empty document, got: %q", html)
	}
}

func TestToHTML_InvalidZip(t *testing.T) {
	_, err := ToHTML([]byte("not a zip file"))
	if err == nil {
		t.Error("expected error for invalid zip, got nil")
	}
}

func TestToHTML_MissingDocumentXML(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("other.xml")
	f.Write([]byte("<root/>"))
	w.Close()

	_, err := ToHTML(buf.Bytes())
	if err == nil {
		t.Error("expected error for missing document.xml, got nil")
	}
}

func TestToHTML_Alignment(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
    <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:t>居中</w:t></w:r></w:p>
    <w:p><w:pPr><w:jc w:val="right"/></w:pPr><w:r><w:t>右对齐</w:t></w:r></w:p>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if !strings.Contains(html, "text-align:center") {
		t.Errorf("expected center alignment, got: %s", html)
	}
	if !strings.Contains(html, "text-align:right") {
		t.Errorf("expected right alignment, got: %s", html)
	}
}

func TestToHTML_FirstLineIndent(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
    <w:p><w:pPr><w:ind w:firstLine="480"/></w:pPr><w:r><w:t>缩进段落</w:t></w:r></w:p>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if !strings.Contains(html, "text-indent") {
		t.Errorf("expected text-indent, got: %s", html)
	}
	if !strings.Contains(html, "2.00em") {
		t.Errorf("expected 2.00em indent (480/240), got: %s", html)
	}
}

func TestToHTML_MixedContent(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document ` + docxNS + `>
  <w:body>
    <w:p><w:r><w:t>合同标题</w:t></w:r></w:p>
    <w:tbl>
      <w:tr>
        <w:tc><w:p><w:r><w:t>甲</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:t>乙</w:t></w:r></w:p></w:tc>
      </w:tr>
    </w:tbl>
    <w:p><w:r><w:t>尾部文字</w:t></w:r></w:p>
  </w:body>
</w:document>`

	data := createTestDocx(t, xml)
	html, err := ToHTML(data)
	if err != nil {
		t.Fatalf("ToHTML error: %v", err)
	}

	if !strings.Contains(html, "<p>合同标题</p>") {
		t.Errorf("expected title paragraph, got: %s", html)
	}
	if !strings.Contains(html, "<table>") {
		t.Errorf("expected table, got: %s", html)
	}
	if !strings.Contains(html, "<p>尾部文字</p>") {
		t.Errorf("expected tail paragraph, got: %s", html)
	}
	// 验证顺序：标题在表格前，尾部在表格后
	titleIdx := strings.Index(html, "合同标题")
	tableIdx := strings.Index(html, "<table>")
	tailIdx := strings.Index(html, "尾部文字")
	if titleIdx > tableIdx || tableIdx > tailIdx {
		t.Errorf("wrong order: title=%d, table=%d, tail=%d", titleIdx, tableIdx, tailIdx)
	}
}
