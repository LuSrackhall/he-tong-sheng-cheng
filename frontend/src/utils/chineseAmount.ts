/**
 * 数字金额转中文大写金额
 * 如：12000 → "壹万贰仟元整"，12345.50 → "壹万贰仟叁佰肆拾伍元伍角"
 */

const DIGITS = ['零', '壹', '贰', '叁', '肆', '伍', '陆', '柒', '捌', '玖']
const UNITS = ['', '拾', '佰', '仟']
const SECTION_UNITS = ['', '万', '亿']

function convertSection(n: number): string {
  if (n === 0) return ''
  let result = ''
  let hasZero = false
  const str = String(n)
  const len = str.length
  for (let i = 0; i < len; i++) {
    const digit = parseInt(str[i], 10)
    const unitIndex = len - 1 - i
    if (digit === 0) {
      hasZero = true
    } else {
      if (hasZero) {
        result += '零'
        hasZero = false
      }
      result += DIGITS[digit] + UNITS[unitIndex]
    }
  }
  return result
}

export function toChineseAmount(num: number | null | undefined): string {
  if (num === null || num === undefined || isNaN(num)) return ''

  // 精确到分
  const rounded = Math.round(num * 100) / 100
  if (rounded === 0) return '零元整'

  const isNegative = rounded < 0
  const absNum = Math.abs(rounded)

  // 分离整数和小数部分
  const intPart = Math.floor(absNum)
  const decPart = Math.round((absNum - intPart) * 100)
  const jiao = Math.floor(decPart / 10)
  const fen = decPart % 10

  let result = ''

  if (isNegative) result += '负'

  if (intPart === 0) {
    // 纯小数
    if (jiao > 0) result += DIGITS[jiao] + '角'
    if (fen > 0) result += DIGITS[fen] + '分'
  } else {
    // 整数部分：按4位分段（个、万、亿）
    const sections: number[] = []
    let remaining = intPart
    while (remaining > 0) {
      sections.push(remaining % 10000)
      remaining = Math.floor(remaining / 10000)
    }

    for (let i = sections.length - 1; i >= 0; i--) {
      const section = sections[i]
      const sectionStr = convertSection(section)
      if (sectionStr) {
        result += sectionStr + SECTION_UNITS[i]
        // 如果下一个低位段存在且不足4位（<1000），需要补零
        if (i > 0 && sections[i - 1] > 0 && sections[i - 1] < 1000) {
          result += '零'
        }
      } else if (i > 0 && sections[i - 1] > 0 && sections[i - 1] < 1000) {
        // 中间段为0，后面段需要读零
        result += '零'
      }
    }

    result += '元'

    if (jiao === 0 && fen === 0) {
      result += '整'
    } else {
      if (jiao > 0) {
        result += DIGITS[jiao] + '角'
      } else if (fen > 0) {
        result += '零'
      }
      if (fen > 0) {
        result += DIGITS[fen] + '分'
      }
    }
  }

  return result
}
