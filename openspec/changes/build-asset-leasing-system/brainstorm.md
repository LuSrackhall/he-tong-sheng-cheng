## Design Summary

一个以「合同」为核心的通用资产租赁与催收管理系统。第一切入场景为社区/村委商铺租赁，架构不绑定「房屋」，可扩展至车位、摊位、设备等任何可出租资产。核心解决签合同、收租金、催欠款三个环节，系统在后台自动完成记录、计算和提醒，不改变人的工作习惯。

技术选型：Go 后端（Gin 框架）分层架构 + Vue 3 前端（Composition API + Pinia + Vue Router）+ SQLite（本地单机）/ PostgreSQL（服务端多人）双数据库，Apple 设计系统，后端生成所有 PDF。

五大核心理念：业务即录入（无痛冷启动）、合同是唯一锚点、收款即记账（总额判欠费）、五级递进催缴、规范与灵活并存。

## Alternatives Considered

### 方案 A：全前端 SPA + BaaS（Supabase/Appwrite）

- **做法**：Vue 3 前端直接对接 BaaS，用 Edge Functions 处理计算逻辑，数据存托管数据库
- **优点**：开发快，免运维，自带认证和存储
- **缺点**：离线场景不可用（村委会网络不稳定）；"钱用到哪天"等核心计算逻辑写在 Edge Functions 里测试困难；数据主权不受控；长期成本不可控
- **为何未采用**：社区/村委场景网络条件参差不齐，离线可用是硬需求；核心计算逻辑需要 100% 单测覆盖，纯 Go 更合适

### 方案 B：Go 后端 + React 前端

- **做法**：Go 后端 + React（或 Next.js）前端，前后端分离部署
- **优点**：React 生态更成熟，组件库更丰富
- **缺点**：团队对 Vue 更熟悉；React 需要更多样板代码；Apple 风格在 Vue 下更容易实现（更少的 CSS-in-JS 心智负担）
- **为何未采用**：Vue 3 Composition API 与 Apple 极简设计理念更契合，且用户明确偏好 Vue

### 方案 C：Go 后端 + HTMX + 模板渲染

- **做法**：Go 模板渲染 HTML，HTMX 做局部刷新，最小化 JavaScript
- **优点**：极简，无前端构建步骤，包体小
- **缺点**：复杂交互（分步表单、动态搜索、五级 Tab 切换）实现困难；Apple 毛玻璃等精致 UI 效果需要大量 CSS hack；交互体验不如 SPA
- **为何未采用**：签新合同分步表单和催缴清单五级 Tab 需要复杂前端交互，HTMX 在此场景下开发体验和用户体验均不如 Vue SPA

## Agreed Approach

**Go 后端分层架构 + Vue 3 前端 + 单一二进制部署**

- Go 后端：cmd → transport → service → domain ← repository，纯函数计算引擎（ContractCalc / ArrearsCalc）在 domain/calc/，100% 可单测覆盖
- Vue 3 前端：Composition API + Pinia + Vue Router，Apple 设计系统（#F5F5F7 背景、毛玻璃卡片、#007AFF 主色、8px 基准网格）
- 数据库：Repository 接口 + SQLite / PostgreSQL 双实现，启动配置切换
- 部署：Go 内嵌 Vue 静态文件，单一二进制分发；支持本地模式（SQLite）和服务端模式（PostgreSQL + Nginx）
- 打印：后端生成 PDF（合同、身份证复印件 85.6mm×54mm、三联收据 A4 三等分）
- 认证：JWT，角色分权（admin / operator）

## Key Decisions

1. **合同是唯一锚点**：所有计算以合同为独立单元，不跨合同关联
2. **总额判欠费**：不预设月付/季付/年付，仅以 TotalReceived < TotalReceivable 判断
3. **五级催缴**：一级应缴预警(≤30天) → 二级近期应缴(≤7天) → 三级逾期未缴 → 四级到期预警(距到期≤30天) → 五级已到期欠费追缴，一个合同只出现在一张清单上
4. **"钱用到哪天"算法**：整月按日历月（自动适应 28/29/30/31 天），余数按日租金（monthlyRent/30）计算
5. **首月折半等商定**：通过人工微调 Contract.TotalReceivable 处理，系统不做特殊逻辑
6. **ExtraFields JSON 机制**：资产和租户的非核心字段通过用户自定义扩展字段动态渲染
7. **收据连号按本管理**：ReceiptBook 管理收据本编号、起止号码，打印时自动取号递增
8. **合同模板**：用户自行上传 .docx，定义占位符→系统字段映射
9. **OCR 暂不实现**：先手动输入身份证信息，预留 OCR 接口后续接入 PaddleOCR
10. **第一个切入场景**：社区/村委商铺租赁，架构不绑定房屋

## Open Questions

- 无。所有关键技术决策已在前期设计讨论中达成共识，详见 `openspec/specs/technical-design.md`。
