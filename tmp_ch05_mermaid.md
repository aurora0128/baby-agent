```mermaid
flowchart TD
    classDef entry fill:#e8f1ff,stroke:#2563eb,stroke-width:2px,color:#0f172a;
    classDef orchestration fill:#f3e8ff,stroke:#9333ea,stroke-width:2px,color:#111827;
    classDef domain fill:#ecfccb,stroke:#65a30d,stroke-width:2px,color:#1f2937;
    classDef data fill:#fef3c7,stroke:#d97706,stroke-width:2px,color:#1f2937;
    classDef external fill:#ffe4e6,stroke:#e11d48,stroke-width:2px,color:#1f2937;
    classDef async fill:#e0f2fe,stroke:#0284c7,stroke-width:2px,color:#1f2937;

    User([User query]) --> TUI[TUI model]
    TUI --> Agent[RunStreaming turn]
    Agent --> ContextReq[Build request context]
    ContextReq --> FrontLLM[Front model]
    FrontLLM --> Assist[Assistant delta stream]
    Assist --> TUI
    FrontLLM --> ToolCall[Tool call request]
    ToolCall --> ToolExec[Execute native or MCP tool]
    ToolExec --> ToolMsg[Append tool message]
    ToolMsg --> FrontLLM
    Assist --> Commit[Commit turn]
    Commit --> PolicyPipe[Apply context policies]
    PolicyPipe --> Offload[Offload long tool output]
    PolicyPipe --> Summarize[Summarize old history]
    PolicyPipe --> Truncate[Trim old messages]
    Offload --> Store[(Memory storage)]
    Summarize --> BackLLM[Back model summarizer]
    PolicyPipe --> ContextState[(Managed message history)]
    ContextState --> ContextReq

    class User,TUI entry;
    class Agent,ContextReq,Commit,PolicyPipe orchestration;
    class Assist,Offload,Summarize,Truncate,ToolExec,ToolMsg domain;
    class Store,ContextState data;
    class FrontLLM,BackLLM external;
    class ToolCall async;
```

```mermaid
flowchart TD
    classDef entry fill:#e8f1ff,stroke:#2563eb,stroke-width:2px,color:#0f172a;
    classDef orchestration fill:#f3e8ff,stroke:#9333ea,stroke-width:2px,color:#111827;
    classDef domain fill:#ecfccb,stroke:#65a30d,stroke-width:2px,color:#1f2937;
    classDef data fill:#fef3c7,stroke:#d97706,stroke-width:2px,color:#1f2937;
    classDef external fill:#ffe4e6,stroke:#e11d48,stroke-width:2px,color:#1f2937;
    classDef async fill:#e0f2fe,stroke:#0284c7,stroke-width:2px,color:#1f2937;

    Submit([Press Enter]) --> Start[startNewTurn]
    Start --> Goroutine[Spawn streaming goroutine]
    Goroutine --> Run[Agent RunStreaming]
    Run --> Stream[Receive model chunks]
    Stream --> Reason[Reasoning events]
    Stream --> Answer[Content events]
    Stream --> Tool[Tool call events]
    Stream --> Policy[Policy status events]
    Reason --> Logs[Append log entries]
    Answer --> Logs
    Tool --> Logs
    Policy --> Logs
    Logs --> Viewport[Refresh viewport]
    Viewport --> Screen[Render Bubble Tea view]
    Abort([Esc cancel]) --> Cancel[Cancel context]
    Cancel --> Rollback[Rollback current turn logs]
    Clear([/clear]) --> Reset[Reset agent session]

    class Submit,Abort,Clear entry;
    class Start,Goroutine,Run,Viewport,Screen orchestration;
    class Reason,Answer,Tool,Policy,Rollback,Reset domain;
    class Logs data;
    class Cancel async;
```

```mermaid
flowchart TD
    classDef entry fill:#e8f1ff,stroke:#2563eb,stroke-width:2px,color:#0f172a;
    classDef orchestration fill:#f3e8ff,stroke:#9333ea,stroke-width:2px,color:#111827;
    classDef domain fill:#ecfccb,stroke:#65a30d,stroke-width:2px,color:#1f2937;
    classDef data fill:#fef3c7,stroke:#d97706,stroke-width:2px,color:#1f2937;
    classDef external fill:#ffe4e6,stroke:#e11d48,stroke-width:2px,color:#1f2937;
    classDef async fill:#e0f2fe,stroke:#0284c7,stroke-width:2px,color:#1f2937;

    Commit([CommitTurn]) --> Append[Append draft messages]
    Append --> Count[Recount tokens]
    Count --> CheckOffload{Usage over 0.4}
    CheckOffload -->|yes| Offload[Offload long tool messages]
    CheckOffload -->|no| CheckSummary{Usage over 0.6}
    Offload --> Storage[(Storage store)]
    Offload --> CheckSummary
    CheckSummary -->|yes| Summary[Batch summarize history]
    CheckSummary -->|no| CheckTruncate{Usage over 0.85}
    Summary --> BackLLM[Back model]
    Summary --> CheckTruncate
    CheckTruncate -->|yes| Truncate[Trim before latest user boundary]
    CheckTruncate -->|no| Finalize[Finalize context]
    Truncate --> Finalize
    Finalize --> History[(Message history)]

    class Commit entry;
    class Append,Count,CheckOffload,CheckSummary,CheckTruncate orchestration;
    class Offload,Summary,Truncate,Finalize domain;
    class Storage,History data;
    class BackLLM external;
```

```mermaid
flowchart TD
    classDef entry fill:#e8f1ff,stroke:#2563eb,stroke-width:2px,color:#0f172a;
    classDef orchestration fill:#f3e8ff,stroke:#9333ea,stroke-width:2px,color:#111827;
    classDef domain fill:#ecfccb,stroke:#65a30d,stroke-width:2px,color:#1f2937;
    classDef data fill:#fef3c7,stroke:#d97706,stroke-width:2px,color:#1f2937;
    classDef external fill:#ffe4e6,stroke:#e11d48,stroke-width:2px,color:#1f2937;
    classDef async fill:#e0f2fe,stroke:#0284c7,stroke-width:2px,color:#1f2937;

    Model([Assistant wants a tool]) --> Router[Tool router]
    Router --> Native[Native tools]
    Router --> MCP[MCP tool bridge]
    Native --> Bash[Bash tool]
    Native --> Load[Load storage tool]
    Bash --> Shell[Local shell]
    MCP --> McpClient[MCP client]
    McpClient --> McpServer[Configured MCP server]
    Shell --> Result[Tool result text]
    Load --> Result
    McpServer --> Result
    Result --> ToolMsg[Tool message in context]
    ToolMsg --> Offload[Offload preview if too long]
    Offload --> Store[(Storage key value)]
    Store --> Reload[load_storage by key]
    Reload --> FullText[Recovered full content]

    class Model entry;
    class Router,Native,MCP,McpClient,ToolMsg orchestration;
    class Bash,Load,Result,Offload,Reload,FullText domain;
    class Store data;
    class Shell,McpServer external;
```
