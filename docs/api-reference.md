# API 参考文档

> 本文档列出系统所有 REST API 端点，供开发者或系统集成参考。

---

## 目录

- [基础信息](#基础信息)
- [认证](#认证)
- [健康检查](#健康检查)
- [资产接口](#资产接口)
- [租户接口](#租户接口)
- [合同接口](#合同接口)
- [收款接口](#收款接口)
- [催缴接口](#催缴接口)
- [模板接口](#模板接口)
- [收据本接口](#收据本接口)
- [用户管理接口](#用户管理接口管理员)
- [错误码说明](#错误码说明)

---

## 基础信息

- **基础路径**：`/api`
- **认证方式**：JWT Bearer Token
- **数据格式**：JSON（请求和响应均为 JSON，文件上传除外）
- **认证请求头**：

```
Authorization: Bearer <token>
```

---

## 认证

### 登录

```
POST /api/auth/login
```

**请求体：**

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**成功响应 (200)：**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin"
  }
}
```

**失败响应 (401)：**

```json
{
  "error": "Invalid credentials"
}
```

### 获取当前用户信息

```
GET /api/auth/me
```

**需要认证**：是

**成功响应 (200)：**

```json
{
  "id": 1,
  "username": "admin",
  "role": "admin"
}
```

---

## 健康检查

```
GET /api/health
```

**不需要认证**

**成功响应 (200)：**

```json
{
  "status": "ok"
}
```

---

## 资产接口

### 获取资产列表

```
GET /api/assets
```

**需要认证**：是

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `search` | string | 否 | 按资产名称搜索 |
| `type` | string | 否 | 按资产类型筛选（shop/parking/stall/equipment/other） |
| `offset` | int | 否 | 分页偏移量，默认 0 |
| `limit` | int | 否 | 每页数量，默认 20 |

**成功响应 (200)：**

```json
{
  "data": [
    {
      "id": 1,
      "name": "人民路128号商铺",
      "assetType": "shop",
      "description": "",
      "status": "idle",
      "extraFields": "",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

### 创建资产

```
POST /api/assets
```

**需要认证**：是

**请求体：**

```json
{
  "name": "人民路128号商铺",
  "assetType": "shop",
  "description": "一楼临街商铺",
  "extraFields": ""
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | **是** | 资产名称 |
| `assetType` | string | 否 | 资产类型，默认 `shop` |
| `description` | string | 否 | 描述 |
| `extraFields` | string | 否 | 扩展字段（JSON 文本） |

**成功响应 (201)：** 返回创建的资产对象

### 获取单个资产

```
GET /api/assets/:id
```

**需要认证**：是

**路径参数：** `id` — 资产 ID

**成功响应 (200)：** 返回资产对象

**失败响应 (404)：**

```json
{
  "error": "Asset not found"
}
```

### 更新资产

```
PATCH /api/assets/:id
```

**需要认证**：是

**请求体：** 与创建相同，所有字段均为可选（仅传需要修改的字段）

**成功响应 (200)：** 返回更新后的资产对象

---

## 租户接口

### 获取租户列表

```
GET /api/tenants
```

**需要认证**：是

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `search` | string | 否 | 按姓名或电话搜索 |
| `offset` | int | 否 | 分页偏移量，默认 0 |
| `limit` | int | 否 | 每页数量，默认 20 |

**成功响应 (200)：**

```json
{
  "data": [
    {
      "id": 1,
      "name": "张三",
      "phone": "13800138000",
      "idCard": "110101199001011234",
      "idCardImage": "",
      "extraFields": "",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

### 创建租户

```
POST /api/tenants
```

**需要认证**：是

**请求体：**

```json
{
  "name": "张三",
  "phone": "13800138000",
  "idCard": "110101199001011234",
  "idCardImage": "",
  "extraFields": ""
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | **是** | 租户姓名 |
| `phone` | string | 否 | 联系电话 |
| `idCard` | string | 否 | 身份证号 |
| `idCardImage` | string | 否 | 身份证图片路径 |
| `extraFields` | string | 否 | 扩展字段（JSON 文本） |

**成功响应 (201)：** 返回创建的租户对象

### 获取单个租户

```
GET /api/tenants/:id
```

**需要认证**：是

**成功响应 (200)：** 返回租户对象

### 更新租户

```
PATCH /api/tenants/:id
```

**需要认证**：是

**请求体：** 与创建相同，所有字段可选

**成功响应 (200)：** 返回更新后的租户对象

---

## 合同接口

### 获取合同列表

```
GET /api/contracts
```

**需要认证**：是

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `search` | string | 否 | 按租户名或资产名搜索 |
| `status` | string | 否 | 按状态筛选：`active`、`paidup`、`arrears`、`expired` |
| `offset` | int | 否 | 分页偏移量，默认 0 |
| `limit` | int | 否 | 每页数量，默认 20 |

**成功响应 (200)：**

```json
{
  "data": [
    {
      "id": 1,
      "assetId": 1,
      "asset": { "id": 1, "name": "人民路128号商铺", "assetType": "shop", ... },
      "tenantId": 1,
      "tenant": { "id": 1, "name": "张三", "phone": "13800138000", ... },
      "startDate": "2024-01-01T00:00:00Z",
      "endDate": "2025-01-01T00:00:00Z",
      "monthlyRent": 5000,
      "totalReceivable": 60000,
      "totalReceived": 55000,
      "deposit": 5000,
      "status": "arrears",
      "templateId": 1,
      "notes": "",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-06-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

### 创建合同

```
POST /api/contracts
```

**需要认证**：是

**请求体：**

```json
{
  "assetId": 1,
  "tenantId": 1,
  "startDate": "2024-01-01",
  "endDate": "2025-01-01",
  "monthlyRent": 5000,
  "totalReceivable": 60000,
  "deposit": 5000,
  "templateId": 1,
  "notes": "特殊约定"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `assetId` | uint | **是** | 资产 ID |
| `tenantId` | uint | **是** | 租户 ID |
| `startDate` | string | **是** | 开始日期，格式 `YYYY-MM-DD` |
| `endDate` | string | **是** | 结束日期，格式 `YYYY-MM-DD` |
| `monthlyRent` | float64 | **是** | 月租金 |
| `totalReceivable` | float64 | 否 | 应收总额，不传则自动计算 |
| `deposit` | float64 | 否 | 押金 |
| `templateId` | uint | 否 | 合同模板 ID |
| `notes` | string | 否 | 备注 |

**成功响应 (201)：** 返回创建的合同对象

**错误响应 (409)：**

```json
{
  "error": "该资产与租户在此时间段已有合同"
}
```

### 获取单个合同

```
GET /api/contracts/:id
```

**需要认证**：是

**成功响应 (200)：** 返回合同对象（含关联的资产和租户信息）

### 更新合同

```
PATCH /api/contracts/:id
```

**需要认证**：是

**请求体：** 可更新字段均为可选

```json
{
  "monthlyRent": 5500,
  "totalReceivable": 66000,
  "startDate": "2024-01-01",
  "endDate": "2025-06-01",
  "deposit": 5500,
  "notes": "新备注"
}
```

**成功响应 (200)：** 返回更新后的合同对象

---

## 收款接口

### 获取合同的收款记录

```
GET /api/contracts/:id/payments
```

**需要认证**：是

**路径参数：** `id` — 合同 ID

**成功响应 (200)：**

```json
[
  {
    "id": 1,
    "contractId": 1,
    "amount": 5000,
    "paidAt": "2024-02-01T00:00:00Z",
    "notes": "2月份租金",
    "createdAt": "2024-02-01T00:00:00Z"
  }
]
```

### 记录收款

```
POST /api/contracts/:id/payments
```

**需要认证**：是

**路径参数：** `id` — 合同 ID

**请求体：**

```json
{
  "amount": 5000,
  "paidAt": "2024-02-01",
  "notes": "2月份租金"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `amount` | float64 | **是** | 收款金额，必须大于 0 |
| `paidAt` | string | 否 | 收款日期，格式 `YYYY-MM-DD`，默认今天 |
| `notes` | string | 否 | 备注 |

**成功响应 (201)：**

```json
{
  "payment": {
    "id": 1,
    "contractId": 1,
    "amount": 5000,
    "paidAt": "2024-02-01T00:00:00Z",
    "notes": "2月份租金"
  },
  "shortfall": 5000
}
```

`shortfall` 表示本次收款后仍差多少金额。

---

## 催缴接口

### 获取催缴清单

```
GET /api/arrears
```

**需要认证**：是

**成功响应 (200)：**

```json
[
  {
    "id": 1,
    "asset": { "id": 1, "name": "人民路128号商铺", ... },
    "tenant": { "id": 1, "name": "张三", ... },
    "totalReceived": 55000,
    "totalReceivable": 60000,
    "usedUpDate": "2024-12-01",
    "endDate": "2025-01-01",
    "arrearsLevel": 3,
    "monthlyRent": 5000,
    "status": "arrears"
  }
]
```

| 字段 | 说明 |
|------|------|
| `usedUpDate` | 已付租金覆盖到的日期（"钱用到"） |
| `arrearsLevel` | 催缴等级（1-5，0 表示正常无需催缴） |

**催缴等级含义：**

| 等级 | 名称 | 触发条件 |
|------|------|----------|
| 1 | 应缴预警 | usedUpDate 在 30 天内到达 |
| 2 | 近期应缴提醒 | usedUpDate 在 7 天内到达 |
| 3 | 逾期未缴催收 | usedUpDate 已过，endDate 尚未到期 |
| 4 | 到期预警 | endDate 在 30 天内到期 |
| 5 | 已到期欠费追缴 | endDate 已过，仍有欠款 |

> 仅返回 arrearsLevel > 0 的合同。

---

## 模板接口

### 获取模板列表

```
GET /api/templates
```

**需要认证**：是

**成功响应 (200)：** 返回模板对象数组

### 创建模板

```
POST /api/templates
```

**需要认证**：是

**请求体：**

```json
{
  "name": "商铺租赁合同模板"
}
```

**成功响应 (201)：** 返回创建的模板对象

### 更新模板字段映射

```
PATCH /api/templates/:id
```

**需要认证**：是

**请求体：**

```json
{
  "fieldMap": "{\"startDate\":\"开始日期\",\"endDate\":\"结束日期\",\"monthlyRent\":\"月租金\"}",
  "activeFields": "{\"startDate\":true,\"endDate\":true,\"monthlyRent\":true}"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `fieldMap` | string | **是** | 字段映射 JSON（系统字段名 → Word 占位符名） |
| `activeFields` | string | 否 | 启用字段 JSON（字段名 → true/false） |

**必填字段映射：** `startDate`、`endDate`、`monthlyRent`、`tenantName`、`assetName`

**成功响应 (200)：** 返回更新后的模板对象

### 上传模板文件

```
POST /api/templates/:id/upload
```

**需要认证**：是

**Content-Type**：`multipart/form-data`

**表单字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | file | **是** | Word 文档（.docx） |

**成功响应 (200)：** 返回更新后的模板对象（`validated: true`）

**失败响应 (400)：**

```json
{
  "error": "Word 文件缺少以下已启用的占位符",
  "missingFields": ["startDate", "monthlyRent"]
}
```

### 删除模板

```
DELETE /api/templates/:id
```

**需要认证**：是

**成功响应 (200)：**

```json
{
  "message": "模板已删除"
}
```

**失败响应 (409)：**

```json
{
  "error": "该模板已被合同引用，无法删除"
}
```

---

## 合同导出接口

### 导出合同文件

```
POST /api/contracts/:id/export
```

**需要认证**：是

**说明：** 使用合同绑定的模板生成 Word 文件

**成功响应 (200)：**

```json
{
  "message": "Contract exported successfully",
  "downloadUrl": "/api/contracts/1/download",
  "filePath": "uploads/exports/contract_1.docx"
}
```

**错误响应 (400)：**

```json
{
  "error": "Contract has no template assigned"
}
```

### 下载合同文件

```
GET /api/contracts/:id/download
```

**需要认证**：是

**响应：** 返回 `.docx` 文件（Content-Disposition: attachment）

**错误响应 (404)：**

```json
{
  "error": "Exported file not found. Please export the contract first."
}
```

---

## 收据本接口

### 获取收据本列表

```
GET /api/receipt-books
```

**需要认证**：是

**成功响应 (200)：**

```json
{
  "data": [
    {
      "id": 1,
      "prefix": "SK-2024",
      "startNum": 1,
      "currentNum": 5,
      "totalPages": 100,
      "status": "active"
    }
  ],
  "total": 1
}
```

### 创建收据本

```
POST /api/receipt-books
```

**需要认证**：是

**请求体：**

```json
{
  "prefix": "SK-2024",
  "startNum": 1,
  "totalPages": 100
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `prefix` | string | **是** | 编号前缀 |
| `startNum` | int | 否 | 起始编号，默认 1 |
| `totalPages` | int | **是** | 总页数 |

**成功响应 (201)：** 返回创建的收据本对象

---

## 用户管理接口（管理员）

> 以下接口需要管理员（admin）角色才能访问。

### 获取用户列表

```
GET /api/admin/users
```

**需要认证**：是（管理员）

**成功响应 (200)：**

```json
[
  {
    "id": 1,
    "username": "admin",
    "role": "admin"
  },
  {
    "id": 2,
    "username": "operator1",
    "role": "operator"
  }
]
```

### 创建用户

```
POST /api/admin/users
```

**需要认证**：是（管理员）

**请求体：**

```json
{
  "username": "operator1",
  "password": "password123",
  "role": "operator"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | **是** | 用户名（不可重复） |
| `password` | string | **是** | 密码 |
| `role` | string | 否 | 角色，默认 `operator`（可选 `admin`） |

**成功响应 (201)：**

```json
{
  "id": 2,
  "username": "operator1",
  "role": "operator"
}
```

**错误响应 (409)：**

```json
{
  "error": "Username already exists"
}
```

### 删除用户

```
DELETE /api/admin/users/:id
```

**需要认证**：是（管理员）

**路径参数：** `id` — 用户 ID

**成功响应 (200)：**

```json
{
  "message": "User deleted"
}
```

**错误响应 (400)：**

```json
{
  "error": "不能删除自己的账号"
}
```

```json
{
  "error": "不能删除最后一个管理员"
}
```

---

## 错误码说明

| HTTP 状态码 | 含义 | 常见原因 |
|-------------|------|----------|
| 200 | 成功 | 请求正常完成 |
| 201 | 创建成功 | 资源创建完成 |
| 400 | 请求参数错误 | 缺少必填字段、格式不对 |
| 401 | 未认证 | 未登录或 Token 过期 |
| 403 | 无权限 | 非管理员访问管理接口 |
| 404 | 资源不存在 | ID 对应的记录不存在 |
| 409 | 冲突 | 重复数据、模板校验失败、合同时间冲突 |
| 500 | 服务器内部错误 | 数据库异常等 |

**通用错误响应格式：**

```json
{
  "error": "错误描述信息"
}
```
