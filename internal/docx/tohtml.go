package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ToHTML 将 .docx 文件字节数据转换为 HTML 字符串。
// 解压 docx ZIP，解析 word/document.xml，将段落和表格转为 HTML。
func ToHTML(docxData []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(docxData), int64(len(docxData)))
	if err != nil {
		return "", fmt.Errorf("无法打开 docx 文件: %w", err)
	}

	var docXML []byte
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("无法读取 document.xml: %w", err)
			}
			docXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return "", fmt.Errorf("无法读取 document.xml: %w", err)
			}
			break
		}
	}

	if docXML == nil {
		return "", fmt.Errorf("docx 中未找到 word/document.xml")
	}

	return parseDocumentXML(docXML)
}

// docx XML 命名空间
const wNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

// docXML 用于解析整个 document.xml 的顶层结构
type docXML struct {
	Body docBody `xml:"body"`
}

type docBody struct {
	Elements []docElement `xml:",any"`
}

// docElement 是一个通配的 XML 元素
type docElement struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
}

// parseDocumentXML 解析 document.xml 并转换为 HTML
func parseDocumentXML(data []byte) (string, error) {
	var doc docXML
	if err := xml.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("解析 document.xml 失败: %w", err)
	}

	var buf strings.Builder
	for _, elem := range doc.Body.Elements {
		switch elem.XMLName.Local {
		case "p":
			if elem.XMLName.Space == "" || elem.XMLName.Space == wNS {
				html := convertParagraph(elem)
				buf.WriteString(html)
			}
		case "tbl":
			if elem.XMLName.Space == "" || elem.XMLName.Space == wNS {
				html := convertTable(elem.Content)
				buf.WriteString(html)
			}
		}
	}

	return strings.TrimSpace(buf.String()), nil
}

// paragraphProps 段落属性
type paragraphProps struct {
	Align     string // left, center, right, both
	FirstLine int    // 首行缩进（twips）
}

// convertParagraph 将 <w:p> 段落转为 <p>
func convertParagraph(elem docElement) string {
	// 解析段落属性和内容
	props, content := parseParagraphParts(elem.Content)

	var style string
	if props.Align != "" && props.Align != "left" {
		style += fmt.Sprintf("text-align:%s;", props.Align)
	}
	if props.FirstLine > 0 {
		// twips 转 em（1em ≈ 240 twips for 12pt）
		em := float64(props.FirstLine) / 240.0
		style += fmt.Sprintf("text-indent:%.2fem;", em)
	}

	if style != "" {
		return fmt.Sprintf("<p style=\"%s\">%s</p>", style, content)
	}
	return fmt.Sprintf("<p>%s</p>", content)
}

// parseParagraphParts 解析段落的属性和内联内容。
// content 来自两条路径：
//   1. xml.Unmarshal innerxml（命名空间已解析，Name.Local="r"）
//   2. RawToken 重建（未解析，Name.Local="w:r"）
// 使用 RawToken 统一处理两种情况。
func parseParagraphParts(content []byte) (paragraphProps, string) {
	var props paragraphProps
	var inlineParts []string

	decoder := xml.NewDecoder(bytes.NewReader(content))
	var inPPR bool
	var inRPR bool
	var currentRunBuf strings.Builder
	var runHasBold, runHasItalic bool

	for {
		raw, err := decoder.RawToken()
		if err != nil {
			break
		}

		switch t := raw.(type) {
		case xml.StartElement:
			local := stripPrefix(t.Name.Local)
			switch local {
			case "pPr":
				inPPR = true
			case "rPr":
				inRPR = true
			case "jc":
				if inPPR {
					for _, attr := range t.Attr {
						if stripPrefix(attr.Name.Local) == "val" {
							props.Align = attr.Value
						}
					}
				}
			case "ind":
				if inPPR {
					for _, attr := range t.Attr {
						if stripPrefix(attr.Name.Local) == "firstLine" {
							fmt.Sscanf(attr.Value, "%d", &props.FirstLine)
						}
					}
				}
			case "b":
				if inRPR {
					runHasBold = true
				}
			case "i":
				if inRPR {
					runHasItalic = true
				}
			case "r":
				runHasBold = false
				runHasItalic = false
				currentRunBuf.Reset()
			}

		case xml.EndElement:
			local := stripPrefix(t.Name.Local)
			switch local {
			case "pPr":
				inPPR = false
			case "rPr":
				inRPR = false
			case "r":
				text := currentRunBuf.String()
				if text != "" {
					if runHasBold {
						text = fmt.Sprintf("<strong>%s</strong>", text)
					}
					if runHasItalic {
						text = fmt.Sprintf("<em>%s</em>", text)
					}
					inlineParts = append(inlineParts, text)
				}
				currentRunBuf.Reset()
				runHasBold = false
				runHasItalic = false
			}

		case xml.CharData:
			text := string(t)
			if text != "" {
				currentRunBuf.WriteString(htmlEscape(text))
			}
		}
	}

	return props, strings.Join(inlineParts, "")
}

