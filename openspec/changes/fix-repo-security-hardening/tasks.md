## 1. Gitignore 增强

- [ ] 1.1 在根目录 `.gitignore` 中添加 `.env` 和 `.env.*` 规则
- [ ] 1.2 在根目录 `.gitignore` 中添加 `*.log` 规则
- [ ] 1.3 在根目录 `.gitignore` 中添加 `.DS_Store` 规则
- [ ] 1.4 在根目录 `.gitignore` 中添加 `__pycache__/` 和 `*.pyc` 规则
- [ ] 1.5 在根目录 `.gitignore` 中添加 `tmp/` 规则
- [ ] 1.6 在根目录 `.gitignore` 中添加 `node_modules/` 规则

## 2. 构建产物清理

- [ ] 2.1 执行 `git rm --cached -r cmd/server/dist/` 移除 dist 目录的 git 跟踪
- [ ] 2.2 验证 `git ls-files cmd/server/dist/` 返回空
- [ ] 2.3 验证本地 `cmd/server/dist/` 目录文件仍在磁盘上

## 3. JWT 密钥强制配置

- [ ] 3.1 修改 `internal/config/config.go`：将 `envDefault("JWT_SECRET", "asset-leasing-secret-change-me")` 改为读取环境变量并检测空值
- [ ] 3.2 在 `Load()` 中添加空值检测：若 `JWT_SECRET` 未设置或为空，`log.Fatalf` 退出并输出明确错误信息
- [ ] 3.3 验证 `go build ./...` 通过

## 4. 验证与文档

- [ ] 4.1 运行 `go build ./...` 确认编译通过
- [ ] 4.2 运行 `npm run build` 确认前端构建通过
- [ ] 4.3 测试未设置 `JWT_SECRET` 时服务器拒绝启动
- [ ] 4.4 测试设置 `JWT_SECRET` 后服务器正常启动
