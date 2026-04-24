# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BabyAgent is a Go-based AI Agent development tutorial that builds a Claude Code-like assistant from scratch. Each chapter (ch01-ch12) is a standalone module that progressively adds capabilities:

- **ch01-03**: Basic LLM interaction, tool calling, TUI visualization
- **ch04-06**: MCP protocol integration, context engineering, memory system
- **ch07**: Agentic RAG with vector search (pgvector)
- **ch08**: Sandbox/security with Docker isolation and tool confirmation
- **ch09**: Skills system for task-specific guidance
- **ch10+**: Web services, evaluation, observability (planned)

Chapters ch01-ch09 are complete. The current implementation combines all features into a full-featured agent.

## Common Commands

### Running Chapters
```bash
# Run individual chapters (each has a TUI interface)
go run ./ch03/main    # Chapter 3: TUI visualization
go run ./ch04/main    # Chapter 4: MCP integration
go run ./ch05/main    # Chapter 5: Context management
go run ./ch06/main    # Chapter 6: Memory system
go run ./ch08/main    # Chapter 8: Full agent with guardrails
go run ./ch09/main    # Chapter 9: Full agent with skills system (latest)
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for specific chapter
go test ./ch08/context/...
go test ./ch08/... -v
go test ./ch09/skill/... -v

# Run a single test
go test -v ./ch08/context/... -run TestTruncatePolicyApply
```

### Building
```bash
# Build a chapter's main executable
go build -o bin/ch09 ./ch09/main
```

## Configuration Requirements

Before running any chapter, set up the environment:

1. **Copy example configs:**
   ```bash
   cp .env.example .env
   cp config.json config.json
   ```

2. **Configure LLM providers** in `config.json`:
   - `front_model`: Main chat model (e.g., gpt-4o, deepseek-chat)
   - `back_model`: Cheaper model for summarization/memory updates
   - Both require: `base_url`, `api_key`, `model`, `context_window`

3. **Optional: Configure MCP servers** in `mcp-server.json` for extending agent capabilities

## Architecture Overview

### Core Components (ch08)

```
ch08/ or ch09/
├── agent.go          # Main Agent: orchestrates LLM calls, tool execution, streaming
├── context/
│   ├── engine.go     # ContextEngine: manages conversation history, applies policies
│   ├── policy.go     # Policy interface: truncate, offload, summarize strategies
│   ├── policy_*.go   # Policy implementations
│   └── share.go      # Token counting with tiktoken-go
├── memory/
│   ├── memory.go     # Memory interface: global + workspace memory
│   └── update.go     # LLM-driven memory extraction/update
├── storage/
│   ├── storage.go    # Storage interface for persistence
│   └── filesystem.go # File system storage implementation
├── tool/
│   ├── tool.go       # Tool interface
│   ├── bash.go       # Shell command execution
│   ├── docker_bash.go # Docker-isolated shell execution (ch08+)
│   ├── load_storage.go # Load offloaded content
│   └── load_skill.go  # Load skill content (ch09)
├── mcp.go            # MCP client integration (Model Context Protocol)
├── skill/            # Skills system (ch09 only)
│   ├── skill.go      # Skill manager
│   └── markdown.go   # Skill file parser
├── main/             # Main entry point
│   └── main.go       # TUI application entry
├── tui/              # Bubble Tea terminal UI
│   ├── tui.go        # TUI model and update logic
│   └── entry.go      # Log entry types and rendering
├── vo.go             # View objects for streaming messages to TUI
└── prompt.go         # System prompt template with {runtime}, {workspace_path}, {memory}, {skills}
```

### Agent Loop Flow

