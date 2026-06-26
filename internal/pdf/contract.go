package pdf

import (
	"fmt"
	"html/template"
	"strings"
	"sync"
)

const templatePreviewHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>模板预览 — {{.Name}}</title>
<style>
  @page { size: A4 portrait; margin: 20mm; }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: "Songti SC", "SimSun", "宋体", serif; font-size: 14px; color: #000; line-height: 1.8; }
  .page { max-width: 700px; margin: 0 auto; padding: 20px; }
  .title { text-align: center; font-size: 20px; font-weight: bold; margin-bottom: 16px; }
  .hint { text-align: center; font-size: 13px; color: #666; margin-bottom: 24px; }
  .field-list { list-style: none; padding: 0; }
  .field-list li { padding: 8px 12px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; }
  .field-key { font-weight: 600; color: #007AFF; }
  .field-label { color: #555; }
  .required { color: #c00; font-size: 12px; }
  .note { margin-top: 24px; padding: 12px; background: #f8f8f8; border-radius: 8px; font-size: 13px; color: #666; }
  @media print { .no-print { display: none; } }
</style>
</head>
<body>
<div class="page">
  <div class="title">模板：{{.Name}}</div>
  <div class="hint">以下是该模板支持的字段列表。在 Word 文件中使用 ${字段名} 格式插入占位符。</div>
  <ul class="field-list">
    {{range .Fields}}
    <li>
      <span class="field-key">${{.Key}}</span>
      <span class="field-label">{{.Label}} {{if .Required}}<span class="required">*必填</span>{{end}}</span>
    </li>
    {{end}}
  </ul>
  <div class="note">
    提示：请确保 Word 文件中的占位符格式为 <strong>${字段名}</strong>，例如 <code>${startDate}</code>、<code>${tenantName}</code>。
    系统会自动将占位符替换为实际合同数据。
  </div>
</div>
</body>
</html>`

type TemplatePreviewField struct {
	Key      string
	Label    string
	Required bool
}

type TemplatePreviewData struct {
	Name   string
	Fields []TemplatePreviewField
}

// 模板预览模板缓存，首次调用时解析一次
var (
	tplPreviewTmplOnce sync.Once
	tplPreviewTmpl     *template.Template
	tplPreviewTmplErr  error
)

func getTemplatePreviewTemplate() (*template.Template, error) {
	tplPreviewTmplOnce.Do(func() {
		tplPreviewTmpl, tplPreviewTmplErr = template.New("tplPreview").Parse(templatePreviewHTML)
	})
	return tplPreviewTmpl, tplPreviewTmplErr
}

// GenerateTemplatePreviewHTML 生成模板预览 HTML
func GenerateTemplatePreviewHTML(data TemplatePreviewData) (string, error) {
	tmpl, err := getTemplatePreviewTemplate()
	if err != nil {
		return "", fmt.Errorf("解析模板预览失败: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板预览失败: %w", err)
	}
	return buf.String(), nil
}
