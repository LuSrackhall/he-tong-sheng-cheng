package pdf

import (
	"fmt"
	"html/template"
	"strings"
	"sync"
)

type ReceiptData struct {
	ReceiptNo   string  // 完整收据号 = 前缀 + 序号
	Amount      float64 // 收款金额
	AmountCN    string  // 大写金额
	TenantName  string  // 交款人
	AssetName   string  // 资产名称
	Purpose     string  // 收款事由
	PaidDate    string  // 收款日期
	Operator    string  // 收款人
	CompanyName string  // 收款单位
}

const receiptHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>收据 {{.ReceiptNo}}</title>
<style>
  @page { size: A4 portrait; margin: 0; }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: "Songti SC", "SimSun", "宋体", serif; font-size: 14px; color: #000; }
  .page { width: 210mm; height: 297mm; padding: 5mm; }
  .receipt { width: 200mm; height: 90mm; padding: 8mm 12mm; position: relative; }
  .receipt:not(:last-child) { border-bottom: 1px dashed #999; }
  .receipt-title { text-align: center; font-size: 22px; font-weight: bold; letter-spacing: 6px; margin-bottom: 4px; }
  .receipt-type { text-align: center; font-size: 13px; color: #666; margin-bottom: 12px; }
  .receipt-row { display: flex; margin-bottom: 8px; line-height: 1.8; }
  .receipt-label { width: 80px; flex-shrink: 0; color: #333; }
  .receipt-value { flex: 1; border-bottom: 1px solid #333; padding-left: 4px; }
  .receipt-amount { font-size: 16px; font-weight: bold; }
  .receipt-footer { display: flex; justify-content: space-between; margin-top: 16px; }
  .receipt-footer div { width: 45%; }
  .receipt-footer .label { color: #333; }
  .receipt-footer .line { border-bottom: 1px solid #333; height: 24px; margin-top: 4px; }
  @media print {
    body { -webkit-print-color-adjust: exact; }
    .no-print { display: none; }
  }
</style>
</head>
<body>
<div class="page">
  {{range .Copies}}
  <div class="receipt">
    <div class="receipt-title">收&nbsp;&nbsp;&nbsp;&nbsp;据</div>
    <div class="receipt-type">{{.CopyLabel}}</div>
    <div class="receipt-row">
      <span class="receipt-label">收据号：</span>
      <span class="receipt-value">{{$.ReceiptNo}}</span>
    </div>
    <div class="receipt-row">
      <span class="receipt-label">交款人：</span>
      <span class="receipt-value">{{$.TenantName}}</span>
    </div>
    <div class="receipt-row">
      <span class="receipt-label">收款事由：</span>
      <span class="receipt-value">{{$.Purpose}}（{{$.AssetName}}）</span>
    </div>
    <div class="receipt-row">
      <span class="receipt-label">金&nbsp;&nbsp;额：</span>
      <span class="receipt-value receipt-amount">¥{{printf "%.2f" $.Amount}}</span>
    </div>
    <div class="receipt-row">
      <span class="receipt-label">大写金额：</span>
      <span class="receipt-value">{{$.AmountCN}}</span>
    </div>
    <div class="receipt-footer">
      <div><span class="label">收款单位（盖章）：</span><div class="line"></div></div>
      <div><span class="label">收款人：{{$.Operator}}</span><div class="line"></div></div>
    </div>
    <div style="text-align: right; margin-top: 6px; font-size: 12px; color: #999;">{{$.PaidDate}}</div>
  </div>
  {{end}}
</div>
<button class="no-print" onclick="window.print()" style="position:fixed;top:20px;right:20px;padding:12px 24px;background:#007AFF;color:#fff;border:none;border-radius:8px;font-size:16px;cursor:pointer;">打印收据</button>
</body>
</html>`

type copyData struct {
	CopyLabel string
}

// 收据模板缓存，首次调用时解析一次
var (
	receiptTmplOnce sync.Once
	receiptTmpl     *template.Template
	receiptTmplErr  error
)

func getReceiptTemplate() (*template.Template, error) {
	receiptTmplOnce.Do(func() {
		receiptTmpl, receiptTmplErr = template.New("receipt").Parse(receiptHTML)
	})
	return receiptTmpl, receiptTmplErr
}

// GenerateReceiptHTML 生成三联收据 HTML
func GenerateReceiptHTML(data ReceiptData) (string, error) {
	tmpl, err := getReceiptTemplate()
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	type pageData struct {
		ReceiptData
		Copies []copyData
	}

	page := pageData{
		ReceiptData: data,
		Copies: []copyData{
			{CopyLabel: "存根联（收款单位留存）"},
			{CopyLabel: "收据联（交款人留存）"},
			{CopyLabel: "记账联（财务留存）"},
		},
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, page); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}
	return buf.String(), nil
}

// ConvertToCNAmount 将金额转换为大写中文
func ConvertToCNAmount(amount float64) string {
	digits := []string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}
	units := []string{"", "拾", "佰", "仟"}
	bigUnits := []string{"", "万", "亿"}

	yuan := int(amount)
	jiao := int(amount*10) % 10
	fen := int(amount*100) % 10

	if yuan == 0 && jiao == 0 && fen == 0 {
		return "零元整"
	}

	var result strings.Builder

	// 整数部分
	if yuan > 0 {
		s := fmt.Sprintf("%d", yuan)
		n := len(s)
		for i, ch := range s {
			d := int(ch - '0')
			pos := n - 1 - i
			bigIdx := pos / 4
			smallIdx := pos % 4

			if d != 0 {
				result.WriteString(digits[d])
				result.WriteString(units[smallIdx])
			} else {
				if smallIdx == 0 && bigIdx > 0 {
					result.WriteString(bigUnits[bigIdx])
				} else if result.Len() > 0 && !strings.HasSuffix(result.String(), "零") {
					result.WriteString("零")
				}
			}
			if smallIdx == 0 && bigIdx > 0 {
				result.WriteString(bigUnits[bigIdx])
			}
		}
		result.WriteString("元")
	}

	// 小数部分
	if jiao > 0 {
		result.WriteString(digits[jiao])
		result.WriteString("角")
	} else if fen > 0 {
		result.WriteString("零")
	}
	if fen > 0 {
		result.WriteString(digits[fen])
		result.WriteString("分")
	} else {
		result.WriteString("整")
	}

	return result.String()
}
