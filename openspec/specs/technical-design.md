# 技术架构设计

## 一、技术栈

| 维度 | 选择 |
|------|------|
| 后端语言 | Go |
| 前端框架 | Vue 3 (Composition API + Pinia + Vue Router) |
| UI 风格 | Apple 设计系统 (SF 色系、毛玻璃、克制阴影) |
| 架构形态 | 单一二进制 (Go 内嵌 Vue 静态文件) |
| 数据库 | SQLite (本地) / PostgreSQL (服务端)，启动配置切换 |
| HTTP 框架 | Gin |
| 认证 | JWT，角色分权 (admin / operator) |
| 打印 | 后端生成 PDF (合同、身份证复印件、三联收据) |
| OCR | 暂不实现，预留接口，后续可接入 PaddleOCR |

---

## 二、Go 后端分层架构

```
cmd/server (入口 / 配置加载 / 依赖注入)
    │
    ▼
transport (HTTP Handler / 中间件)
    │
    ▼
service (业务逻辑层)
    │
    ▼
domain ← repository (接口在 domain，实现在 repository/{sqlite,pg})
```

### 依赖方向

cmd → transport → service → domain ← repository

domain 不依赖任何外层。

### 各层职责

| 层 | 职责 |
|------|------|
| `cmd/server` | 程序入口，配置加载，依赖注入 |
| `transport` | HTTP Handler (Gin)，参数绑定 → 调用 service → 返回响应。中间件链: JWT / Logger / Recovery |
| `service` | 业务逻辑编排：合同服务、收款服务、催缴服务、资产/租户服务、打印服务、认证服务、模板服务。纯 Go 逻辑，不依赖 HTTP 框架 |
| `repository` | 数据持久化。接口定义在 domain 层，SQLite 和 PostgreSQL 分别实现，启动时根据 config 注入 |
| `domain` | 实体定义 (Contract, Asset, Tenant, Payment, Receipt)、值对象 (Money, DateRange, ArrearsLevel)、计算引擎 (ContractCalc, ArrearsCalc)、Repository 接口 |

### 计算引擎 (domain/calc/)

纯函数，无副作用，100% 可单测覆盖。

- **ContractCalc**：应收总额、钱用到哪天、欠费金额
- **ArrearsCalc**：五级分级判断、催缴清单生成

### 模板引擎

合同模板映射和占位符解析在 service 层，独立于打印模块。

---

## 三、核心数据模型

```
Asset {
  ID, Name, MonthlyRent, Group/Tag, ExtraFields(JSON)
}

Tenant {
  ID, Name, IDNumber, Gender(男/女/未知), Phone, Address, ExtraFields(JSON)
}

Contract {
  ID, AssetID(FK), TenantID(FK), StartDate, EndDate,
  MonthlyRent, TotalReceivable, TotalReceived,
  Status(执行中/已缴全/欠费中/已到期)
}

Payment {
  ID, ContractID(FK), Amount, PaidAt
}

Receipt {
  ID, ReceiptBookID, SequenceNum, ContractID, Amount, PrintedAt
}

ReceiptBook {
  ID, Prefix, StartNum, CurrentNum, TotalPages, Status(使用中/已用完/已作废)
}

Template {
  ID, Name, FilePath, FieldMapping(JSON)
}

User {
  ID, Username, PasswordHash, Role(admin/operator)
}
```

### ExtraFields 机制

资产和租户的非核心字段通过用户自定义扩展：
- 系统设置中定义扩展字段列表（如租户额外字段 `["职业", "紧急联系人"]`）
- UI 根据配置动态渲染输入控件
- 数据存储在 JSON 字段中
- 租户字段中系统建议项（姓名、身份证号、住址、手机号、称谓）通过 UI 提示，由用户自行抉择

### 关键设计点

