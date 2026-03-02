## 1. 删除 gsf diff 命令及依赖代码

- [x] 1.1 删除 `internal/cmd/diff.go`
- [x] 1.2 删除 `internal/review/diff.go`、`mapper.go`、`extract.go`、`extract_test.go`
- [x] 1.3 删除 `skills/gsf-diff.md` 和 `internal/cmd/embed_data/skills/gsf-diff.md`
- [x] 1.4 确保 `go build ./...` 编译通过（处理可能的 import 引用）
- [x] 1.5 确保 `go test ./...` 全部通过

## 2. 重写 gsf:review skill

- [x] 2.1 重写 `skills/gsf-review.md`：基于 git diff + gsf trace/callers 的 review 流程
- [x] 2.2 同步更新 `internal/cmd/embed_data/skills/gsf-review.md`

## 3. 自举验证

- [x] 3.1 重新构建 gsf 二进制
- [x] 3.2 验证 `gsf diff` 命令已不存在
- [x] 3.3 验证 `gsf callers` 和 `gsf trace` 仍正常工作
- [x] 3.4 用新的 review skill 流程对本次变更做一次完整 review，确认 diff 完整、动线清晰
