## Context

资产租赁管理系统（Go + Gin + GORM + Vue 3）存在多处代码质量问题和 UX 优化空间。参见 brainstorm-spec.md 和 proposal.md 获取完整背景。

当前仓库结构中，sqlite/ 和 postgres/ 各自维护完整的 repo 实现，但由于 GORM 已抽象 SQL 方言，这些实现完全相同。AuthMiddleware 在 main.go 和 auth.go 中各创建一次。服务器使用 `r.Run()` 启动，无优雅关停。

## Goals / Non-Goals

**Goals:**
- 消除仓库代码重复，建立 common/ 包作为 repo 实现的单一事实来源
- 通过信号处理实现优雅关停，确保 Docker/systemd 部署下的可靠重启
- 统一 API 错误消息为中文，提升终端用户体验
- 新增 Dashboard 概览数据

**Non-Goals:**
- 不改变 GORM 版本或引入新的 ORM
- 不修改数据库 schema 或迁移策略
- 不重构认证/授权机制
- 不引入 Naive UI 等第三方组件库

## Decisions

### 1. 仓库去重：common/ 包方案

**选择**：新建 `internal/repository/common/` 包，将所有 repo struct 定义、构造函数和方法实现放入其中。

**实现细节**：
```
internal/repository/
  common/
    repos.go     # 所有 repo struct + 构造函数 + 方法实现
  sqlite/
    setup.go     # 仅 Setup() — 导入 common 并使用其构造函数
  postgres/
    setup.go     # 仅 Setup() — 导入 common 并使用其构造函数
```

`common/repos.go` 中的 struct 定义：
```go
type AssetRepo struct{ DB *gorm.DB }  // 注意：大写 DB，跨包访问
type TenantRepo struct{ DB *gorm.DB }
// ... 所有 9 个 repo
```

`sqlite/setup.go` 变更：
- 删除所有 struct 定义和方法实现
- 保留 `Setup()` 函数
- 返回值改为 `(*gorm.DB, error)` — 构造函数在 `di/` 包中调用 common 的

**di/ 包适配**：需要更新 `di/` 包中的 repo 构造，改为调用 `common.NewAssetRepo(db)` 等。

**为什么不使用构建标签**：GORM 已消除 SQL 方言差异，构建标签增加复杂度而无收益。

### 2. AuthMiddleware 复用

**选择**：`NewAuthHandler` 接收 `*middleware.AuthMiddleware` 实例。

```go
// 之前
func NewAuthHandler(userRepo domain.UserRepo, jwtSecret string) *AuthHandler

// 之后
func NewAuthHandler(userRepo domain.UserRepo, authmw *middleware.AuthMiddleware) *AuthHandler
```

main.go 中只创建一个 `authmw`，传给 AuthHandler 和路由中间件。

### 3. 优雅关停实现

**选择**：标准 `http.Server` + `os/signal` 方案。

```go
srv := &http.Server{
    Addr:    ":" + cfg.Port,
    Handler: r,
}

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("服务器启动失败: %v", err)
    }
}()

<-quit
log.Println("收到关闭信号，正在优雅关停...")
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Fatalf("服务器关闭失败: %v", err)
}
log.Println("服务器已关闭")
```

**backup.go 适配**：将 `os.Exit(0)` 替换为向 shutdown channel 发送信号。需要在 handler 中注入 shutdown 函数：
```go
type BackupHandler struct {
    db          *gorm.DB
    dbPath      string
    shutdownFn  func()  // 由 main.go 传入
}
```

### 4. Dashboard API

**端点**：`GET /api/dashboard/stats`（需认证）

**响应**：
```json
{
  "activeContracts": 12,
  "monthlyRevenue": 45000.00,
  "overdueContracts": 3,
  "newContractsThisMonth": 2
}
```

**查询方式**：
- `activeContracts`：`COUNT(*) FROM contracts WHERE status IN ('active', 'arrears')`
- `monthlyRevenue`：`SUM(amount) FROM payments WHERE paid_at >= 月初`
- `overdueContracts`：`COUNT(*) FROM contracts WHERE status = 'arrears'`
- `newContractsThisMonth`：`COUNT(*) FROM contracts WHERE created_at >= 月初`

需要在 domain 层添加 `DashboardStats` 结构体和独立的 `DashboardRepo` 接口，避免污染现有 repo 接口。

**选择独立 DashboardRepo 接口**：新建 `domain.DashboardRepo` 接口和 `common.DashboardRepo` 实现，包含 CountActive、MonthlyRevenue、CountOverdue、CountNewThisMonth 方法。保持 ContractRepo 和 PaymentRepo 接口不变。

### 5. 侧边栏分组

当前 App.vue 已有分组结构。调整：
- 新增 "概览" 入口（链接到 `/`，带首页图标）
- "日常操作" 重命名为 "业务管理"
- "数据管理" 重命名为 "基础数据"
- "系统设置" 重命名为 "系统管理"
- 将"合同管理"移入"业务管理"组

### 6. Home.vue Dashboard

- 默认路由从 `/new-contract` 改为 `/`
- 展示 4 个统计卡片（活跃合同、本月收款、逾期合同、本月新增）
- 使用现有 CSS 变量和卡片样式
- 加载中骨架屏 + 错误处理

## Risks / Trade-offs

1. **[风险] 仓库重构影响 di/ 包** → 缓解：同步更新 di 包中的 repo 构造调用，编译即验证
2. **[风险] 优雅关停注入 shutdownFn 到 handler** → 缓解：简单的函数注入，不影响现有 handler 接口
3. **[风险] Dashboard 查询可能需要新 repo 方法** → 缓解：先尝试用已有方法组合，不足时添加简单方法
4. **[取舍] 错误消息中文化 vs 国际化** → 选择中文化：系统面向中文用户，无需 i18n 框架