1. **User Input** → `Agent.RunStreaming(query, viewCh, confirmCh)`
2. **Context Build** → ContextEngine builds messages with system prompt (includes memory + skills for ch09)
3. **LLM Call** → Streaming response via openai-go SDK
4. **Tool Calls** → Check if confirmation needed → Show confirm UI or execute directly
5. **Confirmation** → User selects Allow/Reject/Always Allow (via `confirmCh`)
6. **Execute Tools** → Run native tools (bash, read, write, load_skill) or MCP tools
7. **ESC Handling** → User can cancel anytime; messages preserved, policies/memory skipped
8. **Normal End** → Apply truncate/offload/summarize policies
9. **Memory Update** → LLM extracts new facts from conversation
10. **TUI Display** → Show reasoning, content, tool calls, confirm dialog via MessageVO channel

### Skill Loading Flow (ch09)

1. **Agent Startup** → Skill manager scans `.babyagent/skills/` directory
2. **Metadata Extraction** → Parse YAML front matter (name, description) from each `SKILL.md`
3. **System Prompt** → Inject skill list into `{skills}` placeholder
4. **User Request** → LLM analyzes task and identifies relevant skill
5. **load_skill Call** → LLM calls `load_skill(name="<skill-id>")`
6. **Skill Content** → Tool reads full Markdown content and returns to LLM
7. **Task Execution** → LLM follows skill guidance to complete task

### Key Architectural Patterns

**Context Management (ch05)**:
- Messages wrapped with token counts (`messageWrap`)
- Policies applied when context usage exceeds thresholds
- Draft/commit pattern: `StartTurn()` → modify → `CommitTurn()` or `AbortTurn()`
- Policy event hooks notify TUI during execution

**Memory System (ch06)**:
- Two-level: Global (home dir) + Workspace (project dir)
- LLM-driven extraction: runs a cheaper model to extract facts
- Memory injected into system prompt via `{memory}` placeholder

**Tool System**:
- Native tools: `bash`, `read`, `write`, `edit`, `load_storage`, `load_skill`
- `bash`: Shell command execution with optional Docker isolation (ch08+)
- `docker_bash`: Docker-isolated shell execution (auto-detects Docker, falls back to bash)
- `load_storage`: Load offloaded message content from storage
- `load_skill`: Load full skill content for LLM guidance (ch09)
- MCP tools: Namespaced as `babyagent_mcp__<server>__<tool>`
- Tool results become `openai.ToolMessage` in conversation

**MCP Integration (ch04)**:
- Connects via stdio (subprocess) or HTTP (SSE)
- Tool discovery via `ListTools()`, execution via `CallTool()`
- Workspace variable `${workspaceFolder}` auto-replaced with current dir

**Tool Confirmation (ch08)**:
- `ToolConfirmConfig.RequireConfirmTools` maps tool names to confirmation requirement
- Before execution, check if tool requires confirmation AND not in `alwaysAllowTools`
- If needed, send `MessageTypeToolConfirm` to TUI and wait for `ConfirmationAction`
- Three actions: `ConfirmAllow`, `ConfirmReject` (terminates loop), `ConfirmAlwaysAllow`
- `alwaysAllowTools` tracks user's "always allow" choices for current session

**ESC Cancellation (ch08)**:
- TUI creates `context.WithCancel()` and passes to `Agent.RunStreaming()`
- User presses ESC → TUI calls `cancel()` → Agent detects `ctx.Done()`
- On cancel: `CommitTurn(..., skipPoliciesAndMemory=true)` preserves messages without running policies/memory
- Draft messages are saved to context, allowing user to continue from that point

**Skills System (ch09)**:
- Skills are descriptive prompts, not code - they guide LLM on how to use tools
- Progressive loading saves tokens: metadata (~50 tokens) at startup, full content (~500 tokens) on-demand
- Skill metadata (name + description) injected into `{skills}` placeholder in system prompt
- LLM calls `load_skill(name="<skill-id>")` when it needs full guidance
- Skill files in `.babyagent/skills/<skill-id>/SKILL.md` with YAML front matter
- Skill manager (`skill.Manager`) auto-discovers and loads skills at startup
- Skills can be shared across projects (workspace-level) or globally (home dir)

## Testing Patterns

Tests use lightweight fakes to avoid LLM API calls:

- **fakeSummarizer**: Returns pre-defined summaries for testing policy logic
- **fakeStorage**: In-memory key-value store for testing offload policy
- **buildEngine()**: Helper to create test engines with pre-populated messages

