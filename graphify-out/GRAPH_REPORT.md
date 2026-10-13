# Graph Report - svyaz-cli  (2026-10-12)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 99 nodes · 184 edges · 10 communities (7 shown, 3 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 4 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- tui.go
- chat.go
- run.go
- badge.go
- types.go
- config.go
- validator.go
- Verifier
- Patcher
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 15 edges
2. `Config` - 7 edges
3. `MessageItem` - 5 edges
4. `Message` - 5 edges
5. `Validator` - 5 edges
6. `Verifier` - 5 edges
7. `HeaderInfo` - 4 edges
8. `DiffPatch` - 4 edges
9. `ToolBadge` - 4 edges
10. `Load()` - 4 edges

## Surprising Connections (you probably didn't know these)
- `initConfig()` --calls--> `Load()`  [EXTRACTED]
  cmd/root.go → internal/config/config.go
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → cmd/root.go
- `MessageItem` --references--> `DiffPatch`  [EXTRACTED]
  internal/ui/tui.go → internal/ui/diffview.go
- `MessageItem` --references--> `ToolBadge`  [EXTRACTED]
  internal/ui/tui.go → internal/ui/badge.go
- `Model` --references--> `DiffPatch`  [EXTRACTED]
  internal/ui/tui.go → internal/ui/diffview.go

## Import Cycles
- None detected.

## Communities (10 total, 3 thin omitted)

### Community 0 - "tui.go"
Cohesion: 0.15
Nodes (9): RenderHeader(), NewModel(), AppState, coderFinishedMsg, HeaderInfo, MessageItem, Model, plannerFinishedMsg (+1 more)

### Community 1 - "chat.go"
Cohesion: 0.17
Nodes (5): Execute(), initConfig(), RenderStatusBar(), main(), KeyHint

### Community 3 - "badge.go"
Cohesion: 0.24
Nodes (3): RenderToolBadge(), ActionStatus, ToolBadge

### Community 4 - "types.go"
Cohesion: 0.29
Nodes (7): Message, Patch, Plan, PlanStep, Role, ToolCall, ToolResult

### Community 5 - "config.go"
Cohesion: 0.42
Nodes (7): AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig(), Load()

### Community 6 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

### Community 7 - "Verifier"
Cohesion: 0.47
Nodes (3): VerificationResult, Verifier, NewVerifier()

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges. (Counts symbols only; 23 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `tui.go` to `run.go`, `badge.go`?**
  _High betweenness centrality (0.179) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Why does `Verifier` connect `Verifier` to `run.go`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `Patcher` connect `Patcher` to `run.go`?**
  _High betweenness centrality (0.036) - this node is a cross-community bridge._