# 第九章：Agent 技能系统（Skills）

欢迎来到第九章！在前面的章节中，我们已经构建了一个功能完善的 AI Agent，具备了工具调用、上下文管理、记忆系统、RAG 能力和安全防护机制。

本章介绍 Agent 技能系统（Skills）——**本质上是一段描述性文字，指导模型在不同场景下如何使用工具**。

> **核心设计理念**：
> - **技能即提示**：技能不是代码，而是精心设计的提示词，告诉模型在特定场景下该怎么做
> - **渐进式加载**：只在 system prompt 中注入技能元数据（名称+描述），完整内容按需加载，避免占用过多 token
>
> 例如："代码审查"技能不会改变 Agent 的能力，而是告诉模型：审查代码时应该检查哪些方面、按照什么步骤、输出什么格式。模型仍然使用相同的工具（read、write 等），只是更有章法。

---

## 🎯 你将学到什么

1. **技能抽象**：将常见任务模式抽象为可复用的技能描述
2. **技能路由**：LLM 自动识别任务类型并选择合适的技能
3. **Markdown 技能文件**：使用 YAML front matter + Markdown 正文定义技能
4. **按需加载**：技能元数据注入 system prompt，完整内容按需加载

---

## 🛠 核心功能

### 1. 技能系统设计

**核心概念**：

> 💡 **技能不是新功能，而是使用说明书**
>
> Agent 的能力来自工具（bash、read、write 等），技能只是告诉模型：
> - 面对这类任务时，该用什么步骤？
> - 该检查哪些方面？
> - 该输出什么格式？
>
> 就像厨师的刀不变，但切菜、切片、切丁有不同的手法。

**渐进式加载设计** - 解决 token 消耗问题：

```
初始化时（每次对话）：
  System Prompt: {skills}
  ↓
  - **Code Review**: Review code for bugs...
  - **Debug**: Diagnose and fix issues...
  - **Refactor**: Improve code structure...
  ↓
  Token 成本：~50 tokens（只有名称+描述）

运行时（仅当需要时）：
  LLM 决定需要 Code Review 技能
  ↓
  调用 load_skill(name="code-review")
  ↓
  返回完整技能内容（~500 tokens）
  ↓
  LLM 按照指导执行任务
```

**对比：如果不分渐进式加载**
- 10 个技能 × 500 tokens = 5000 tokens 每次请求
- 渐进式：50 tokens + 按需 500 tokens
- 节省：90%+ 的 token 成本

**实现机制**：

**问题**：Agent 面对不同类型的任务时，往往需要遵循特定的模式和最佳实践。

**解决方案**：将常见任务模式抽象为技能（Skills），通过两层机制让 Agent 使用：

**技能发现与元数据注入**：
- Agent 初始化时扫描 `.babyagent/skills/` 目录
- 解析每个技能的 front matter（name, description）
- 将技能列表注入到 system prompt 的 `{skills}` 占位符

**按需加载完整内容**：
- LLM 分析任务后调用 `load_skill` 工具
- 工具读取完整的技能 Markdown 内容
- 技能指导注入到对话中，引导 LLM 执行

### 2. 技能文件格式

技能使用标准的 Markdown + YAML front matter 格式：

```
.babyagent/skills/<skill-id>/SKILL.md
---
name: Code Review
description: Review code for bugs, style issues, and best practices
---

# Code Review Skill

## Checklist
- [ ] Check for potential bugs and edge cases
- [ ] Verify error handling
- [ ] Assess code readability and maintainability
- [ ] Suggest performance optimizations if applicable

## Process
1. Read the target files completely
2. Analyze each component systematically
3. Provide specific, actionable feedback
4. Suggest code improvements with examples
```

**Front Matter 字段**：
- `name`：技能显示名称（用于 system prompt）
- `description`：技能描述（帮助 LLM 理解何时使用）

**正文内容**：详细的步骤、检查清单、最佳实践等

### 3. 技能加载流程

