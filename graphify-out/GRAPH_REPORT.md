# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 106 nodes · 198 edges · 11 communities (10 shown, 1 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 5 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `51aa7ea8`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Model
- tui.go
- config.go
- go_pkg_fmt
- types.go
- header.go
- validator.go
- Verifier
- badge.go
- statusbar.go
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 16 edges
2. `Config` - 7 edges
3. `MessageItem` - 5 edges
4. `Message` - 5 edges
5. `Validator` - 5 edges
6. `Verifier` - 5 edges
7. `Load()` - 4 edges
8. `Patcher` - 4 edges
9. `RenderHeader()` - 4 edges
10. `NewModel()` - 4 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → cmd/root.go
- `initConfig()` --calls--> `Load()`  [EXTRACTED]
  cmd/root.go → internal/config/config.go
- `Model` --references--> `HeaderInfo`  [EXTRACTED]
  internal/ui/tui.go → internal/ui/header.go
- `TestRenderHeaderAndFooter()` --calls--> `NewModel()`  [INFERRED]
  internal/ui/header_test.go → internal/ui/tui.go
- `NewModel()` --references--> `HeaderInfo`  [EXTRACTED]
  internal/ui/tui.go → internal/ui/header.go

## Import Cycles
- None detected.

## Communities (11 total, 1 thin omitted)

### Community 0 - "Model"
Cohesion: 0.16
Nodes (3): AppState, MessageItem, Model

### Community 1 - "tui.go"
Cohesion: 0.16
Nodes (3): coderFinishedMsg, plannerFinishedMsg, reviewerFinishedMsg

### Community 2 - "config.go"
Cohesion: 0.19
Nodes (10): Execute(), initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig() (+2 more)

### Community 3 - "go_pkg_fmt"
Cohesion: 0.24
Nodes (4): Patcher, NewPatcher(), RenderDiffView(), DiffPatch

### Community 4 - "types.go"
Cohesion: 0.29
Nodes (7): Message, Patch, Plan, PlanStep, Role, ToolCall, ToolResult

### Community 5 - "header.go"
Cohesion: 0.38
Nodes (5): GetActiveWorkDir(), RenderHeader(), TestRenderHeaderAndFooter(), NewModel(), HeaderInfo

### Community 6 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

### Community 7 - "Verifier"
Cohesion: 0.47
Nodes (3): VerificationResult, Verifier, NewVerifier()

### Community 8 - "badge.go"
Cohesion: 0.83
Nodes (3): RenderToolBadge(), ActionStatus, ToolBadge

### Community 9 - "statusbar.go"
Cohesion: 0.83
Nodes (3): RenderFooter(), RenderStatusBar(), KeyHint

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges. (Counts symbols only; 24 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `Model` to `tui.go`, `header.go`?**
  _High betweenness centrality (0.189) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Why does `Verifier` connect `Verifier` to `go_pkg_fmt`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._