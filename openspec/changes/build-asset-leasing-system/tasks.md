## 1. Go 后端基础架构

- [ ] 1.1 初始化 Go module（`go mod init`），创建项目骨架：`cmd/server/main.go`、`internal/` 各子包空文件
- [ ] 1.2 实现配置加载模块：命令行参数解析（`--mode`, `--db-host`, `--db-name`），统一配置结构体
- [ ] 1.3 实现 domain 层实体定义：Asset、Tenant、Contract、Payment、Receipt、ReceiptBook、Template、User
- [ ] 1.4 实现 domain 层 Repository 接口：AssetRepo、TenantRepo、ContractRepo、PaymentRepo、ReceiptRepo、ReceiptBookRepo、TemplateRepo、UserRepo
- [ ] 1.5 实现 SQLite repository：表创建 + 迁移（gorm auto-migrate），所有 repository 接口的 SQLite 实现
- [ ] 1.6 实现 PostgreSQL repository：所有 repository 接口的 PostgreSQL 实现
- [ ] 1.7 实现依赖注入：根据 config 选择 SQLite 或 PostgreSQL，初始化各层依赖

## 2. 核心计算引擎

- [ ] 2.1 实现 domain/calc/ContractCalc：TotalReceivable（应收总额）、UsedUpDate（钱用到哪天）、Arrears（欠费金额），纯函数 + 单测
- [ ] 2.2 实现 domain/calc/ArrearsCalc：五级分级判断（Level 1-5）、催缴清单生成，纯函数 + 单测
- [ ] 2.3 实现 domain/calc/ContractStatusCalc：合同状态自动判断（执行中/已缴全/欠费中/已到期），纯函数 + 单测

## 3. 认证与授权

- [ ] 3.1 实现 JWT 中间件：token 生成、验证、刷新，中间件链集成
- [ ] 3.2 实现 auth service：登录、获取当前用户信息
- [ ] 3.3 实现角色分权中间件：admin/operator 权限校验
- [ ] 3.4 实现用户管理 API（admin）：创建用户、删除用户、用户列表
- [ ] 3.5 首次启动创建默认 admin 账号

## 4. 资产与租户管理

- [ ] 4.1 实现 Asset service + transport：CRUD API（`GET/POST /api/assets`、`GET/PATCH /api/assets/:id`），搜索和分页
- [ ] 4.2 实现 Tenant service + transport：CRUD API（`GET/POST /api/tenants`、`GET/PATCH /api/tenants/:id`），搜索和分页
- [ ] 4.3 实现 ExtraFields 机制：系统设置定义字段列表 → UI 动态渲染，数据 JSON 存取
- [ ] 4.4 实现 Settings API：`GET/PUT /api/settings/fields`，扩展字段配置管理

## 5. 合同管理

- [ ] 5.1 实现 Contract service：创建合同（关联资产+租户，自动计算应收总额，支持人工微调）、查询合同、更新合同
- [ ] 5.2 实现 Contract transport：`GET/POST /api/contracts`、`GET/PATCH /api/contracts/:id`
- [ ] 5.3 实现合同状态自动流转：基于 TotalReceived vs TotalReceivable + 当前日期
- [ ] 5.4 实现模板上传与字段映射 API：`GET/POST /api/templates`、`GET/PATCH /api/templates/:id/mapping`

## 6. 收款与催缴

- [ ] 6.1 实现 Payment service：记录收款、更新合同 TotalReceived、触发计算引擎重新计算
- [ ] 6.2 实现 Payment transport：`GET /api/contracts/:id/payments`、`POST /api/contracts/:id/payments`
- [ ] 6.3 实现 Arrears service：每日自动生成五级催缴清单，查询催缴历史
- [ ] 6.4 实现 Arrears transport：`GET /api/arrears/lists`、`GET /api/arrears/history/:id`

## 7. 打印模块

- [ ] 7.1 实现合同 PDF 生成：加载模板 .docx → 字段映射替换占位符 → 渲染 PDF
- [ ] 7.2 实现身份证复印件 PDF：按 85.6mm×54mm 缩放居中 A4 纸，加"仅用于租赁合同备案"标注
- [ ] 7.3 实现三联收据 PDF：A4 竖版三等分（存根联/收据联/记账联），裁切线分隔
- [ ] 7.4 实现 ReceiptBook 管理：收据本 CRUD API，打印自动取号递增
- [ ] 7.5 实现 Print transport：`POST /api/print/contract/:id`、`POST /api/print/receipt/:id`

## 8. Vue 3 前端搭建

- [ ] 8.1 初始化 Vue 3 项目（Vite + Composition API + Pinia + Vue Router），配置构建为内嵌静态文件
- [ ] 8.2 搭建 Apple 设计系统基础：全局样式变量（色系、毛玻璃、圆角、间距、过渡、字体）、基础组件（按钮、输入框、卡片、步骤指示器）
- [ ] 8.3 实现认证模块：登录页、Token 管理、路由守卫、API 拦截器
- [ ] 8.4 实现布局框架：侧边导航 + 主内容区，10 个页面路由配置

## 9. 三大入口页面

- [ ] 9.1 实现「签新合同」页面：分步表单（选资产→录租户→定合同→预览打印），资产/租户下拉搜索+新建，OCR 占位
- [ ] 9.2 实现「收租金」页面：合同搜索 + 卡片列表 + 收款弹窗，显示"还差多少"，收款后一键打印收据
- [ ] 9.3 实现「催缴清单」页面：五级 Tab 切换，按紧急程度着色，每行附带建议动作

## 10. 后台管理页面

- [ ] 10.1 实现资产管理页面：列表+搜索+新建/编辑弹窗，历史租赁记录
- [ ] 10.2 实现租户管理页面：列表+搜索+新建/编辑弹窗，历史合同记录
- [ ] 10.3 实现合同管理页面：合同列表+状态筛选+详情（含收款记录），模板管理
- [ ] 10.4 实现收据本管理页面：列表+新建+状态显示
- [ ] 10.5 实现用户管理页面（admin）：用户列表+新建+删除
- [ ] 10.6 实现系统设置页面：扩展字段配置

## 11. 部署与集成

- [ ] 11.1 实现 Go embed 嵌入 Vue 构建产物，单一二进制运行
- [ ] 11.2 实现数据迁移工具：SQLite → PostgreSQL
- [ ] 11.3 编写 README 和启动脚本
- [ ] 11.4 集成测试：三大入口完整流程端到端验证

## 12. 前端构建配置

- [ ] 12.1 配置 Vite build：输出到 Go embed 目标目录，确保刷新不 404
- [ ] 12.2 Go embed.FS 嵌入 dist/，静态文件服务 + SPA fallback（所有非 API 路由返回 index.html）
