package docx

import (
	"math"
	"strings"
)

var (
	cnDigits       = [10]string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}
	cnUnits        = [4]string{"", "拾", "佰", "仟"}
	cnSectionUnits = [3]string{"", "万", "亿"}
)

func convertSection(n int) string {
	if n == 0 {
		return ""
	}
	var b strings.Builder
	hasZero := false
	s := []rune(intToString(n))
	for i, ch := range s {
		digit := int(ch - '0')
		unitIdx := len(s) - 1 - i
		if digit == 0 {
			hasZero = true
		} else {
			if hasZero {
				b.WriteString("零")
				hasZero = false
			}
			b.WriteString(cnDigits[digit] + cnUnits[unitIdx])
		}
	}
	return b.String()
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	var b strings.Builder
	for n > 0 {
		b.WriteByte(byte('0' + n%10))
		n /= 10
	}
	// reverse
	s := []byte(b.String())
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return string(s)
}

// ToChineseAmount 将浮点数金额转为中文大写金额
// 如 12000.00 → "壹万贰仟元整"，12345.50 → "壹万贰仟叁佰肆拾伍元伍角"
func ToChineseAmount(n float64) string {
	// 精确到分
	rounded := math.Round(n*100) / 100
	if rounded == 0 {
		return "零元整"
	}

	isNegative := rounded < 0
	absNum := math.Abs(rounded)

	intPart := int(absNum)
	decPart := int(math.Round((absNum-float64(intPart))*100)) % 100
	jiao := decPart / 10
	fen := decPart % 10

	var b strings.Builder

	if isNegative {
		b.WriteString("负")
	}

	if intPart == 0 {
		// 纯小数
		if jiao > 0 {
			b.WriteString(cnDigits[jiao] + "角")
		}
		if fen > 0 {
			b.WriteString(cnDigits[fen] + "分")
		}
	} else {
		// 整数部分按4位分段（个、万、亿）
		var sections []int
		rem := intPart
		for rem > 0 {
			sections = append(sections, rem%10000)
			rem /= 10000
		}

		for i := len(sections) - 1; i >= 0; i-- {
			section := sections[i]
			sectionStr := convertSection(section)
			if sectionStr != "" {
				b.WriteString(sectionStr + cnSectionUnits[i])
				if i > 0 && sections[i-1] > 0 && sections[i-1] < 1000 {
					b.WriteString("零")
				}
			} else if i > 0 && sections[i-1] > 0 {
				b.WriteString("零")
			}
		}

		b.WriteString("元")

		if jiao == 0 && fen == 0 {
			b.WriteString("整")
		} else {
			if jiao > 0 {
				b.WriteString(cnDigits[jiao] + "角")
			} else if fen > 0 {
				b.WriteString("零")
			}
			if fen > 0 {
				b.WriteString(cnDigits[fen] + "分")
			}
		}
	}

	return b.String()
}
