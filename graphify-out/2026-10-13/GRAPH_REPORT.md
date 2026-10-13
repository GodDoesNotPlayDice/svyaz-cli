# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 228 nodes · 556 edges · 11 communities (10 shown, 1 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 26 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `04bd6910`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- tui.go
- Model
- header_test.go
- llm/types.go
- verifier.go
- config.go
- Client
- agent/types.go
- header.go
- validator.go
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 20 edges
2. `Client` - 16 edges
3. `ModelPool` - 15 edges
4. `Orchestrator` - 12 edges
5. `NewModel()` - 9 edges
6. `Patch` - 9 edges
7. `NewModelPool()` - 8 edges
8. `NewOrchestrator()` - 8 edges
9. `Plan` - 8 edges
10. `ReviewerAgent` - 7 edges

## Surprising Connections (you probably didn't know these)
- `TestOrchestrator_GreetingIntent()` --calls--> `NewOrchestrator()`  [INFERRED]
  internal/agent/agent_test.go → internal/agent/orchestrator.go
- `TestParsePlannerResult()` --calls--> `parsePlannerResult()`  [INFERRED]
  internal/agent/agent_test.go → internal/agent/planner.go
- `NewOrchestrator()` --calls--> `NewReviewerAgent()`  [INFERRED]
  internal/agent/orchestrator.go → internal/agent/reviewer.go
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → cmd/root.go
- `initConfig()` --calls--> `Load()`  [EXTRACTED]
  cmd/root.go → internal/config/config.go

## Import Cycles
- None detected.

## Communities (11 total, 1 thin omitted)

### Community 0 - "tui.go"
Cohesion: 0.13
Nodes (7): CompressOutput(), TestCompressOutput(), APIError, coderFinishedMsg, fallbackEventMsg, plannerFinishedMsg, reviewerFinishedMsg

### Community 1 - "Model"
Cohesion: 0.10
Nodes (11): RenderToolBadge(), RenderDiffView(), renderPromptInput(), toActionStatus(), ActionStatus, AppState, DiffPatch, MessageItem (+3 more)

### Community 2 - "header_test.go"
Cohesion: 0.11
Nodes (19): TestExtractDiff_MarkdownBlock(), TestOrchestrator_GreetingIntent(), TestParsePlannerResult(), ExtractDiff(), NewClient(), TestClient_StreamAndFallback(), TestLive_FreeLLMAPI_Connection(), NewModelPool() (+11 more)

### Community 3 - "llm/types.go"
Cohesion: 0.15
Nodes (15): ChatChoice, ChatChunkChoice, ChatChunkDelta, ChatCompletionChunk, ChatCompletionRequest, ChatCompletionResponse, ChatMessage, ChatUsage (+7 more)

### Community 4 - "verifier.go"
Cohesion: 0.15
Nodes (9): ReviewerAgent, Patcher, VerificationResult, Verifier, NewReviewerAgent(), Patch, NewPatcher(), NewVerifier() (+1 more)

### Community 5 - "config.go"
Cohesion: 0.15
Nodes (10): Execute(), initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig() (+2 more)

### Community 6 - "Client"
Cohesion: 0.18
Nodes (12): CoderAgent, OrchestratorCallbacks, PlannerAgent, RepoMap, NewCoderAgent(), Orchestrator, NewOrchestrator(), NewPlannerAgent() (+4 more)

### Community 7 - "agent/types.go"
Cohesion: 0.20
Nodes (14): ExecutionStep, IntentType, Message, Plan, PlannerResult, PlanStep, Role, ToolCall (+6 more)

### Community 8 - "header.go"
Cohesion: 0.33
Nodes (6): GetActiveWorkDir(), RenderHeader(), RenderFooter(), RenderStatusBar(), HeaderInfo, KeyHint

### Community 9 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges. (Counts symbols only; 39 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `Model` to `tui.go`, `header_test.go`, `verifier.go`, `Client`?**
  _High betweenness centrality (0.124) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `tui.go` be split into smaller, more focused modules?**
  _Cohesion score 0.13399778516057587 - nodes in this community are weakly interconnected._
- **Why does `ModelPool` connect `header_test.go` to `tui.go`, `llm/types.go`, `Client`, `agent/types.go`?**
  _High betweenness centrality (0.085) - this node is a cross-community bridge._
- **Should `Model` be split into smaller, more focused modules?**
  _Cohesion score 0.0967741935483871 - nodes in this community are weakly interconnected._
- **Why does `Client` connect `Client` to `tui.go`, `header_test.go`, `llm/types.go`?**
  _High betweenness centrality (0.079) - this node is a cross-community bridge._
- **Should `header_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11182795698924732 - nodes in this community are weakly interconnected._