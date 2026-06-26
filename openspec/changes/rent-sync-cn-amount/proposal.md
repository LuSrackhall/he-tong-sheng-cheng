## Why

合同创建页面需要在填写金额时即时展示中文大写金额，方便用户确认大额数字无误，同时支持在 .docx 合同模板中使用大写金额占位符（如 ${monthlyRentCN}），满足正式合同文书的格式要求。

## What Changes

1. **新增 `toChineseAmount` 工具函数** — 将数字转为中文大写金额（如 12000 → "壹万贰仟元整"），放在 `frontend/src/utils/chineseAmount.ts`
2. **扩展 SYSTEM_AUTO_FIELDS** — 注册 5 个大写字段：monthlyRentCN, yearlyRentCN, totalReceivableCN, totalReceivedCN, depositCN
3. **合同详情表单展示** — 在金额字段下方自动追加大写金额的只读提示文本
4. **getSystemAutoValue 扩展** — 处理 CN 字段的值生成
5. **后端占位符注册** — 在 `buildReplaceValues` 中注册 5 个 CN 占位符，支持 .docx 模板使用

## Capabilities

### New Capabilities
- `chinese-amount-display`: 金额中文大写显示能力，包括前端 toChineseAmount 函数、SYSTEM_AUTO_FIELDS 注册、UI 只读展示、后端占位符注册

### Modified Capabilities
- `template-driven-contract-form`: 扩展 getSystemAutoValue 以支持 CN 后缀字段的值生成

## Impact

- **前端**: 新增 `frontend/src/utils/chineseAmount.ts`，修改 `NewContract.vue`（SYSTEM_AUTO_FIELDS、getSystemAutoValue、模板渲染区域）
- **后端**: 修改 `internal/transport/handler/template.go` 的 `buildReplaceValues` 函数，新增 Go 侧中文大写转换函数
- **兼容性**: 纯新增功能，不影响现有字段和模板