// stripPrefix 移除命名空间前缀（如 "w:r" → "r"，"r" → "r"）
func stripPrefix(local string) string {
	if i := strings.Index(local, ":"); i >= 0 {
		return local[i+1:]
	}
	return local
}

// convertTable 将 <w:tbl> 表格转为 <table>
func convertTable(content []byte) string {
	var buf strings.Builder
	buf.WriteString("<table>")

	decoder := xml.NewDecoder(bytes.NewReader(content))
	var inTC bool
	var paraDepth int
	var paraContent bytes.Buffer

	for {
		raw, err := decoder.RawToken()
		if err != nil {
			break
		}

		switch t := raw.(type) {
		case xml.StartElement:
			local := stripPrefix(t.Name.Local)
			switch {
			case local == "tr":
				buf.WriteString("<tr>")
			case local == "tc":
				inTC = true
				buf.WriteString("<td>")
			case inTC && local == "p":
				paraDepth = 1
				paraContent.Reset()
			case paraDepth > 0:
				paraDepth++
				writeStartElement(&paraContent, t)
			}

		case xml.EndElement:
			local := stripPrefix(t.Name.Local)
			switch {
			case paraDepth > 0:
				writeEndElement(&paraContent, t)
				paraDepth--
				if paraDepth == 0 {
					props, html := parseParagraphParts(paraContent.Bytes())
					var style string
					if props.Align != "" && props.Align != "left" {
						style += fmt.Sprintf("text-align:%s;", props.Align)
					}
					if props.FirstLine > 0 {
						em := float64(props.FirstLine) / 240.0
						style += fmt.Sprintf("text-indent:%.2fem;", em)
					}
					if style != "" {
						buf.WriteString(fmt.Sprintf("<p style=\"%s\">%s</p>", style, html))
					} else {
						buf.WriteString(fmt.Sprintf("<p>%s</p>", html))
					}
				}
			case local == "tr":
				buf.WriteString("</tr>")
			case local == "tc":
				inTC = false
				buf.WriteString("</td>")
			}

		case xml.CharData:
			if paraDepth > 0 {
				paraContent.Write(t)
			}
		}
	}

	buf.WriteString("</table>")
	return buf.String()
}

// writeStartElement 将 xml.StartElement（含完整限定名）序列化为 XML 字节
func writeStartElement(buf *bytes.Buffer, el xml.StartElement) {
	buf.WriteByte('<')
	writeName(buf, el.Name)
	for _, attr := range el.Attr {
		buf.WriteByte(' ')
		writeName(buf, attr.Name)
		buf.WriteString(`="`)
		xml.EscapeText(buf, []byte(attr.Value))
		buf.WriteByte('"')
	}
	buf.WriteByte('>')
}

// writeEndElement 将 xml.EndElement 序列化为 XML 字节
func writeEndElement(buf *bytes.Buffer, el xml.EndElement) {
	buf.WriteString("</")
	writeName(buf, el.Name)
	buf.WriteByte('>')
}

// writeName 输出 xml.Name（含命名空间前缀）
func writeName(buf *bytes.Buffer, name xml.Name) {
	if name.Space != "" {
		buf.WriteString(name.Space)
		buf.WriteByte(':')
	}
	buf.WriteString(name.Local)
}

// htmlEscape 转义 HTML 特殊字符
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
