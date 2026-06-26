## Context

合同创建页面（NewContract.vue）使用 template-driven 渲染，activeFieldsList 驱动表单字段展示。SYSTEM_AUTO_FIELDS 定义了自动生成的字段集合，getSystemAutoValue 负责计算这些字段的值。后端 buildReplaceValues（template.go）组装 ${key} 占位符的替换值 map。

月租金/年租金双向同步已有完整实现（watch + ignoreNext 防循环），无需修改。

## Goals / Non-Goals

**Goals:**
- 实现 toChineseAmount 纯函数，支持整数和小数（角/分），放在独立文件中便于测试
- 在前端 SYSTEM_AUTO_FIELDS 注册 5 个 CN 字段，getSystemAutoValue 中调用 toChineseAmount 生成值
- 在 UI 金额字段下方自动追加大写提示（不依赖 activeFieldsList 配置）
- 在后端 buildReplaceValues 中注册 CN 占位符，使用 Go 侧独立的中文大写转换函数

**Non-Goals:**
- 不修改现有月租金/年租金同步逻辑
- 不修改后端数据库 schema
- 不支持超过万亿的数字

## Decisions

### D1: toChineseAmount 算法设计
- 整数部分按 4 位一组分段（个、万、亿）
- 每段内按千百十个位转换
- 零的处理：连续零只读一个，段尾零不读
- 小数部分：角位和分位，0 角 0 分读"整"
- 精度：传入前先 Math.round(num * 100) / 100 确保两位小数

### D2: 前端 UI 渲染策略
- 在模板驱动表单中，每个金额字段（monthlyRent, yearlyRent, totalReceivable, deposit）的 form-group 内部，紧跟 input 之后追加一个 `<p>` 标签显示大写金额
- 使用 computed 或直接在模板中调用 toChineseAmount
- 样式：`font-size: 0.75rem; color: var(--color-text-tertiary); margin-top: 4px;`
- 非模板驱动的 fallback 表单（step 3 else 分支）也需要同样的大写提示

### D3: 后端 Go 实现
- 新增 `internal/docx/chinese_amount.go` 文件
- 导出 `ToChineseAmount(n float64) string` 函数
- 在 template.go 的 buildReplaceValues 中调用，注册 5 个 CN 占位符

## Risks / Trade-offs

- **[精度风险]** JS 浮点数 0.1 + 0.2 !== 0.3 → toChineseAmount 内部先 round 到两位小数
- **[前后端一致性]** JS/Go 两套实现需保持逻辑一致 → 使用相同的算法（4位分段法）
- **[性能]** 大写转换在每次金额输入时触发 → 纯计算函数，无性能问题

## Open Questions

无。
