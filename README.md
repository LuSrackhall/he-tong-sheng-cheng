# 资产租赁与催缴管理系统

村/社区资产（商铺、车位、摊位、设备）租赁合同管理、租金收缴与欠费催缴一体化系统。

## 技术栈

- **后端:** Go + Gin + GORM
- **前端:** Vue 3 (Composition API) + TypeScript + Pinia + Vue Router
- **数据库:** SQLite（本地单文件）/ PostgreSQL（服务器）
- **构建:** Vite → `go:embed` 单一二进制部署

## 快速启动

```bash
# 1. 构建前端（输出到 cmd/server/dist/）
cd frontend && npm install && npm run build && cd ..

# 2. 编译并启动服务（SQLite 模式）
go build -o server ./cmd/server && ./server

# 服务运行在 http://localhost:8080
# 默认管理员账号: admin / admin123
```

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--mode` | `sqlite` | 数据库模式：`sqlite` 或 `postgres` |
| `--db-host` | `localhost` | PostgreSQL 主机地址 |
| `--db-name` | `asset_leasing` | PostgreSQL 数据库名 |
| `--port` | `8080` | 服务端口 |

环境变量 `JWT_SECRET` 用于 JWT 签名密钥。

## PostgreSQL 模式

```bash
./server --mode postgres --db-host 192.168.1.100 --db-name asset_leasing
```

## 功能概览

### 三大入口
- **签新合同** — 分步表单：选资产 → 录租户 → 定合同 → 预览
- **收租金** — 合同搜索 + 收款弹窗 + 还差多少
- **催缴清单** — 五级分级（应缴预警/近期提醒/逾期催收/到期预警/欠费追缴）

### 后台管理
- 资产管理、租户管理、合同管理
- 收据本管理、用户管理（管理员/操作员角色分权）
- 系统设置（合同模板字段映射）

## 项目结构

```
cmd/server/main.go          # 入口 + 路由 + go:embed
internal/
  config/                    # CLI 参数解析
  di/                        # 依赖注入
  domain/                    # 实体 + Repository 接口
    calc/                    # 纯函数计算引擎（已测）
  repository/
    sqlite/                   # SQLite 实现
    postgres/                 # PostgreSQL 实现
  transport/
    handler/                 # HTTP handlers
    middleware/               # JWT 认证 + SPA fallback
frontend/
  src/
    views/                   # 10 个 Vue 页面
    api/                     # Axios 封装 + 接口类型
    stores/                  # Pinia 状态管理
    router/                  # 路由配置
    styles/                  # Apple 设计系统
```

## 开发

```bash
# 后端
go test ./...               # 运行测试
go build ./...              # 检查编译

# 前端
cd frontend
npm run dev                 # Vite 开发服务器
npx vue-tsc --noEmit       # TypeScript 类型检查
npm run build               # 生产构建
```
