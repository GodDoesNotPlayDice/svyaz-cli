# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 105 nodes · 195 edges · 9 communities (8 shown, 1 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 6 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6268e898`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- tui.go
- config.go
- go_pkg_fmt
- header.go
- chat.go
- types.go
- validator.go
- Verifier
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 15 edges
2. `Config` - 7 edges
3. `MessageItem` - 5 edges
4. `RenderHeader()` - 5 edges
5. `Message` - 5 edges
6. `Validator` - 5 edges
7. `Verifier` - 5 edges
8. `Load()` - 4 edges
9. `Patcher` - 4 edges
10. `TestRenderHeaderAndFooter()` - 4 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → cmd/root.go
- `initConfig()` --calls--> `Load()`  [EXTRACTED]
  cmd/root.go → internal/config/config.go
- `NewModel()` --references--> `HeaderInfo`  [EXTRACTED]
  internal/ui/tui.go → internal/ui/header.go
- `Model` --references--> `HeaderInfo`  [EXTRACTED]
  internal/ui/tui.go → internal/ui/header.go
- `TestRenderHeaderAndFooter()` --calls--> `RenderHeader()`  [INFERRED]
  internal/ui/header_test.go → internal/ui/header.go

## Import Cycles
- None detected.

## Communities (9 total, 1 thin omitted)

### Community 0 - "tui.go"
Cohesion: 0.12
Nodes (7): NewModel(), AppState, coderFinishedMsg, MessageItem, Model, plannerFinishedMsg, reviewerFinishedMsg

### Community 1 - "config.go"
Cohesion: 0.19
Nodes (10): Execute(), initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig() (+2 more)

### Community 2 - "go_pkg_fmt"
Cohesion: 0.24
Nodes (4): Patcher, NewPatcher(), RenderDiffView(), DiffPatch

### Community 3 - "header.go"
Cohesion: 0.23
Nodes (7): GetActiveWorkDir(), RenderHeader(), TestRenderHeaderAndFooter(), RenderFooter(), RenderStatusBar(), HeaderInfo, KeyHint

### Community 4 - "chat.go"
Cohesion: 0.23
Nodes (3): RenderToolBadge(), ActionStatus, ToolBadge

### Community 5 - "types.go"
Cohesion: 0.29
Nodes (7): Message, Patch, Plan, PlanStep, Role, ToolCall, ToolResult

### Community 6 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

### Community 7 - "Verifier"
Cohesion: 0.47
Nodes (3): VerificationResult, Verifier, NewVerifier()

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges. (Counts symbols only; 24 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `tui.go` to `header.go`?**
  _High betweenness centrality (0.176) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `tui.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11956521739130435 - nodes in this community are weakly interconnected._
- **Why does `Verifier` connect `Verifier` to `go_pkg_fmt`?**
  _High betweenness centrality (0.042) - this node is a cross-community bridge._