# 设计文档 — 金额大写显示字段

## Context

资产租赁与催缴管理系统（Go + Vue3）的合同创建页面（NewContract.vue）已有月租金/年租金双向同步机制（watch + ignoreNext 防循环）。现需新增金额中文大写展示功能，方便用户在创建合同时直观确认金额，同时支持在 .docx 模板中使用大写金额占位符。

当前 SYSTEM_AUTO_FIELDS 包含：contractId, totalReceivable, totalReceived, deposit, status, notes。

后端 buildReplaceValues（template.go:470-529）使用 `map[string]string` 注册占位符，支持 `${key}` 格式替换。

## Goals / Non-Goals

**Goals:**
- 实现 `toChineseAmount(num: number): string` 纯函数，将数字转为中文大写金额
- 在 SYSTEM_AUTO_FIELDS 注册 5 个大写字段：monthlyRentCN, yearlyRentCN, totalReceivableCN, totalReceivedCN, depositCN
- 在合同详情表单中以只读方式展示大写金额（紧跟在对应金额字段下方）
- 在 getSystemAutoValue 中处理大写字段的值生成
- 在后端 buildReplaceValues 中注册大写占位符，支持 .docx 模板使用

**Non-Goals:**
- 不修改现有月租金/年租金双向同步逻辑（已正确实现）
- 不修改后端数据库 schema（CN 字段是纯展示，不持久化）
- 不支持超大数字（超过万亿）的边界处理

## Decisions

### D1: toChineseAmount 函数位置
**决策**: 放在独立文件 `frontend/src/utils/chineseAmount.ts` 中。
**理由**: 便于独立测试和复用，符合项目 utils 目录的组织方式。

### D2: 小数处理策略
**决策**: 使用"角""分"表示小数部分（如 12345.50 → "壹万贰仟叁佰肆拾伍元伍角"）。
**理由**: 符合中文大写金额的正式书写规范（银行/财务标准）。

### D3: 零值和空值策略
**决策**:
- 0 → "零元整"
- null/undefined → ""（空字符串，不显示）
**理由**: 零值有明确含义需要展示，空值表示未填写不应显示。

### D4: UI 渲染方式
**决策**: 在金额字段下方自动追加大写提示文本（不依赖模板 activeFieldsList 配置）。
**理由**: 避免要求模板管理员手动添加 CN 字段到 activeFieldsList，提升用户体验。同时 getSystemAutoValue 也支持 CN 字段，如果模板管理员主动将 CN 字段加入 activeFieldsList，也能正常显示。

### D5: 后端占位符注册
**决策**: 在 buildReplaceValues 中显式注册 5 个 CN 占位符，调用 Go 侧的中文大写转换函数。
**理由**: .docx 模板需要使用 `${monthlyRentCN}` 等占位符，必须在后端注册才能正确替换。

## Risks / Trade-offs

- **[精度风险]** JavaScript 浮点数精度问题可能导致小数转换不准确 → 使用 Math.round 和 toFixed 控制精度
- **[前后端一致性]** 前端 toChineseAmount 和后端 Go 实现可能产生不同结果 → 两端实现保持逻辑一致
- **[模板兼容性]** 现有模板可能无意中包含 `${monthlyRentCN}` 等占位符 → 在 buildReplaceValues 中注册后，未使用的占位符不影响生成
