# review-trim-func-diff

裁剪 extractFuncDiff 输出，只保留函数行范围内的 diff 行。

## Requirements

### R1: 函数范围裁剪
- hunk 内的 `+` 行和 ` ` 上下文行，只在对应的 newLine 落在 [lineStart, lineEnd] 时保留
- `-` 行保留条件：位于第一个保留行和最后一个保留行之间

### R2: `@@` header 保留
- 裁剪后的 hunk 保留原始 `@@` header 不修改
- 如果裁剪后 hunk 内无任何保留行，整个 hunk 不输出

### R3: 接口不变
- `extractFuncDiff(fileDiff string, lineStart, lineEnd int) string` 签名不变
- 返回空字符串仍表示该函数无变更

## Acceptance Criteria

- 对一个 300B 代码的函数，附带的 funcDiff 大小应接近实际变更行数，不应包含函数范围外的大段上下文
- 同一文件内多个函数不再共享完全相同的大块 diff
- HTML review 中 Diff 视图显示的内容精确对应函数范围内的变更
- 现有测试通过，新增针对裁剪逻辑的单元测试