- **Contract.TotalReceivable**：签合同时由开始日、结束日、月租金算出（可人工微调，覆盖首月折半等线下商定）
- **Contract.TotalReceived**：每次收款后累加，TotalReceived ≥ TotalReceivable → 已缴全
- **Payment**：每笔收款独立记录，不做收款周期假设
- **Receipt**：收据连号按本管理，ReceiptBookID 标识收据本，SequenceNum 标识本内序号，打印时自动递增
- **Template.FieldMapping**：用户定义"模板占位符 → 系统字段"映射

---

## 四、核心计算逻辑

### 应收总额

```
TotalReceivable(startDate, endDate, monthlyRent):
  整月数 = 日期跨度的完整自然月数
  剩余天数 = 不满整月的天数
  return 整月数 × monthlyRent + 剩余天数 × (monthlyRent / 30)
  结果可人工微调（覆盖首月折半等线下商定）
```

### 钱用到哪天

```
UsedUpDate(startDate, totalReceived, monthlyRent):
  N 整月 = totalReceived / monthlyRent (整数除)
  R 余数 = totalReceived % monthlyRent
  结果 = 起始日 + N 个自然月 (按日历月，自动适应 28/29/30/31 天)
  if R > 0:
    日租金 = monthlyRent / 30
    余数天数 = ceil(R / 日租金)
    结果 += 余数天数
  return min(结果, endDate)
```

### 欠费判断

```
Arrears(contract, totalReceived):
  return max(0, TotalReceivable - totalReceived)
```

单一标准：已收总额 < 应收总额。不预设月付/季付/年付。

### 五级催缴分级

| 等级 | 名称 | 条件 |
|------|------|------|
| 一级 | 应缴预警 | usedUpDate - now ≤ 30天 且 > 7天 |
| 二级 | 近期应缴提醒 | usedUpDate - now ≤ 7天 且 ≥ 0天 |
| 三级 | 逾期未缴催收 | now > usedUpDate 且 now ≤ endDate |
| 四级 | 到期预警 | endDate - now ≤ 30天 且未缴全 |
| 五级 | 已到期欠费追缴 | now > endDate 且未缴全 |

优先级: 取最高匹配等级。一个合同只出现在一张清单上。

---

## 五、API 设计

统一响应格式：`{"code": 0, "data": {...}, "message": "ok"}`

### 认证
```
POST   /api/auth/login
GET    /api/auth/me
```

### 资产
```
GET    /api/assets
POST   /api/assets
GET    /api/assets/:id
PATCH  /api/assets/:id
```

### 租户
```
GET    /api/tenants
POST   /api/tenants
GET    /api/tenants/:id
PATCH  /api/tenants/:id
```

### 合同
```
GET    /api/contracts
POST   /api/contracts
GET    /api/contracts/:id
PATCH  /api/contracts/:id
GET    /api/contracts/:id/payments
POST   /api/contracts/:id/payments
```

### 催缴
```
GET    /api/arrears/lists
GET    /api/arrears/history/:id
```

### 模板
```
GET    /api/templates
POST   /api/templates
GET    /api/templates/:id
PATCH  /api/templates/:id/mapping
```

### 打印
```
POST   /api/print/contract/:id
POST   /api/print/receipt/:id
GET    /api/receipt-books
POST   /api/receipt-books
```

### 设置
```
GET    /api/settings/fields
PUT    /api/settings/fields
```

### 用户管理 (admin)
```
GET    /api/users
POST   /api/users
DELETE /api/users/:id
```

---

## 六、打印模块

所有打印由后端生成 PDF，保证尺寸精度和跨平台一致性。

### 合同 PDF
1. 查询合同数据 (Contract + Asset + Tenant)
2. 查询模板文件 (用户上传的 .docx)
3. 查询字段映射，执行占位符替换
4. 渲染为 PDF
5. 返回 PDF 二进制流

### 身份证复印件
1. 查询租户身份证图片
2. 按原件尺寸缩放 (85.6mm × 54mm)
3. A4 纸居中排版，标注"仅用于租赁合同备案"
4. 返回 PDF

