## 1. 前端 toChineseAmount 工具函数

- [x] 1.1 创建 `frontend/src/utils/chineseAmount.ts`，实现 `toChineseAmount(num: number): string` 纯函数（支持整数、小数角/分、零值、空值）
- [x] 1.2 确保函数通过边界测试：0 → "零元整"，null → ""，12000 → "壹万贰仟元整"，12345.50 → "壹万贰仟叁佰肆拾伍元伍角"，100.05 → "壹佰元零伍分"

## 2. 前端 NewContract.vue 扩展

- [x] 2.1 在 SYSTEM_AUTO_FIELDS 中添加 5 个 CN 字段：monthlyRentCN, yearlyRentCN, totalReceivableCN, totalReceivedCN, depositCN
- [x] 2.2 在 getSystemAutoValue 中添加 CN 字段的 case 分支，调用 toChineseAmount 转换对应金额值
- [x] 2.3 在 fieldLabels 中添加 CN 字段的中文标签
- [x] 2.4 在模板驱动表单的金额字段（monthlyRent, yearlyRent, totalReceivable, deposit）下方追加大写金额只读提示（灰色小字 `<p>` 标签）
- [x] 2.5 在非模板驱动的 fallback 表单（step 3 else 分支）的金额字段下方追加同样的大写提示

## 3. 后端 Go 实现

- [x] 3.1 创建 `internal/docx/chinese_amount.go`，实现 `ToChineseAmount(n float64) string` 函数
- [x] 3.2 修改 `internal/transport/handler/template.go` 的 `buildReplaceValues` 函数，注册 5 个 CN 占位符

## 4. 验证

- [x] 4.1 运行 `go test ./... -count=1` 确保后端测试通过
- [x] 4.2 运行 `vue-tsc --noEmit` 确保前端类型检查通过
- [x] 4.3 运行 `npm run build` 确保前端构建通过

---

## Post-Implementation Workflow

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run `myspec-verify` to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, notify main session and wait for merge signal
4. **Archive**: Run `myspec-merge` on main
5. **Cleanup**: `git worktree remove .worktrees/change/<name>`
