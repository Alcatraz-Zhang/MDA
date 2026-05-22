---
name: generate-pi-task
description: 根据 pipeline 定义生成或更新 MaaFramework Project Interface 任务 JSON 文件。用于创建、编辑或扩展 assets/tasks/ 下的任务文件，将 pipeline 节点映射为用户可配置的选项（switch/select/checkbox/input）并绑定 pipeline_override。触发词："generate task"、"PI task"、"pipeline option"、"interface task"。
license: MIT
compatibility: Designed for Codex
metadata:
    version: "1.1"
    project: MDA
---

## 功能说明

- 读取 pipeline JSON 文件，识别可切换节点（含 `enabled` 字段的节点）
- 分析节点关系，确定选项类型（switch/select/checkbox/input）
- 生成或更新 `assets/tasks/*.json` 文件，遵循 `interface_import.schema.json` 规范
- 更新 `assets/interface.json` 的 import 数组和 locale 文件

## 使用场景

- 为 pipeline 创建新的任务文件
- 向已有任务文件添加选项
- 用户提到"生成任务"、"PI 任务"、"pipeline 选项"、"interface 任务"
- 用户正在处理 pipeline 文件，需要生成对应的任务配置

## 工作流程

### 第一步：读取规范文件

从 `tools/schema/` 读取：

1. **`interface_import.schema.json`** — 任务文件的 schema（顶层结构：`task[]`、`option{}`、`preset[]`）
2. **`interface.schema.json`** — 主 PI schema（import/group/controller/resource 上下文）
3. **`pipeline.schema.json`** — Pipeline 节点 schema（`enabled`、`next`、识别字段）

### 第二步：读取 Interface 上下文

1. 读取 `assets/interface.json`
2. 读取 `import[]` 数组中的所有文件 — 收集已有的任务名和选项 key，避免冲突
3. 记录 `group[]` 定义

### 第三步：确定目标文件

1. 当前活动文件在 `import[]` 中 → 编辑该文件
2. 当前活动文件是 pipeline JSON → 从目录名推导 pipeline 名称，检查 `assets/tasks/{Name}.json` 是否存在
3. 无上下文 → 询问用户要为哪个 pipeline 生成
4. 新建文件 → 同时在 `interface.json` 的 `import[]` 中添加条目

### 第四步：读取 Pipeline 文件

1. 读取 `assets/resource/pipeline/{PipelineName}/` 下所有 `.json` 文件
2. 提取每个节点的：name、`enabled`、`next[]`、`desc`
3. 构建节点图：通过 `next[]` 建立父→子映射

### 第五步：确定选项类型

详见[选项类型决策参考](references/option-types.md)。

快速判断规则：

- 单个含 `enabled` 的节点 → **switch**（开关）
- 同一父节点下互斥的 `enabled` 子节点 → **select**（单选）
- 多个可同时启用的独立项目 → **checkbox**（多选）
- 选择某个 case 后才显示的子选项 → **嵌套选项**（通过 `case.option[]`）
- 可配置的参数 → **input**（输入框），配合 `pipeline_type`

### 第六步：生成任务 JSON

严格遵循 `interface_import.schema.json`：

```json
{
    "task": [
        {
            "name": "PipelineName",
            "label": "$task.PipelineName.label",
            "entry": "PipelineName",
            "description": "$task.PipelineName.description",
            "option": ["OptionKey1"],
            "group": ["daily"]
        }
    ],
    "option": {
        "OptionKey1": {
            "type": "switch",
            "label": "$option.OptionKey1.label",
            "cases": [
                {"name": "Yes", "pipeline_override": {"TargetNode": {"enabled": true}}},
                {"name": "No", "pipeline_override": {"TargetNode": {"enabled": false}}}
            ]
        }
    }
}
```

### 第七步：更新 Locale 文件

向 `assets/locales/interface/zh_cn.json` 和 `en_us.json` 添加 i18n 键：

- `task.{Name}.label` / `task.{Name}.description`
- `option.{OptionKey}.label`
- `option.{OptionKey}.{CaseName}`（用于 checkbox/select 的 case）
- 使用 pipeline 节点的 `desc` 字段作为中文描述的来源

### 第八步：验证

- [ ] JSON 语法合法
- [ ] 所有 `task[].option[]` 的 key 在 `option` 对象中存在
- [ ] 所有 `pipeline_override` 的目标节点在 pipeline 文件中存在
- [ ] 所有嵌套 `case.option[]` 引用能正确解析
- [ ] 新文件已添加到 `interface.json` 的 `import[]`（如果是新建）
- [ ] Locale 键已添加（zh_cn 和 en_us）

## 注意事项

- **`interface_import.schema.json`** 才是任务文件的 schema，不是 `interface.schema.json`
- **`entry`** 必须与根 pipeline 节点名完全一致（区分大小写）
- **`interface.json` 中的导入路径** 相对于 `interface.json` 的位置：`"tasks/Arena.json"`，而非 `"assets/tasks/Arena.json"`
- **`$` 前缀** 是所有用户可见字符串的必需前缀 — 绝不硬编码中英文文本
- **`enabled: false`** 表示选项让用户将其开启；`enabled: true` 表示让用户将其关闭
- **switch 的 cases**：必须恰好 2 个，名称为 "Yes" 和 "No"
- **select 的 cases**：每个 case 需要启用自身节点，同时禁用同级兄弟节点
- **checkbox 的 cases**：每个 case 只启用自身节点（不禁用兄弟节点）
- **部分已有文件 `No` 排在 `Yes` 前面** — 当默认值为关闭时这是有意为之，遵循 pipeline 的 `enabled` 默认值顺序
