## Context

仓库当前 `.gitignore` 仅包含 5 条规则，存在以下问题：

1. **构建产物已跟踪**：`cmd/server/dist/` 在 `.gitignore` 添加前已被 git 跟踪，4 个文件仍被 git 索引，历史中累计 200+ 个版本的编译 JS/CSS/HTML 文件
2. **gitignore 规则不完整**：缺少 `.env`、`*.log`、`.DS_Store`、`__pycache__/`、`*.pyc`、`tmp/`、`node_modules/` 等常见排除模式
3. **JWT 默认密钥硬编码**：`config.go:31` 中 `envDefault("JWT_SECRET", "asset-leasing-secret-change-me")` 提供了弱默认值

项目使用 Go + Gin 后端 + Vue 3 前端，通过 `go:embed` 将前端构建产物嵌入二进制文件。`cmd/server/dist/` 的构建产物被嵌入而非从文件系统读取，因此从 git 中移除它们不影响运行时行为（构建时会重新生成）。

## Goals / Non-Goals

**Goals:**
- 从 git 跟踪中移除 `cmd/server/dist/` 构建产物
- 增强 `.gitignore` 规则，防止未来意外提交敏感文件
- JWT 密钥强制通过环境变量配置，消除硬编码默认值
- 提供安全检查清单文档

**Non-Goals:**
- 不重写 git 历史（使用 `git rm --cached` 而非 `git filter-branch`）
- 不修改 `frontend/.gitignore`（已有完善规则）
- 不添加 pre-commit hooks 或 CI 安全扫描（可作为后续改进）

## Decisions

### 1. `git rm --cached` vs `git filter-branch`

**选择：** `git rm --cached`

**理由：** `filter-branch` 会重写整个 git 历史，需要强制推送，对协作造成破坏。`git rm --cached` 仅从索引中移除文件，下次提交后不再跟踪，历史中的旧版本保留但不影响当前和未来。

**替代方案：** `git filter-repo`（更安全的 filter-branch 替代品）—— 但需要额外安装，且重写历史的代价对本项目不值得。

### 2. JWT 默认值处理

**选择：** 默认值改为空字符串，`Load()` 中检测空值时 `log.Fatalf` 退出

**理由：** 强制要求显式配置，避免任何人在不知情的情况下使用弱密钥。开发环境可通过 `.env` 文件或 shell 配置设置。

**替代方案：** 自动生成随机密钥（每次启动变化）—— 会导致重启后所有 token 失效，对开发不友好。

### 3. `.gitignore` 增强范围

**选择：** 添加通用规则（`.env`、`*.log`、`.DS_Store`、`__pycache__/`、`*.pyc`、`tmp/`、`node_modules/`），不添加项目特定的测试数据规则

**理由：** 测试数据（SQLite 数据库、上传文件）已被现有 `*.db` 和 `uploads/` 规则覆盖。通用规则防止操作系统和开发工具产生的文件意外提交。

## Risks / Trade-offs

- **[Breaking Change]** 未设置 `JWT_SECRET` 的部署将无法启动 → 迁移计划中包含检查步骤，README 中添加说明
- **[dist 文件历史保留]** `git rm --cached` 不清除历史中的 dist 文件 → 仓库体积不会立即缩小，但未来的 `git clone --depth 1` 不会拉取它们
- **[.claude/ 目录]** 包含 OpenSpec 工具配置，非敏感数据 → 保持跟踪，不加入 gitignore