Example from `ch08/context/policy_test.go`:
```go
func TestOffloadPolicyApply_OffloadsLongToolMessagesOnly(t *testing.T) {
    st := &fakeStorage{store: map[string]string{}}
    p := NewOffloadPolicy(st, 0.8, 1, 10)
    engine := buildEngine([]shared.OpenAIMessage{...})
    result, err := p.Apply(context.Background(), engine)
    // assertions...
}
```

## Shared Utilities

The `shared/` package contains common code used across chapters:

- `client.go`: `NewLLMClient()` creates openai-go client with custom headers
- `config.go`: ModelConfig, AppConfig, environment variable loading
- `env.go`: `GetHomeDir()`, `GetWorkspaceDir()`, `GetProjectName()`
- `mcp.go`: MCP server configuration loading and placeholder replacement
- `type.go`: Common type aliases (e.g., `OpenAIMessage`)

## Chapter-Specific Notes

**Chapter 7 (RAG)**:
- Requires PostgreSQL with pgvector extension
- Uses GORM for database operations
- Implements semantic search via `SemanticSearchTool`
- Supports line-based and paragraph-based chunking

**Chapter 8 (Guardrails)**:
- Tool confirmation UI: Allow/Reject/Always Allow options with ↑↓ navigation
- ESC cancellation: Preserves conversation context, skips summary/memory
- `CommitTurn(skipPoliciesAndMemory bool)`: Controls whether to run policies/memory on commit
- `findTool(toolName)`: Unified tool lookup across native and MCP tools

**Chapter 9 (Skills System)**:
- Skills are descriptive prompts that guide models on how to use tools in specific scenarios
- Progressive loading: Only skill metadata (name + description) injected into system prompt at startup
- Full skill content loaded on-demand via `load_skill` tool
- Skill files in `.babyagent/skills/<skill-id>/SKILL.md` with YAML front matter
- Skill manager automatically discovers and loads skills from workspace

**System Prompt Placeholders** (ch08/prompt.go, ch09/prompt.go):
- `{runtime}`: OS (darwin/linux/windows)
- `{workspace_path}`: Current working directory
- `{memory}`: Injected global + workspace memory
- `{skills}`: List of available skills (name + description) - ch09 only

## Development Workflow

When modifying agent behavior:
1. Changes to `Agent.RunStreaming()` affect the core loop
2. Changes to policies affect how context is managed after each turn
3. Changes to system prompt affect how the agent behaves
4. TUI changes go in `ch08/tui/` or `ch09/tui/`, MessageVO types in `vo.go`
5. Skills are managed in `.babyagent/skills/` directory (ch09)

When adding new tools:
1. Implement `Tool` interface: `ToolName()`, `Info()`, `Execute()`
2. Register in `ch08/main/main.go` or `ch09/main/main.go` tool list
3. Tool will automatically be included in LLM requests

When adding skills (ch09):
1. Create skill directory: `.babyagent/skills/<skill-id>/`
2. Create `SKILL.md` with YAML front matter (name, description)
3. Skill will be auto-discovered and metadata injected into system prompt
4. LLM can call `load_skill(name="<skill-id>")` to load full content

When configuring tool confirmation:
1. Create `ToolConfirmConfig` with `RequireConfirmTools` map in `ch08/main/main.go` or `ch09/main/main.go`
2. Map tool names (e.g., `tool.AgentToolBash`) to `true` for tools requiring confirmation
3. For MCP tools, use full tool name (e.g., `babyagent_mcp__filesystem__write_file`)
4. `alwaysAllowTools` is managed automatically by Agent during session

When debugging issues:
- Set `log.SetOutput(os.Stdout)` in main/main.go to see logs
- Check policy execution via TUI events
- Verify token counts with `CountTokens()` from context/share.go
- For confirmation flow: check `MessageTypeToolConfirm` in TUI event handling
- For skills (ch09): Check `.babyagent/skills/` directory and verify skill metadata is loaded