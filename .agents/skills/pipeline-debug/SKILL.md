---
name: pipeline-debug
description: 调试和优化 MaaFramework pipeline JSON 片段。对照 schema 验证配置，检测结构性问题（缺失引用、循环依赖、命名与代码不匹配），识别常见错误，并提供优化建议。用于调试 pipeline 执行、验证配置或提升性能。触发词："debug pipeline"、"validate pipeline"、"optimize pipeline"、"pipeline error"。
license: MIT
compatibility: Designed for Codex
metadata:
    version: "1.1"
    project: MDA
    author: MDA Team
allowed-tools: Read Grep Glob
---

## 功能说明

- 对照 `pipeline.schema.json` 及相关 schema 验证 pipeline JSON
- 检测结构性问题：缺失引用、循环依赖、孤立节点
- **通过命名规范分析节点角色** — 从节点名推断其功能
- **检测命名与代码的不匹配**（如节点名为 "Click" 但缺少 action）
- 提供性能和可维护性的优化建议
- 生成修正后的 pipeline 片段
- **允许增删节点** 以实现正确的 pipeline 行为

## 使用场景

- Pipeline JSON 运行不符合预期
- 部署前验证 pipeline 配置
- 优化 pipeline 性能或降低复杂度
- 调试 pipeline 执行问题
- 根据命名理解节点功能
- 验证节点名是否与实际行为匹配
- 用户提到"debug pipeline"、"validate pipeline"、"optimize pipeline"、"pipeline error"

## 工作流程

### 第一步：读取规范文件

从项目 `tools/schema/` 目录读取：

1. **`pipeline.schema.json`** — Pipeline 节点 schema（识别/动作类型、节点属性）
2. **`interface_import.schema.json`** — 任务文件 schema（pipeline_override 上下文）
3. **`interface.schema.json`** — 主 PI schema（resource/controller 上下文）
4. **`custom.recognition.schema.json`** — 自定义识别扩展
5. **`custom.action.schema.json`** — 自定义动作扩展

### 第二步：分析 Pipeline 片段

1. 解析 JSON，识别所有节点及其属性
2. 构建节点图：通过 `next[]` 建立父→子映射
3. 提取关键配置：`recognition`、`action`、`enabled`、`next`、`interrupt`、`sub`、`on_error`

### 第三步：对照 Schema 验证

逐节点检查 `pipeline.schema.json`：

- 必需字段是否齐全
- 字段类型是否匹配
- 枚举值、数值范围、字符串模式是否合法
- 识别类型及参数是否正确
- 动作类型及参数是否正确

### 第四步：通过命名分析节点语义

使用 [Pipeline 节点命名规范](references/pipeline-node-naming.md) 理解节点角色：

1. **解析名称结构**：`<Domain><ActionOrObject><Role>`
2. **通过后缀识别类型**：
    - `Main` → 入口节点，组织后续节点
    - `Flow` → 编排节点，无直接识别/动作
    - `Enter<Page>` → 导航节点，点击进入某页面
    - `On<Page>Page` / `Visible` → 状态检测节点
    - `Click<Object>` / `Select<Object>` / `Claim<Object>` → 动作节点
    - `Confirm<Object>` → 确认弹窗处理
    - `Scroll<Direction>` / `Swipe<Object>` → 滚动动作
    - `End` / `EndTask` → 终止节点
    - `Entered` → 成功哨兵，确认导航完成
3. **验证域一致性**：同一模块应使用相同域前缀
4. **检测命名与代码的不匹配**：
    - `Click<Object>` 但无 `action` → 可能有误
    - `Visible` 但有 `action: Click` → 应为 `Click<Object>`
    - `Flow` 但有识别参数 → 应为纯编排节点
    - `Enter<Page>` 但无 `next` 重试 → 缺少成功哨兵

### 第五步：分析节点关系

- **引用验证**：所有 `next[]`、`interrupt[]`、`sub[]` 的目标必须存在
- **循环检测**：识别 `next` 链中的无限循环
- **孤立检测**：找到从未被引用的节点
- **入口点**：未被任何其他节点引用的根节点
- **死胡同**：无 `next` 且非终止动作的节点

### 第六步：识别常见问题

详见[调试规则参考](references/debug-rules.md)。

快速检查项：

- 缺少识别 / 识别类型错误
- 缺少动作 / ROI 不当 / 阈值问题
- 性能瓶颈 / 逻辑错误

### 第七步：生成优化建议

1. **性能**：减少不必要的识别、优化 ROI、调整阈值
2. **可维护性**：简化结构、降低嵌套深度
3. **可靠性**：添加错误处理、提升识别精度
4. **可读性**：添加 `desc` 文档、使用有意义的节点名

### 第八步：提供修正方案

如发现问题：生成修正后的 JSON，解释每处改动，提供前后对比。

响应格式参见[输出格式参考](references/output-format.md)。

## 注意事项

- **V1/V2 模式**：部分节点使用 `recognition`（V1），部分使用 `type`（V2）— 两者均合法
- **`next` 可选**：无 `next` 的节点是合法的终止节点
- **`enabled` 默认为 true**：节点默认启用，除非显式设为 `false`
- **`interrupt` 与 `sub` 的区别**：`interrupt` 暂停当前执行，`sub` 并行运行
- **ROI 格式**：`[x, y, w, h]` 数组或引用前一节点的字符串
- **模板路径**：相对于 `image/` 文件夹，而非项目根目录
- **OCR expected**：支持正则表达式，不仅是字面字符串
- **自定义节点**：可能包含 base schema 中没有的额外属性
- **节点命名**：必须遵循 PascalCase 的 Domain + ActionOrObject + Role 格式

## 参考资料

- [Pipeline Schema](../../../tools/schema/pipeline.schema.json)
- [Pipeline 节点命名规范](references/pipeline-node-naming.md)
- [调试规则](references/debug-rules.md)
- [常见模式](references/common-patterns.md)
- [输出格式](references/output-format.md)
