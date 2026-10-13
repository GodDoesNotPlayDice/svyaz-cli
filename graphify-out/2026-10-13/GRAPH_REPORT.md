# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 216 nodes · 528 edges · 12 communities (9 shown, 3 thin omitted)
- Extraction: 96% EXTRACTED · 4% INFERRED · 0% AMBIGUOUS · INFERRED: 21 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6f389717`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- chat.go
- tui.go
- llm/types.go
- agent/types.go
- Patch
- ModelPool
- planner.go
- config.go
- header.go
- validator.go
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 18 edges
2. `Client` - 16 edges
3. `ModelPool` - 15 edges
4. `Orchestrator` - 12 edges
5. `Patch` - 9 edges
6. `Plan` - 8 edges
7. `ReviewerAgent` - 8 edges
8. `NewModelPool()` - 8 edges
9. `NewOrchestrator()` - 8 edges
10. `Config` - 7 edges

## Surprising Connections (you probably didn't know these)
- `TestOrchestrator_GreetingIntent()` --calls--> `NewOrchestrator()`  [INFERRED]
  internal/agent/agent_test.go → internal/agent/orchestrator.go
- `NewOrchestrator()` --calls--> `NewReviewerAgent()`  [INFERRED]
  internal/agent/orchestrator.go → internal/agent/reviewer.go
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → cmd/root.go
- `initConfig()` --calls--> `Load()`  [EXTRACTED]
  cmd/root.go → internal/config/config.go
- `TestRenderHeaderAndFooter()` --calls--> `NewModel()`  [INFERRED]
  internal/ui/header_test.go → internal/ui/tui.go

## Import Cycles
- None detected.

## Communities (12 total, 3 thin omitted)

### Community 1 - "tui.go"
Cohesion: 0.09
Nodes (15): RenderToolBadge(), RenderDiffView(), TestRenderHeaderAndFooter(), NewModel(), renderPromptInput(), ActionStatus, AppState, coderFinishedMsg (+7 more)

### Community 2 - "llm/types.go"
Cohesion: 0.15
Nodes (15): ChatChoice, ChatChunkChoice, ChatChunkDelta, ChatCompletionChunk, ChatCompletionRequest, ChatCompletionResponse, ChatMessage, ChatUsage (+7 more)

### Community 3 - "agent/types.go"
Cohesion: 0.14
Nodes (17): IntentType, Message, Plan, PlannerResult, PlanStep, Role, ToolCall, ToolResult (+9 more)

### Community 4 - "Patch"
Cohesion: 0.14
Nodes (9): ReviewerAgent, Patcher, VerificationResult, Verifier, NewReviewerAgent(), Patch, NewPatcher(), NewVerifier() (+1 more)

### Community 5 - "ModelPool"
Cohesion: 0.16
Nodes (10): CompressOutput(), TestCompressOutput(), NewClient(), TestClient_StreamAndFallback(), TestLive_FreeLLMAPI_Connection(), NewModelPool(), TestModelPool_BasicOrder(), TestModelPool_CircuitBreakerOn429() (+2 more)

### Community 6 - "planner.go"
Cohesion: 0.18
Nodes (12): CoderAgent, OrchestratorCallbacks, PlannerAgent, RepoMap, NewCoderAgent(), Orchestrator, NewOrchestrator(), NewPlannerAgent() (+4 more)

### Community 7 - "config.go"
Cohesion: 0.23
Nodes (10): Execute(), initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig() (+2 more)

### Community 8 - "header.go"
Cohesion: 0.33
Nodes (6): GetActiveWorkDir(), RenderHeader(), RenderFooter(), RenderStatusBar(), HeaderInfo, KeyHint

### Community 9 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges. (Counts symbols only; 37 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `tui.go` to `Patch`, `planner.go`?**
  _High betweenness centrality (0.123) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `chat.go` be split into smaller, more focused modules?**
  _Cohesion score 0.13240418118466898 - nodes in this community are weakly interconnected._
- **Why does `ModelPool` connect `ModelPool` to `chat.go`, `llm/types.go`, `agent/types.go`, `planner.go`?**
  _High betweenness centrality (0.085) - this node is a cross-community bridge._
- **Should `tui.go` be split into smaller, more focused modules?**
  _Cohesion score 0.0855614973262032 - nodes in this community are weakly interconnected._
- **Why does `Orchestrator` connect `planner.go` to `chat.go`, `tui.go`, `Patch`?**
  _High betweenness centrality (0.078) - this node is a cross-community bridge._
- **Should `agent/types.go` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._