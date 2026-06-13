## Why

安全审计发现仓库存在三个安全隐患：(1) 构建产物 `cmd/server/dist/` 在 `.gitignore` 添加前已被 git 跟踪，历史提交中累计 200+ 个版本的编译文件使仓库臃肿；(2) `.gitignore` 规则不完整，缺少 `.env`、`*.log`、`.DS_Store` 等常见排除模式，存在意外泄露敏感文件的风险；(3) JWT 默认密钥 `"asset-leasing-secret-change-me"` 硬编码在源码中，部署时若未设置环境变量将导致可被伪造的认证。

## What Changes

- 清除 `cmd/server/dist/` 的 git 跟踪（`git rm --cached`），并从历史中移除已跟踪的构建产物
- 增强根目录 `.gitignore`，添加 `.env`/`.env.*`、`*.log`、`.DS_Store`、`__pycache__/`、`*.pyc`、`tmp/`、`node_modules/` 等规则
- 修改 `internal/config/config.go`：JWT 默认密钥改为空字符串，启动时检测未设置 `JWT_SECRET` 环境变量则报错退出
- 添加 `docs/security-checklist.md` 记录安全审计结论和后续检查清单

## Capabilities

### New Capabilities
- `repo-security-hardening`: 仓库安全加固——gitignore 增强、构建产物清理、JWT 密钥强制配置

### Modified Capabilities

（无现有 capability 的需求变更）

## Impact

- `cmd/server/dist/` 从 git 跟踪中移除（不影响本地文件）
- `.gitignore` 规则变更影响后续所有提交
- `internal/config/config.go` 的 `Load()` 函数行为变更：未设置 `JWT_SECRET` 时将 fatal exit，**BREAKING** 现有不设置环境变量的开发/部署流程
- 需要所有部署环境确认已设置 `JWT_SECRET` 环境变量
