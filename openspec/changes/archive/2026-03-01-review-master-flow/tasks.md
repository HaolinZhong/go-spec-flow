## 1. HTML 模板 — Group 节点描述展示

- [x] 1.1 `internal/review/templates/review.html`：`showCode()` 函数改为 — 节点无 code 但有 description 时，展示 description 卡片（而非 "No code available"）

## 2. Skill 指令 — 主动线编排

- [x] 2.1 `skills/gsf-review.md` Step 3：重写编排指令，要求 AI 构建单一主动线 root → 步骤节点 → 函数节点的三层结构
- [x] 2.2 `skills/gsf-review.md` Step 3：更新 JSON 示例为三层嵌套格式
- [x] 2.3 `skills/gsf-review.md` Step 3：添加小包简化规则（<5 个函数可跳过中间层）

## 3. 同步与构建

- [x] 3.1 同步 `internal/cmd/embed_data/skills/gsf-review.md`
- [x] 3.2 `go build` 重新构建 gsf 二进制
- [x] 3.3 `gsf init` 安装更新后的 skill

## 4. 验证

- [x] 4.1 `go test ./...` 全部通过
- [x] 4.2 `/gsf:review` 自举验证：确认生成的 HTML 有主动线结构，group 节点展示 description
