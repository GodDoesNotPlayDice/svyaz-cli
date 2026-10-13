# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 224 nodes · 548 edges · 13 communities (11 shown, 2 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 25 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `00cba5a7`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- tui.go
- Model
- Client
- testing.T
- verifier.go
- agent/types.go
- llm/types.go
- config.go
- repomap.go
- header.go
- validator.go
- APIError
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 18 edges
2. `Client` - 16 edges
3. `ModelPool` - 15 edges
4. `Orchestrator` - 12 edges
5. `Patch` - 9 edges
6. `NewOrchestrator()` - 8 edges
7. `NewModelPool()` - 8 edges
8. `NewModel()` - 8 edges
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

## Communities (13 total, 2 thin omitted)

### Community 0 - "tui.go"
Cohesion: 0.16
Nodes (4): coderFinishedMsg, fallbackEventMsg, plannerFinishedMsg, reviewerFinishedMsg

### Community 1 - "Model"
Cohesion: 0.10
Nodes (10): RenderToolBadge(), RenderDiffView(), renderPromptInput(), toActionStatus(), ActionStatus, AppState, DiffPatch, MessageItem (+2 more)

### Community 2 - "Client"
Cohesion: 0.15
Nodes (14): CoderAgent, OrchestratorCallbacks, PlannerAgent, TestExtractDiff_MarkdownBlock(), TestOrchestrator_GreetingIntent(), ExtractDiff(), NewCoderAgent(), Orchestrator (+6 more)

### Community 3 - "testing.T"
Cohesion: 0.12
Nodes (16): CompressOutput(), TestCompressOutput(), NewClient(), TestClient_StreamAndFallback(), TestLive_FreeLLMAPI_Connection(), NewModelPool(), TestModelPool_BasicOrder(), TestModelPool_CircuitBreakerOn429() (+8 more)

### Community 4 - "verifier.go"
Cohesion: 0.14
Nodes (9): ReviewerAgent, Patcher, VerificationResult, Verifier, NewReviewerAgent(), Patch, NewPatcher(), NewVerifier() (+1 more)

### Community 5 - "agent/types.go"
Cohesion: 0.20
Nodes (15): ExecutionStep, IntentType, Message, Plan, PlannerResult, PlanStep, Role, ToolCall (+7 more)

### Community 6 - "llm/types.go"
Cohesion: 0.19
Nodes (13): ChatChoice, ChatChunkChoice, ChatChunkDelta, ChatCompletionChunk, ChatMessage, ChatUsage, FunctionCall, ModelMetadata (+5 more)

### Community 7 - "config.go"
Cohesion: 0.22
Nodes (10): Execute(), initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig() (+2 more)

### Community 8 - "repomap.go"
Cohesion: 0.24
Nodes (3): RepoMap, NewRepoMap(), TestRepoMap_Generate()

### Community 9 - "header.go"
Cohesion: 0.33
Nodes (6): GetActiveWorkDir(), RenderHeader(), RenderFooter(), RenderStatusBar(), HeaderInfo, KeyHint

### Community 10 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges. (Counts symbols only; 39 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `Model` to `tui.go`, `Client`, `testing.T`, `verifier.go`?**
  _High betweenness centrality (0.109) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Model` be split into smaller, more focused modules?**
  _Cohesion score 0.10317460317460317 - nodes in this community are weakly interconnected._
- **Why does `ModelPool` connect `testing.T` to `tui.go`, `Client`, `llm/types.go`?**
  _High betweenness centrality (0.085) - this node is a cross-community bridge._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.11692307692307692 - nodes in this community are weakly interconnected._
- **Why does `Client` connect `Client` to `tui.go`, `testing.T`?**
  _High betweenness centrality (0.079) - this node is a cross-community bridge._
- **Should `verifier.go` be split into smaller, more focused modules?**
  _Cohesion score 0.1383399209486166 - nodes in this community are weakly interconnected._