```
┌─────────────────────────────────────────────────┐
│  Agent 初始化                                     │
│  ┌──────────────────────────────────────────┐   │
│  │ 1. 扫描 .babyagent/skills/               │   │
│  │ 2. 解析每个技能的 front matter           │   │
│  │ 3. 注入到 system prompt {skills}         │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│  System Prompt                                   │
│  ## Available Skills                             │
│  - **Code Review**: Review code for bugs...      │
│  - **Debug**: Diagnose and fix issues...         │
│  - **Refactor**: Improve code structure...       │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│  用户请求: "帮我审查这段代码"                     │
│  ↓                                                │
│  LLM: 任务匹配 "Code Review" 技能                │
│  ↓                                                │
│  LLM: 调用 load_skill(name="code-review")        │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│  load_skill 工具                                 │
│  • 读取 .babyagent/skills/code-review/SKILL.md  │
│  • 返回完整技能内容                              │
│  • LLM 根据指导执行任务                          │
└─────────────────────────────────────────────────┘
```

---

## 📖 代码结构速览

### Skill 包 (`ch09/skill/`)

**`skill.go`** - 技能管理器：
- `Manager`：技能管理器
- `LoadAll()`：扫描并加载所有技能元数据
- `FormatForPrompt()`：格式化为 system prompt 片段

**`markdown.go`** - Markdown 解析：
- `LoadSkill(id)`：加载技能的元数据和正文
- 解析 YAML front matter
- 提取 Markdown 正文内容

### Tool (`ch09/tool/load_skill.go`)

**`LoadSkillTool`** - 技能加载工具：
- `ToolName()`: 返回 "load_skill"
- `Execute()`: 调用 `skill.LoadSkill()` 获取完整内容

### Context (`ch09/context/engine.go`)

**初始化时**：
- `NewContextEngine()` 内部创建 `skill.Manager`
- 调用 `LoadAll()` 加载技能元数据
- `BuildSystemPrompt()` 替换 `{skills}` 占位符

---

## 💡 使用示例

### 创建技能

```bash
# 创建技能目录
mkdir -p .babyagent/skills/code-review

# 创建技能文件
cat > .babyagent/skills/code-review/SKILL.md << 'EOF'
---
name: Code Review
description: Review code for bugs, style issues, and best practices
---

# Code Review Skill

## Review Checklist
- **Correctness**: Check for bugs, edge cases, error handling
- **Style**: Verify naming, formatting, code organization
- **Performance**: Identify potential optimizations
- **Maintainability**: Assess clarity and extensibility

## Process
1. Read all related files completely
2. Understand the intended functionality
3. Review systematically using checklist
4. Provide specific, actionable feedback
5. Suggest improvements with code examples

## Output Format
\```
## Summary
[Brief overview of findings]

## Issues Found
### Critical
- [Issue description]

### Minor
- [Issue description]

## Suggestions
- [Improvement suggestion]
\```
EOF
```

### 运行 Agent

```bash
go run ./ch09/tui
```

### 对话示例

```
你: 帮我审查一下 main.go 的代码

Agent: 我会使用 Code Review 技能来审查代码。

[调用 load_skill(name="code-review")]

# Skill: Code Review

[完整的技能指导内容...]

---

现在我来审查 main.go：

## Summary
[审查结果...]
```

---

## 🔧 技能文件约定

### 目录结构
```
.babyagent/
└── skills/
    ├── code-review/
    │   └── SKILL.md
    ├── debug/
    │   └── SKILL.md
    └── refactor/
        └── SKILL.md
```

### 文件命名
- 文件夹名 = **技能 ID**（用于 load_skill 调用）
- 文件名固定为 `SKILL.md`

### Front Matter 格式
```yaml
---
name: 显示名称
description: 技能描述，帮助 LLM 理解使用场景
---
```

### 内容建议
1. **步骤清晰**：使用编号列表说明执行步骤
2. **检查清单**：提供 checklist 确保完整性
3. **输出格式**：定义标准输出格式
4. **最佳实践**：包含领域最佳实践
5. **示例**：提供具体示例

---

## 🚀 扩展方向

### 技能组合
- 一个技能可以调用其他技能
- 实现技能层次结构（基础技能 → 高级技能）

### 技能分享
- 技能市场/仓库
- 社区贡献技能模板

### 动态技能
- 根据项目类型自动选择技能集
- 从远程仓库拉取技能

---

## 📚 相关资源

- [Agent Skills](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview)