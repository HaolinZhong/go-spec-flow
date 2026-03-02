## 1. 修改 extractFuncDiff

- [x] 1.1 重写 hunk 内行遍历逻辑：追踪 newLine，逐行判断是否在 [lineStart, lineEnd] 范围内
- [x] 1.2 `+` 行和 ` ` 行：newLine 在范围内时保留，范围外时跳过
- [x] 1.3 `-` 行：保留位于第一个保留行和最后一个保留行之间的
- [x] 1.4 裁剪后无保留行的 hunk 整体跳过（不输出 `@@` header）

## 2. 测试

- [x] 2.1 单元测试：单个 hunk 跨越函数边界，验证裁剪后只保留函数范围内的行
- [x] 2.2 单元测试：函数完全在 hunk 内部，验证只保留函数相关的 diff 行
- [x] 2.3 单元测试：函数无变更（hunk 不与函数重叠），验证返回空
- [x] 2.4 `go build` + `go test ./...`

## 3. 验证

- [x] 3.1 E2E：`gsf review --commit HEAD --json` 检查函数 diff 大小合理
- [x] 3.2 E2E：HTML review 中 Diff 视图显示正确