### 三联收据

收据连号按本管理：
- ReceiptBook 管理收据本编号、前缀、起止号码、当前已用号码
- 打印时自动取号，CurrentNum 递增
- A4 竖版三等分排版（每份约 99mm × 210mm）
- 每联内容相同但标注不同："存根联" / "收据联" / "记账联"
- 联与联之间裁切线分隔

---

## 七、前端架构

```
web/src/
├── views/          # 页面
│   ├── NewContract.vue      # 签新合同 (分步表单)
│   ├── CollectRent.vue      # 收租金 (搜索 + 卡片 + 弹窗)
│   ├── ArrearsList.vue      # 催缴清单 (五级 Tab)
│   ├── AssetManage.vue      # 资产管理
│   ├── TenantManage.vue     # 租户管理
│   ├── ContractManage.vue   # 合同管理
│   ├── TemplateManage.vue   # 模板管理
│   ├── ReceiptBooks.vue     # 收据本管理
│   ├── UserManage.vue       # 用户管理
│   └── Settings.vue         # 系统设置
├── components/     # 通用组件
├── stores/         # Pinia stores
├── api/            # Axios 封装
└── router/         # 路由配置
```

### Apple 设计系统

| 要素 | 规格 |
|------|------|
| 背景色 | 浅灰 `#F5F5F7` |
| 卡片 | 毛玻璃 `rgba(255,255,255,0.72)` + `0 2px 12px rgba(0,0,0,0.06)` |
| 主色 | 蓝 `#007AFF` |
| 圆角 | 大卡片 16px，按钮 8px，输入框 10px |
| 字体 | 系统默认中文，标题 600 字重，正文 400 |
| 间距 | 8px 基准网格，常用 16/24/32px |
| 过渡 | `cubic-bezier(0.25, 0.1, 0.25, 1)`，0.2s |
| 图标 | 线性风格，1.5px 描边 |

### 三大入口页面

1. **签新合同**：分步表单 (选资产 → 录租户 → 定合同 → 预览打印)，步骤指示器顶部
2. **收租金**：搜索条 + 合同卡片列表 + 收款弹窗，金额大字展示
3. **催缴清单**：五级 Tab 切换，按紧急程度着色，每行附带建议动作

---

## 八、部署形态

### 本地模式
```
./server --mode=local
# SQLite 自动创建在 ./data/
# Vue 前端内嵌，访问 http://localhost:8080
# 零外部依赖，数据文件可拷贝备份
```

### 服务端模式
```
./server --mode=server --db-host=pg.internal --db-name=rent_system
# PostgreSQL，支持多人多地同时访问
# Nginx 反代前端 + Go API
```

### 数据迁移
```
./server --mode=migrate \
  --from=sqlite://data/rent.db \
  --to=postgres://pg.internal/rent_system
# 支持从单机升级到联网
```

---

## 九、项目目录结构

```
he-tong-sheng-cheng/
├── cmd/server/main.go
├── internal/
│   ├── transport/ (handler, middleware, router)
│   ├── service/   (contract, payment, arrears, asset, tenant, template, print, auth)
│   ├── repository/ (接口 + sqlite/ + postgres/)
│   └── domain/     (entity, calc/, repository 接口)
├── web/            (Vue 3 前端)
├── openspec/       (config, specs, changes)
├── data/           (本地 SQLite 数据目录)
├── AGENTS.md
├── go.mod
└── go.sum
```

---

## 十、关键约束

- 合同模板由用户自行上传，系统不预设合同格式
- 不预设缴费方式 (月付/季付/年付)，仅以总额判断欠费
- 首月折半等特殊商定，通过调整合同应收总额处理，系统不做特殊逻辑
- 资产字段跟随合同模板走，系统核心计算仅依赖租期和租金
- 架构预留多资产类型扩展
