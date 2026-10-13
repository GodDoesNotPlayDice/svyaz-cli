# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 233 nodes · 569 edges · 13 communities (10 shown, 3 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 28 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `da265189`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- tui.go
- Model
- header_test.go
- llm/types.go
- verifier.go
- Client
- config.go
- agent/types.go
- header.go
- validator.go
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 20 edges
2. `Client` - 16 edges
3. `ModelPool` - 15 edges
4. `Orchestrator` - 12 edges
5. `NewModel()` - 10 edges
6. `Patch` - 9 edges
7. `NewModelPool()` - 8 edges
8. `NewOrchestrator()` - 8 edges
9. `Plan` - 8 edges
10. `ReviewerAgent` - 7 edges

## Surprising Connections (you probably didn't know these)
- `NewOrchestrator()` --calls--> `NewReviewerAgent()`  [INFERRED]
  internal/agent/orchestrator.go → internal/agent/reviewer.go
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → cmd/root.go
- `initConfig()` --calls--> `Load()`  [EXTRACTED]
  cmd/root.go → internal/config/config.go
- `Model` --references--> `Orchestrator`  [EXTRACTED]
  internal/ui/tui.go → internal/agent/orchestrator.go
- `Model` --references--> `Patch`  [EXTRACTED]
  internal/ui/tui.go → internal/agent/types.go

## Import Cycles
- None detected.

## Communities (13 total, 3 thin omitted)

### Community 0 - "tui.go"
Cohesion: 0.14
Nodes (5): APIError, coderFinishedMsg, fallbackEventMsg, plannerFinishedMsg, reviewerFinishedMsg

### Community 1 - "Model"
Cohesion: 0.10
Nodes (13): RenderThought(), RenderToolBadge(), RenderDiffView(), renderPromptInput(), toActionStatus(), ActionStatus, AppState, DiffPatch (+5 more)

### Community 2 - "header_test.go"
Cohesion: 0.12
Nodes (18): CompressOutput(), TestCompressOutput(), NewClient(), TestClient_StreamAndFallback(), TestLive_FreeLLMAPI_Connection(), NewModelPool(), TestModelPool_BasicOrder(), TestModelPool_CircuitBreakerOn429() (+10 more)

### Community 3 - "llm/types.go"
Cohesion: 0.15
Nodes (15): ChatChoice, ChatChunkChoice, ChatChunkDelta, ChatCompletionChunk, ChatCompletionRequest, ChatCompletionResponse, ChatMessage, ChatUsage (+7 more)

### Community 4 - "verifier.go"
Cohesion: 0.15
Nodes (9): ReviewerAgent, Patcher, VerificationResult, Verifier, NewReviewerAgent(), Patch, NewPatcher(), NewVerifier() (+1 more)

### Community 5 - "Client"
Cohesion: 0.13
Nodes (13): CoderAgent, OrchestratorCallbacks, PlannerAgent, RepoMap, TestOrchestrator_GreetingIntent(), NewCoderAgent(), Orchestrator, NewOrchestrator() (+5 more)

### Community 6 - "config.go"
Cohesion: 0.15
Nodes (10): Execute(), initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig() (+2 more)

### Community 7 - "agent/types.go"
Cohesion: 0.16
Nodes (17): ExecutionStep, IntentType, Message, Plan, PlannerResult, PlanStep, Role, ToolCall (+9 more)

### Community 8 - "header.go"
Cohesion: 0.33
Nodes (6): GetActiveWorkDir(), RenderHeader(), RenderFooter(), RenderStatusBar(), HeaderInfo, KeyHint

### Community 9 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges. (Counts symbols only; 41 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `Model` to `tui.go`, `header_test.go`, `verifier.go`, `Client`?**
  _High betweenness centrality (0.119) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `tui.go` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._
- **Why does `ModelPool` connect `header_test.go` to `tui.go`, `llm/types.go`, `Client`?**
  _High betweenness centrality (0.081) - this node is a cross-community bridge._
- **Should `Model` be split into smaller, more focused modules?**
  _Cohesion score 0.10344827586206896 - nodes in this community are weakly interconnected._
- **Why does `Client` connect `Client` to `tui.go`, `header_test.go`, `llm/types.go`?**
  _High betweenness centrality (0.080) - this node is a cross-community bridge._
- **Should `header_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.1206896551724138 - nodes in this community are weakly interconnected._