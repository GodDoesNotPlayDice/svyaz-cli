# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- 40 files · ~17,840 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 2 file(s) not represented in the graph (top: .example 1, (none) 1)

## Summary
- 270 nodes · 720 edges · 12 communities (11 shown, 1 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 41 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `18a35398`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- tui.go
- Model
- testing.T
- llm/types.go
- agent/types.go
- mcp.go
- config.go
- ModelPool
- svyaz-cli (Связь)
- validator.go
- RepoMap
- github.com/xaxa/svyaz-cli

## God Nodes (most connected - your core abstractions)
1. `Model` - 23 edges
2. `Client` - 16 edges
3. `ModelPool` - 15 edges
4. `Orchestrator` - 12 edges
5. `NewModel()` - 12 edges
6. `Patch` - 9 edges
7. `Session` - 9 edges
8. `NewOrchestrator()` - 8 edges
9. `ReviewerAgent` - 8 edges
10. `Plan` - 8 edges

## Surprising Connections (you probably didn't know these)
- `initConfig()` --calls--> `Load()`  [EXTRACTED]
  cmd/root.go → internal/config/config.go
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → cmd/root.go
- `TestParsePlanContent()` --calls--> `parsePlanContent()`  [INFERRED]
  internal/agent/agent_test.go → internal/agent/planner.go
- `TestParsePlannerResult()` --calls--> `parsePlannerResult()`  [INFERRED]
  internal/agent/agent_test.go → internal/agent/planner.go
- `TestOrchestrator_GreetingIntent()` --calls--> `NewOrchestrator()`  [INFERRED]
  internal/agent/agent_test.go → internal/agent/orchestrator.go

## Import Cycles
- None detected.

## Communities (12 total, 1 thin omitted)

### Community 0 - "tui.go"
Cohesion: 0.12
Nodes (5): APIError, coderFinishedMsg, fallbackEventMsg, plannerFinishedMsg, reviewerFinishedMsg

### Community 1 - "Model"
Cohesion: 0.09
Nodes (25): DeleteSession(), GetStorageDir(), Session, ListSessions(), LoadSession(), NewSessionID(), SaveSession(), TestSessionSaveListLoadDelete() (+17 more)

### Community 2 - "testing.T"
Cohesion: 0.16
Nodes (18): TestExtractDiff_MarkdownBlock(), TestOrchestrator_GreetingIntent(), TestParsePlanContent(), TestParsePlannerResult(), ExtractDiff(), CompressOutput(), TestCompressOutput(), TestAntigravityVerbosityAndThoughts() (+10 more)

### Community 3 - "llm/types.go"
Cohesion: 0.15
Nodes (15): ChatChoice, ChatChunkChoice, ChatChunkDelta, ChatCompletionChunk, ChatCompletionRequest, ChatCompletionResponse, ChatMessage, ChatUsage (+7 more)

### Community 4 - "agent/types.go"
Cohesion: 0.08
Nodes (31): CoderAgent, ExecutionStep, IntentType, Message, OrchestratorCallbacks, Plan, PlannerAgent, PlannerResult (+23 more)

### Community 5 - "mcp.go"
Cohesion: 0.13
Nodes (15): defaultFallbackServers(), DiscoverServers(), MCPServerInfo, GetActiveWorkDir(), RenderHeader(), formatToolsStyling(), formatToolsSummary(), RenderMCPView() (+7 more)

### Community 6 - "config.go"
Cohesion: 0.18
Nodes (14): Execute(), initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig() (+6 more)

### Community 7 - "ModelPool"
Cohesion: 0.15
Nodes (8): NewClient(), TestClient_StreamAndFallback(), TestLive_FreeLLMAPI_Connection(), NewModelPool(), TestModelPool_BasicOrder(), TestModelPool_CircuitBreakerOn429(), TestModelPool_FilterRequiresTools(), ModelPool

### Community 8 - "svyaz-cli (Связь)"
Cohesion: 0.50
Nodes (3): ⚡ Características Principales, 🛠 Instalación y Compilación, svyaz-cli (Связь)

### Community 9 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

### Community 10 - "RepoMap"
Cohesion: 0.47
Nodes (3): RepoMap, NewRepoMap(), TestRepoMap_Generate()

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 43 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `Model` to `tui.go`, `testing.T`, `agent/types.go`, `mcp.go`?**
  _High betweenness centrality (0.106) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `tui.go` be split into smaller, more focused modules?**
  _Cohesion score 0.12326530612244897 - nodes in this community are weakly interconnected._
- **Why does `ModelPool` connect `ModelPool` to `tui.go`, `llm/types.go`, `agent/types.go`?**
  _High betweenness centrality (0.064) - this node is a cross-community bridge._
- **Should `Model` be split into smaller, more focused modules?**
  _Cohesion score 0.09090909090909091 - nodes in this community are weakly interconnected._
- **Why does `Orchestrator` connect `agent/types.go` to `tui.go`, `Model`, `RepoMap`?**
  _High betweenness centrality (0.057) - this node is a cross-community bridge._
- **Should `agent/types.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08181818181818182 - nodes in this community are weakly interconnected._