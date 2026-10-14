# Graph Report - svyaz-cli  (2026-10-13)

## Corpus Check
- 40 files · ~18,379 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 2 file(s) not represented in the graph (top: .example 1, (none) 1)

## Summary
- 272 nodes · 730 edges · 13 communities (12 shown, 1 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 41 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `18a35398`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- tui.go
- Model
- verifier.go
- llm/types.go
- agent/types.go
- discovery.go
- config.go
- testing.T
- svyaz-cli (Связь)
- validator.go
- repomap.go
- github.com/xaxa/svyaz-cli
- statusbar.go

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
- `NewOrchestrator()` --calls--> `NewReviewerAgent()`  [INFERRED]
  internal/agent/orchestrator.go → internal/agent/reviewer.go
- `taskResultMsg` --references--> `ExecutionResult`  [EXTRACTED]
  internal/ui/tui.go → internal/agent/types.go
- `TestExtractDiff_MarkdownBlock()` --calls--> `ExtractDiff()`  [INFERRED]
  internal/agent/agent_test.go → internal/agent/coder.go

## Import Cycles
- None detected.

## Communities (13 total, 1 thin omitted)

### Community 0 - "tui.go"
Cohesion: 0.11
Nodes (9): Execute(), FallbackEvent, APIError, main(), coderFinishedMsg, fallbackEventMsg, plannerFinishedMsg, reviewerFinishedMsg (+1 more)

### Community 1 - "Model"
Cohesion: 0.08
Nodes (29): DeleteSession(), GetStorageDir(), Session, ListSessions(), LoadSession(), NewSessionID(), SaveSession(), TestSessionSaveListLoadDelete() (+21 more)

### Community 2 - "verifier.go"
Cohesion: 0.15
Nodes (8): ReviewerAgent, NewReviewerAgent(), Patcher, NewPatcher(), VerificationResult, Verifier, NewVerifier(), verificationDoneMsg

### Community 3 - "llm/types.go"
Cohesion: 0.15
Nodes (15): ChatChoice, ChatChunkChoice, ChatChunkDelta, ChatCompletionChunk, ChatCompletionRequest, ChatCompletionResponse, ChatMessage, ChatUsage (+7 more)

### Community 4 - "agent/types.go"
Cohesion: 0.12
Nodes (24): CoderAgent, ExecutionStep, IntentType, Message, OrchestratorCallbacks, Plan, PlannerAgent, PlannerResult (+16 more)

### Community 5 - "discovery.go"
Cohesion: 0.19
Nodes (12): defaultFallbackServers(), DiscoverServers(), MCPServerInfo, queryLiveMCPTools(), TestRenderMCPView(), formatToolsStyling(), formatToolsSummary(), RenderMCPView() (+4 more)

### Community 6 - "config.go"
Cohesion: 0.26
Nodes (12): initConfig(), AgentConfig, Config, LLMConfig, MCPServerConfig, SkillsConfig, DefaultConfig(), Load() (+4 more)

### Community 7 - "testing.T"
Cohesion: 0.11
Nodes (21): TestExtractDiff_MarkdownBlock(), ExtractDiff(), CompressOutput(), TestCompressOutput(), NewClient(), TestClient_StreamAndFallback(), TestLive_FreeLLMAPI_Connection(), NewModelPool() (+13 more)

### Community 8 - "svyaz-cli (Связь)"
Cohesion: 0.50
Nodes (3): ⚡ Características Principales, 🛠 Instalación y Compilación, svyaz-cli (Связь)

### Community 9 - "validator.go"
Cohesion: 0.48
Nodes (4): NewValidator(), ToolDefinition, ToolParameterSpec, Validator

### Community 10 - "repomap.go"
Cohesion: 0.24
Nodes (3): RepoMap, NewRepoMap(), TestRepoMap_Generate()

### Community 12 - "statusbar.go"
Cohesion: 0.83
Nodes (3): RenderFooter(), RenderStatusBar(), KeyHint

## Knowledge Gaps
- **1 isolated node(s):** `github.com/xaxa/svyaz-cli`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 41 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Model` connect `Model` to `tui.go`, `agent/types.go`, `discovery.go`, `testing.T`?**
  _High betweenness centrality (0.105) - this node is a cross-community bridge._
- **What connects `github.com/xaxa/svyaz-cli` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `tui.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11363636363636363 - nodes in this community are weakly interconnected._
- **Why does `ModelPool` connect `testing.T` to `tui.go`, `llm/types.go`, `agent/types.go`?**
  _High betweenness centrality (0.064) - this node is a cross-community bridge._
- **Should `Model` be split into smaller, more focused modules?**
  _Cohesion score 0.0797872340425532 - nodes in this community are weakly interconnected._
- **Why does `Orchestrator` connect `agent/types.go` to `tui.go`, `Model`, `verifier.go`, `repomap.go`?**
  _High betweenness centrality (0.056) - this node is a cross-community bridge._
- **Should `agent/types.go` be split into smaller, more focused modules?**
  _Cohesion score 0.12258064516129032 - nodes in this community are weakly interconnected._