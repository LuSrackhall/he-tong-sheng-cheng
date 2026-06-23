package pdf

import (
	"fmt"
	"html/template"
	"strings"
)

type ContractData struct {
	ContractID      string
	StartDate       string
	EndDate         string
	MonthlyRent     string
	YearlyRent      string
	TotalReceivable string
	TotalReceived   string
	Deposit         string
	Status          string
	Notes           string
	TenantName      string
	TenantIDCard    string
	TenantPhone     string
	AssetName       string
	AssetType       string
	AssetDescription string
	Today           string
	// 自定义字段
	CustomFields    map[string]string
}

const contractHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>合同 {{.ContractID}}</title>
<style>
  @page { size: A4 portrait; margin: 20mm; }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: "Songti SC", "SimSun", "宋体", serif; font-size: 14px; color: #000; line-height: 1.8; }
  .page { max-width: 700px; margin: 0 auto; padding: 20px; }
  .title { text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 8px; margin-bottom: 30px; }
  .subtitle { text-align: center; font-size: 13px; color: #666; margin-bottom: 24px; }
  .section { margin-bottom: 20px; }
  .section-title { font-size: 16px; font-weight: bold; margin-bottom: 8px; padding-bottom: 4px; border-bottom: 2px solid #333; }
  .info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px 24px; }
  .info-row { display: flex; }
  .info-label { color: #555; min-width: 80px; flex-shrink: 0; }
  .info-value { flex: 1; border-bottom: 1px solid #ccc; padding-left: 4px; }
  .clause { margin-bottom: 12px; text-indent: 2em; }
  .amount { font-size: 18px; font-weight: bold; color: #c00; }
  .sign-area { display: flex; justify-content: space-between; margin-top: 60px; }
  .sign-box { width: 45%; }
  .sign-box .line { border-bottom: 1px solid #333; height: 40px; margin-top: 8px; }
  .sign-label { font-size: 13px; color: #555; }
  .footer { text-align: center; margin-top: 40px; font-size: 12px; color: #999; }
  .custom-fields { margin-top: 16px; }
  .custom-fields .info-row { margin-bottom: 4px; }
  @media print {
    body { -webkit-print-color-adjust: exact; }
    .no-print { display: none; }
    .page { padding: 0; }
  }
</style>
</head>
<body>
<div class="page">
  <div class="title">租赁合同</div>
  <div class="subtitle">合同编号：{{.ContractID}}</div>

  <div class="section">
    <div class="section-title">合同双方</div>
    <div class="info-grid">
      <div class="info-row"><span class="info-label">出租方：</span><span class="info-value">______________________</span></div>
      <div class="info-row"><span class="info-label">承租方：</span><span class="info-value">{{.TenantName}}</span></div>
      <div class="info-row"><span class="info-label">身份证：</span><span class="info-value">{{.TenantIDCard}}</span></div>
      <div class="info-row"><span class="info-label">联系电话：</span><span class="info-value">{{.TenantPhone}}</span></div>
    </div>
  </div>

  <div class="section">
    <div class="section-title">租赁资产</div>
    <div class="info-grid">
      <div class="info-row"><span class="info-label">资产名称：</span><span class="info-value">{{.AssetName}}</span></div>
      <div class="info-row"><span class="info-label">资产类型：</span><span class="info-value">{{.AssetType}}</span></div>
      {{if .AssetDescription}}<div class="info-row" style="grid-column: 1/-1;"><span class="info-label">资产描述：</span><span class="info-value">{{.AssetDescription}}</span></div>{{end}}
    </div>
  </div>

  <div class="section">
    <div class="section-title">租赁条款</div>
    <div class="info-grid">
      <div class="info-row"><span class="info-label">起始日期：</span><span class="info-value">{{.StartDate}}</span></div>
      <div class="info-row"><span class="info-label">结束日期：</span><span class="info-value">{{.EndDate}}</span></div>
      <div class="info-row"><span class="info-label">月租金：</span><span class="info-value amount">¥{{.MonthlyRent}}</span></div>
      <div class="info-row"><span class="info-label">年租金：</span><span class="info-value">¥{{.YearlyRent}}</span></div>
      <div class="info-row"><span class="info-label">押金：</span><span class="info-value">¥{{.Deposit}}</span></div>
      <div class="info-row"><span class="info-label">应收总额：</span><span class="info-value amount">¥{{.TotalReceivable}}</span></div>
    </div>
  </div>

  {{if .CustomFields}}
  <div class="section">
    <div class="section-title">补充信息</div>
    <div class="custom-fields">
      {{range $key, $val := .CustomFields}}
      <div class="info-row"><span class="info-label">{{$key}}：</span><span class="info-value">{{$val}}</span></div>
      {{end}}
    </div>
  </div>
  {{end}}

  {{if .Notes}}
  <div class="section">
    <div class="section-title">备注</div>
    <div class="clause">{{.Notes}}</div>
  </div>
  {{end}}

  <div class="section">
    <div class="section-title">合同条款</div>
    <div class="clause">一、租赁期间，承租方应按时支付租金，不得拖欠。逾期未缴的，出租方有权按催缴程序处理。</div>
    <div class="clause">二、承租方应合理使用租赁资产，不得擅自转租、改造或用于非法用途。</div>
    <div class="clause">三、租赁期满，承租方应按时归还资产。如需续租，双方应另行协商签订新合同。</div>
    <div class="clause">四、本合同一式两份，双方各执一份，自双方签字之日起生效。</div>
  </div>

  <div class="sign-area">
    <div class="sign-box">
      <div class="sign-label">出租方（盖章）：</div>
      <div class="line"></div>
      <div class="sign-label" style="margin-top: 8px;">日期：____年____月____日</div>
    </div>
    <div class="sign-box">
      <div class="sign-label">承租方（签字）：</div>
      <div class="line"></div>
      <div class="sign-label" style="margin-top: 8px;">日期：{{.StartDate}}</div>
    </div>
  </div>

  <div class="footer">
    本合同由租赁管理系统自动生成 · {{.Today}}
  </div>
</div>
<button class="no-print" onclick="window.print()" style="position:fixed;top:20px;right:20px;padding:12px 24px;background:#007AFF;color:#fff;border:none;border-radius:8px;font-size:16px;cursor:pointer;z-index:1000;">打印 / 保存为 PDF</button>
</body>
</html>`

// GenerateContractHTML 生成合同预览 HTML
func GenerateContractHTML(data ContractData) (string, error) {
	tmpl, err := template.New("contract").Parse(contractHTML)
	if err != nil {
		return "", fmt.Errorf("解析合同模板失败: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染合同模板失败: %w", err)
	}
	return buf.String(), nil
}

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

// GenerateTemplatePreviewHTML 生成模板预览 HTML
func GenerateTemplatePreviewHTML(data TemplatePreviewData) (string, error) {
	tmpl, err := template.New("tplPreview").Parse(templatePreviewHTML)
	if err != nil {
		return "", fmt.Errorf("解析模板预览失败: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板预览失败: %w", err)
	}
	return buf.String(), nil
}
