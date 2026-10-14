package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xaxa/svyaz-cli/internal/agent"
	"github.com/xaxa/svyaz-cli/internal/harness"
	"github.com/xaxa/svyaz-cli/internal/llm"
	"github.com/xaxa/svyaz-cli/internal/mcp"
	"github.com/xaxa/svyaz-cli/internal/session"
)

type AppState int

const (
	StateIdle AppState = iota
	StatePlanning
	StateCoding
	StateConfirmPatch
	StateReviewing
	StateMCPMenu
	StateResumeMenu
)

type MessageItem struct {
	Role    string // "user", "agent", "tool", "diff", "thought"
	Content string
	Badge   *ToolBadge
	Patch   *DiffPatch
	Thought *ThoughtInfo
}

type SlashCommand struct {
	Name        string
	Description string
}

var defaultSlashCommands = []SlashCommand{
	{Name: "/help", Description: "Show available commands and shortcuts"},
	{Name: "/clear", Description: "Clear screen history and start fresh chat"},
	{Name: "/resume", Description: "Resume previous chat session"},
	{Name: "/mcp", Description: "Inspect MCP servers and available tools"},
	{Name: "/status", Description: "System, model, and harness status"},
	{Name: "/plan", Description: "Plan task with Graphify and AST"},
	{Name: "/verify", Description: "Run local deterministic verification"},
	{Name: "/quit", Description: "Exit svyaz-cli"},
}

type Model struct {
	state                 AppState
	headerInfo            HeaderInfo
	input                 textinput.Model
	viewport              viewport.Model
	spinner               spinner.Model
	spinnerText           string
	messages              []MessageItem
	currentPatch          *DiffPatch
	rawPatch              *agent.Patch
	orchestrator          *agent.Orchestrator
	totalTokens           int
	autocompleteIndex     int
	autocompleteDismissed bool
	width                 int
	height                int
	ready                 bool
	currentSessionID      string
	mcpServers            []mcp.MCPServerInfo
	mcpLocation           string
	mcpIndex              int
	mcpExpanded           string
	resumeSessions        []session.Session
	resumeIndex           int
}

func NewModel(info HeaderInfo, orch ...*agent.Orchestrator) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.PromptStyle = PromptPromptStyle
	ti.TextStyle = PromptInputTextStyle
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 2000

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBlue))

	if info.ModelStatus == "" {
		info.ModelStatus = "Ready"
	}

	var o *agent.Orchestrator
	if len(orch) > 0 {
		o = orch[0]
	}

	return Model{
		state:            StateIdle,
		headerInfo:       info,
		input:            ti,
		spinner:          sp,
		spinnerText:      "Analyzing project with Graphify and AST...",
		orchestrator:     o,
		messages:         []MessageItem{},
		currentSessionID: session.NewSessionID(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == StateMCPMenu {
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc, tea.KeyLeft:
				m.state = StateIdle
				m.headerInfo.ModelStatus = "Ready"
				return m, nil
			case tea.KeyUp:
				if len(m.mcpServers) > 0 {
					m.mcpIndex--
					if m.mcpIndex < 0 {
						m.mcpIndex = len(m.mcpServers) - 1
					}
				}
				return m, nil
			case tea.KeyDown:
				if len(m.mcpServers) > 0 {
					m.mcpIndex++
					if m.mcpIndex >= len(m.mcpServers) {
						m.mcpIndex = 0
					}
				}
				return m, nil
			case tea.KeyEnter:
				if len(m.mcpServers) > 0 && m.mcpIndex >= 0 && m.mcpIndex < len(m.mcpServers) {
					srv := m.mcpServers[m.mcpIndex]
					if m.mcpExpanded == srv.Name {
						m.mcpExpanded = ""
					} else {
						m.mcpExpanded = srv.Name
					}
				}
				return m, nil
			}
			return m, nil
		}

		if m.state == StateResumeMenu {
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc, tea.KeyLeft:
				m.state = StateIdle
				m.headerInfo.ModelStatus = "Ready"
				return m, nil
			case tea.KeyUp:
				if len(m.resumeSessions) > 0 {
					m.resumeIndex--
					if m.resumeIndex < 0 {
						m.resumeIndex = len(m.resumeSessions) - 1
					}
				}
				return m, nil
			case tea.KeyDown:
				if len(m.resumeSessions) > 0 {
					m.resumeIndex++
					if m.resumeIndex >= len(m.resumeSessions) {
						m.resumeIndex = 0
					}
				}
				return m, nil
			case tea.KeyEnter:
				if len(m.resumeSessions) > 0 && m.resumeIndex >= 0 && m.resumeIndex < len(m.resumeSessions) {
					m.saveCurrentSession()
					sel := m.resumeSessions[m.resumeIndex]
					loaded, err := session.LoadSession(sel.ID)
					if err == nil && loaded != nil {
						m.messages = loadSessionToMessages(loaded)
						m.currentSessionID = loaded.ID
						m.addMessage(MessageItem{
							Role:    "agent",
							Content: fmt.Sprintf("Resumed session \"%s\" (%s)", loaded.Title, loaded.UpdatedAt.Format("2006-01-02 15:04")),
						})
					}
					m.state = StateIdle
					m.headerInfo.ModelStatus = "Ready"
					m.updateViewportContent()
				}
				return m, nil
			}

			if msg.String() == "d" || msg.String() == "D" {
				if len(m.resumeSessions) > 0 && m.resumeIndex >= 0 && m.resumeIndex < len(m.resumeSessions) {
					sel := m.resumeSessions[m.resumeIndex]
					_ = session.DeleteSession(sel.ID)
					m.resumeSessions, _ = session.ListSessions()
					if m.resumeIndex >= len(m.resumeSessions) && m.resumeIndex > 0 {
						m.resumeIndex = len(m.resumeSessions) - 1
					}
				}
				return m, nil
			}
			return m, nil
		}

		matching := m.getMatchingCommands()
		isMenuOpen := len(matching) > 0 && !m.autocompleteDismissed

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			if isMenuOpen {
				m.autocompleteDismissed = true
				return m, nil
			}
			if m.state == StateConfirmPatch {
				m.state = StateIdle
				m.currentPatch = nil
				m.headerInfo.ModelStatus = "Ready"
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Patch cancelled.",
				})
				m.updateViewportContent()
				return m, nil
			}
			m.input.SetValue("")
			return m, nil

		case tea.KeyUp:
			if isMenuOpen {
				m.autocompleteIndex--
				if m.autocompleteIndex < 0 {
					m.autocompleteIndex = len(matching) - 1
				}
				return m, nil
			}

		case tea.KeyDown:
			if isMenuOpen {
				m.autocompleteIndex++
				if m.autocompleteIndex >= len(matching) {
					m.autocompleteIndex = 0
				}
				return m, nil
			}

		case tea.KeyTab:
			if isMenuOpen {
				idx := m.autocompleteIndex
				if idx >= 0 && idx < len(matching) {
					selected := matching[idx]
					m.input.SetValue(selected.Name + " ")
					m.input.SetCursor(len(m.input.Value()))
					m.autocompleteDismissed = true
				}
				return m, nil
			}

		case tea.KeyEnter:
			if m.state == StateConfirmPatch {
				return m, nil
			}

			// If autocomplete menu is active, select the highlighted command
			if isMenuOpen && len(matching) > 0 {
				idx := m.autocompleteIndex
				if idx >= 0 && idx < len(matching) {
					m.input.SetValue(matching[idx].Name)
				}
				m.autocompleteDismissed = true
			}

			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				return m, nil
			}
			m.autocompleteDismissed = false

			if val == "/quit" || val == "/exit" {
				return m, tea.Quit
			}

			if val == "/clear" {
				m.saveCurrentSession()
				m.messages = nil
				m.currentSessionID = session.NewSessionID()
				m.input.SetValue("")
				m.updateViewportContent()
				return m, nil
			}

			if val == "/mcp" {
				m.mcpServers, m.mcpLocation = mcp.DiscoverServers()
				m.mcpIndex = 0
				m.mcpExpanded = ""
				m.state = StateMCPMenu
				m.headerInfo.ModelStatus = "MCP View"
				m.input.SetValue("")
				return m, nil
			}

			if val == "/resume" {
				m.saveCurrentSession()
				sessions, _ := session.ListSessions()
				m.resumeSessions = sessions
				m.resumeIndex = 0
				m.state = StateResumeMenu
				m.headerInfo.ModelStatus = "Resume Session"
				m.input.SetValue("")
				return m, nil
			}

			if val == "?" || val == "/help" {
				m.addMessage(MessageItem{
					Role: "agent",
					Content: "Available commands and shortcuts:\n" +
						"  /help   - Show available commands and shortcuts\n" +
						"  /clear  - Clear screen history and start fresh chat\n" +
						"  /resume - View and restore previous chat sessions\n" +
						"  /mcp    - Inspect MCP servers and available tools\n" +
						"  /status - System, model, and harness status\n" +
						"  /plan   - Plan task with Graphify and AST\n" +
						"  /verify - Run local deterministic verification\n" +
						"  /quit   - Exit svyaz-cli\n" +
						"  Esc     - Cancel action or close menu",
				})
				m.input.SetValue("")
				m.updateViewportContent()
				return m, nil
			}

			if val == "/status" {
				m.addMessage(MessageItem{
					Role: "agent",
					Content: fmt.Sprintf("System status:\n"+
						"  Active model: %s\n"+
						"  Directory: %s\n"+
						"  Harness: 4 Active Layers (Grounding, MCP Gate, Dry-Run, Linter Loop)\n"+
						"  Skills: ponytail, ast-navigator",
						m.headerInfo.CurrentModel, m.headerInfo.WorkDir),
				})
				m.input.SetValue("")
				m.updateViewportContent()
				return m, nil
			}

			if val == "/plan" {
				m.addMessage(MessageItem{
					Role:    "user",
					Content: "/plan",
				})
				m.input.SetValue("")
				m.state = StatePlanning
				m.headerInfo.ModelStatus = "Planning..."
				m.spinnerText = "Planner Agent: Querying dependency graph and AST..."
				m.addMessage(MessageItem{
					Role: "tool",
					Badge: &ToolBadge{
						ToolName: "planner",
						Action:   "repo-map",
						Details:  "Analyzing repository map and key symbols...",
						Status:   ActionSuccess,
					},
				})
				m.updateViewportContent()
				if m.orchestrator != nil {
					orch := m.orchestrator
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							res, err := orch.ProcessInput(context.Background(), "Analyze project architecture and plan tasks", nil)
							return taskResultMsg{Result: res, Err: err}
						},
					)
				}
				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						return plannerFinishedMsg{}
					},
				)
			}

			if val == "/verify" {
				m.addMessage(MessageItem{
					Role:    "user",
					Content: "/verify",
				})
				m.input.SetValue("")
				m.state = StateReviewing
				m.headerInfo.ModelStatus = "Verifying..."
				m.spinnerText = "Reviewer Agent: Running local deterministic verification..."
				m.updateViewportContent()
				if m.orchestrator != nil {
					orch := m.orchestrator
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							res, err := orch.ApplyAndVerify(nil)
							return verificationDoneMsg{Result: res, Err: err}
						},
					)
				}
				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						return reviewerFinishedMsg{}
					},
				)
			}

			// Add user message
			m.addMessage(MessageItem{
				Role:    "user",
				Content: val,
			})
			m.input.SetValue("")

			// Start Planner/Coder pipeline
			m.state = StatePlanning
			m.headerInfo.ModelStatus = "Thinking..."
			m.spinnerText = "Planner Agent: Querying dependency graph and AST..."
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "Graphify/AST",
					Action:   "ast-graph",
					Details:  "ast-graph: querying dependencies and symbols",
					Status:   ActionSuccess,
				},
			})
			m.updateViewportContent()

			if m.orchestrator != nil {
				taskVal := val
				orch := m.orchestrator
				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						res, err := orch.ProcessInput(context.Background(), taskVal, nil)
						return taskResultMsg{
							Result: res,
							Err:    err,
						}
					},
				)
			}

			// Offline demo fallback if orchestrator is nil
			return m, tea.Batch(
				m.spinner.Tick,
				func() tea.Msg {
					return plannerFinishedMsg{}
				},
			)
		}

		// Handle patch confirmation keys
		if m.state == StateConfirmPatch {
			switch msg.String() {
			case "y", "Y":
				m.state = StateReviewing
				m.headerInfo.ModelStatus = "Verifying..."
				m.spinnerText = "Reviewer Agent: Running local deterministic verification..."
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Patch approved. Applying to disk and verifying...",
				})
				m.updateViewportContent()

				if m.orchestrator != nil && m.rawPatch != nil {
					rawP := m.rawPatch
					orch := m.orchestrator
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							res, err := orch.ApplyAndVerify(rawP)
							return verificationDoneMsg{Result: res, Err: err}
						},
					)
				}

				return m, func() tea.Msg {
					return reviewerFinishedMsg{}
				}
			case "e", "E":
				m.state = StateIdle
				m.headerInfo.ModelStatus = "Edit Mode"
				m.input.SetValue("")
				m.input.Placeholder = "Enter adjustment instructions for the patch..."
				m.input.Focus()
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Patch edit mode: type your refinement instructions below to adjust the patch, or press Esc to cancel.",
				})
				m.updateViewportContent()
				return m, textinput.Blink
			case "d", "D", "n", "N":
				m.state = StateIdle
				m.headerInfo.ModelStatus = "Ready"
				m.currentPatch = nil
				m.rawPatch = nil
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Patch discarded by user.",
				})
				m.updateViewportContent()
				return m, nil
			}
		}

	case taskResultMsg:
		if msg.Err != nil {
			m.state = StateIdle
			m.headerInfo.ModelStatus = "Error"
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "orchestrator",
					Action:   "error",
					Details:  fmt.Sprintf("Error: %v", msg.Err),
					Status:   ActionError,
				},
			})
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: fmt.Sprintf("Error processing request: %v", msg.Err),
			})
			m.updateViewportContent()
			return m, nil
		}

		if msg.Result == nil {
			m.state = StateIdle
			m.headerInfo.ModelStatus = "Ready"
			return m, nil
		}

		// Add Thought block if present
		if msg.Result.Thought != "" || msg.Result.TokensUsed > 0 {
			durStr := fmt.Sprintf("%ds", msg.Result.DurationSec)
			if msg.Result.DurationSec <= 0 {
				durStr = "2s"
			}
			thoughtText := msg.Result.Thought
			if thoughtText == "" {
				thoughtText = "Procesando razonamiento y analizando contexto..."
			}
			m.addMessage(MessageItem{
				Role: "thought",
				Thought: &ThoughtInfo{
					Duration: durStr,
					Tokens:   msg.Result.TokensUsed,
					Summary:  thoughtText,
				},
			})
		}

		// Preserve verbosity: record all execution steps taken
		for _, step := range msg.Result.Steps {
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName:   step.ToolName,
					Action:     step.Action,
					Details:    step.Details,
					Status:     toActionStatus(step.Status),
					Expandable: step.Expandable,
				},
			})
		}

		if msg.Result.TokensUsed > 0 {
			m.totalTokens += msg.Result.TokensUsed
		}

		// Case 1: Pure conversation, question, or text explanation
		if msg.Result.Intent == agent.IntentChat {
			m.state = StateIdle
			m.headerInfo.ModelStatus = "Ready"
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: msg.Result.Response,
			})
			m.saveCurrentSession()
			m.updateViewportContent()
			return m, nil
		}

		// Case 2: Code modification task
		if msg.Result.Patch != nil {
			m.rawPatch = msg.Result.Patch
			samplePatch := &DiffPatch{
				FilePath: msg.Result.Patch.FilePath,
				DiffText: msg.Result.Patch.Diff,
				AddCount: msg.Result.Patch.AddCount,
				DelCount: msg.Result.Patch.DelCount,
			}
			m.currentPatch = samplePatch
			m.state = StateConfirmPatch
			m.headerInfo.ModelStatus = "Awaiting Approval"
			m.addMessage(MessageItem{
				Role:  "diff",
				Patch: samplePatch,
			})
			m.updateViewportContent()
			return m, nil
		}

		m.state = StateIdle
		m.headerInfo.ModelStatus = "Ready"
		return m, nil

	case verificationDoneMsg:
		m.state = StateIdle
		m.currentPatch = nil
		m.rawPatch = nil
		if msg.Err != nil {
			m.headerInfo.ModelStatus = "Failed"
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "Bash",
					Details:  fmt.Sprintf("verification failed: %v", msg.Err),
					Status:   ActionError,
				},
			})
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: fmt.Sprintf("Error applying/verifying: %v", msg.Err),
			})
		} else if msg.Result != nil && !msg.Result.Success {
			m.headerInfo.ModelStatus = "Failed"
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "Bash",
					Details:  fmt.Sprintf("%s (exit status 1)", msg.Result.Command),
					Status:   ActionError,
				},
			})
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: fmt.Sprintf("Compiler/linter output:\n%s", msg.Result.Output),
			})
		} else {
			m.headerInfo.ModelStatus = "Ready"
			cmdStr := "go test -v ./..."
			if msg.Result != nil && msg.Result.Command != "" {
				cmdStr = msg.Result.Command
			}
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "Bash",
					Details:  cmdStr,
					Status:   ActionSuccess,
				},
			})
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName:   "engram/mem_session_summary",
					Details:    "Save Engram session summary",
					Status:     ActionSuccess,
					Expandable: true,
				},
			})
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: "Task completed successfully. All deterministic checks passed.",
			})
		}
		m.saveCurrentSession()
		m.updateViewportContent()
		return m, nil

	case fallbackEventMsg:
		m.headerInfo.CurrentModel = msg.NextModel
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "LLM/Fallback",
				Details:  fmt.Sprintf("%s (429) -> %s", msg.FailedModel, msg.NextModel),
				Status:   ActionWarning,
			},
		})
		m.updateViewportContent()
		return m, nil

	case plannerFinishedMsg:
		m.state = StateCoding
		m.headerInfo.ModelStatus = "Generating Patch..."
		m.spinnerText = "Coder Agent: Generating concise patch with 'ponytail' skill..."
		m.addMessage(MessageItem{
			Role: "thought",
			Thought: &ThoughtInfo{
				Duration: "9s",
				Tokens:   195,
				Summary:  "Examining the project structure and AST graph to plan concise modifications...",
			},
		})
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "Graphify/AST",
				Details:  "graphify extract . --code-only",
				Status:   ActionSuccess,
			},
		})
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName:   "Bash",
				Details:    "go test -v ./internal/ui/...",
				Status:     ActionSuccess,
				Expandable: true,
			},
		})
		m.updateViewportContent()
		return m, func() tea.Msg {
			return coderFinishedMsg{}
		}

	case coderFinishedMsg:
		samplePatch := &DiffPatch{
			FilePath: "cmd/root.go",
			DiffText: "--- a/cmd/root.go\n+++ b/cmd/root.go\n@@ -15,3 +15,5 @@\n func Execute() error {\n+    // Inicializar arnés determinista de 4 capas\n+    harness.Init()\n     return rootCmd.Execute()\n }",
			AddCount: 2,
			DelCount: 0,
		}
		m.currentPatch = samplePatch
		m.state = StateConfirmPatch
		m.headerInfo.ModelStatus = "Awaiting Approval"
		tokens := 188
		m.totalTokens += tokens
		m.addMessage(MessageItem{
			Role: "thought",
			Thought: &ThoughtInfo{
				Duration: "6s",
				Tokens:   tokens,
				Summary:  "I'm thinking through how to approach this. Applying minimal diff with ponytail skill...",
			},
		})
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "Edit",
				Details:  "cmd/root.go",
				Status:   ActionSuccess,
			},
		})
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName:   "Bash",
				Details:    "git apply --check cmd/root.go",
				Status:     ActionSuccess,
				Expandable: true,
			},
		})
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "engram/mem_save",
				Details:  "Save architecture decision",
				Status:   ActionSuccess,
			},
		})
		m.addMessage(MessageItem{
			Role:  "diff",
			Patch: samplePatch,
		})
		m.updateViewportContent()
		return m, nil

	case reviewerFinishedMsg:
		m.state = StateIdle
		m.headerInfo.ModelStatus = "Ready"
		m.currentPatch = nil
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "Bash",
				Details:  "go test -v ./...",
				Status:   ActionSuccess,
			},
		})
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName:   "engram/mem_session_summary",
				Details:    "Save Engram session summary",
				Status:     ActionSuccess,
				Expandable: true,
			},
		})
		m.addMessage(MessageItem{
			Role:    "agent",
			Content: "Tarea completada con éxito. Todos los chequeos pasaron.",
		})
		m.updateViewportContent()
		return m, nil

	case spinner.TickMsg:
		m.spinner, spCmd = m.spinner.Update(msg)
		return m, spCmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 7 // logo (5) + padding (2)
		inputHeight := 3  // top bar (1) + prompt line (1) + bottom bar (1)
		footerHeight := 1 // status bar
		vpHeight := msg.Height - headerHeight - inputHeight - footerHeight
		if vpHeight < 4 {
			vpHeight = 4
		}

		if !m.ready {
			m.viewport = viewport.New(msg.Width-2, vpHeight)
			m.viewport.SetContent(m.renderContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 2
			m.viewport.Height = vpHeight
			m.updateViewportContent()
		}
		m.input.Width = msg.Width - 4
	}

	prevVal := m.input.Value()
	m.input, tiCmd = m.input.Update(msg)
	newVal := m.input.Value()
	if prevVal != newVal {
		m.autocompleteDismissed = false
		m.autocompleteIndex = 0
	}
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

type taskResultMsg struct {
	Result *agent.ExecutionResult
	Err    error
}

type verificationDoneMsg struct {
	Result *harness.VerificationResult
	Err    error
}

type fallbackEventMsg llm.FallbackEvent

type plannerFinishedMsg struct{}
type coderFinishedMsg struct{}
type reviewerFinishedMsg struct{}

func toActionStatus(s string) ActionStatus {
	switch s {
	case "warning":
		return ActionWarning
	case "error":
		return ActionError
	case "pending":
		return ActionPending
	default:
		return ActionSuccess
	}
}

func (m *Model) addMessage(item MessageItem) {
	m.messages = append(m.messages, item)
}

func (m *Model) updateViewportContent() {
	if m.ready {
		m.viewport.SetContent(m.renderContent())
		m.viewport.GotoBottom()
	}
}

func (m Model) renderContent() string {
	var b strings.Builder

	wrapWidth := m.width - 4
	if wrapWidth < 20 {
		wrapWidth = 76
	}
	wrapStyle := lipgloss.NewStyle().Width(wrapWidth)

	for i, msg := range m.messages {
		switch msg.Role {
		case "user":
			line := UserPrefixStyle.Render("> ") + UserMessageStyle.Render(msg.Content)
			b.WriteString(wrapStyle.Render(line) + "\n\n")
		case "thought":
			if msg.Thought != nil {
				thStr := RenderThought(*msg.Thought, wrapWidth)
				b.WriteString(thStr + "\n\n")
			}
		case "agent":
			agentBlock := lipgloss.NewStyle().
				PaddingLeft(2).
				Width(wrapWidth).
				Render(AgentMessageStyle.Render(msg.Content))
			b.WriteString(agentBlock + "\n\n")
		case "tool":
			if msg.Badge != nil {
				badgeStr := RenderToolBadge(*msg.Badge)
				// Compactly stack consecutive tool items with a single newline
				if i+1 < len(m.messages) && m.messages[i+1].Role == "tool" {
					b.WriteString(wrapStyle.Render(badgeStr) + "\n")
				} else {
					b.WriteString(wrapStyle.Render(badgeStr) + "\n\n")
				}
			}
		case "diff":
			if msg.Patch != nil {
				b.WriteString(RenderDiffView(*msg.Patch, wrapWidth) + "\n\n")
			}
		}
	}

	if m.state != StateIdle && m.state != StateConfirmPatch {
		sp := m.spinner.View() + " " + ActionTextStyle.Render(m.spinnerText)
		b.WriteString(wrapStyle.Render("  "+sp) + "\n\n")
	}

	return b.String()
}

// renderPromptInput renders the prompt with a blinking underscore cursor '_'
func renderPromptInput(ti textinput.Model) string {
	prompt := PromptPromptStyle.Render(ti.Prompt)
	val := []rune(ti.Value())
	pos := ti.Position()
	if pos < 0 {
		pos = 0
	}
	if pos > len(val) {
		pos = len(val)
	}

	if !ti.Focused() {
		return prompt + ti.TextStyle.Render(string(val))
	}

	cursorChar := " "
	if !ti.Cursor.Blink {
		cursorChar = "_"
	}
	cursorRendered := PromptCursorStyle.Render(cursorChar)

	if len(val) == 0 {
		return prompt + cursorRendered
	}

	if pos == len(val) {
		return prompt + ti.TextStyle.Render(string(val)) + cursorRendered
	}

	// Cursor is over an existing character
	before := ti.TextStyle.Render(string(val[:pos]))
	after := ti.TextStyle.Render(string(val[pos+1:]))

	midChar := string(val[pos])
	if !ti.Cursor.Blink {
		midChar = ti.TextStyle.Underline(true).Render(midChar)
	} else {
		midChar = ti.TextStyle.Render(midChar)
	}

	return prompt + before + midChar + after
}

// renderInputArea renders the prompt framed immediately above and below by two subtle gray bars
func (m Model) renderInputArea() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	bar := SeparatorLineStyle.Render(strings.Repeat("─", w))
	promptLine := lipgloss.NewStyle().PaddingLeft(1).Render(renderPromptInput(m.input))

	return bar + "\n" + promptLine + "\n" + bar
}

func (m Model) getMatchingCommands() []SlashCommand {
	val := m.input.Value()
	if !strings.HasPrefix(val, "/") {
		return nil
	}
	if strings.Contains(val, " ") {
		return nil
	}

	var matches []SlashCommand
	for _, cmd := range defaultSlashCommands {
		if strings.HasPrefix(cmd.Name, val) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

// renderAutocompleteMenu renders the floating command navigation below the input line
func (m Model) renderAutocompleteMenu() string {
	matching := m.getMatchingCommands()
	if len(matching) == 0 || m.autocompleteDismissed {
		return ""
	}

	idx := m.autocompleteIndex
	if idx < 0 {
		idx = 0
	}
	if idx >= len(matching) {
		idx = len(matching) - 1
	}

	var sb strings.Builder

	maxNameLen := 0
	for _, cmd := range matching {
		if len(cmd.Name) > maxNameLen {
			maxNameLen = len(cmd.Name)
		}
	}

	for i, cmd := range matching {
		isSelected := i == idx
		padding := strings.Repeat(" ", maxNameLen-len(cmd.Name)+3)

		if isSelected {
			cursor := MenuCursorStyle.Render("› ")
			cmdStr := MenuCmdSelectedStyle.Render(cmd.Name + padding)
			descStr := MenuDescSelectedStyle.Render(cmd.Description)
			sb.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, cmdStr, descStr))
		} else {
			cmdStr := MenuCmdNormalStyle.Render(cmd.Name + padding)
			descStr := MenuDescNormalStyle.Render(cmd.Description)
			sb.WriteString(fmt.Sprintf("    %s%s\n", cmdStr, descStr))
		}
	}

	sb.WriteString("\n")

	// Shortcut navigation hint line 1:
	// ↑ / ↓ Navigate  ·  enter Select  ·  tab complete
	keyNav := MenuHintKeyStyle.Render("↑ / ↓")
	textNav := MenuHintTextStyle.Render(" Navigate")
	dot := MenuHintSeparatorStyle.Render("  ·  ")
	keyEnter := MenuHintKeyStyle.Render("enter")
	textEnter := MenuHintTextStyle.Render(" Select")
	keyTab := MenuHintKeyStyle.Render("tab")
	textTab := MenuHintTextStyle.Render(" complete")

	sb.WriteString(fmt.Sprintf("  %s%s%s%s%s%s%s%s\n",
		keyNav, textNav, dot, keyEnter, textEnter, dot, keyTab, textTab))

	// Shortcut navigation hint line 2 (directly under navigate):
	// esc to exit menu
	keyEsc := MenuHintKeyStyle.Render("esc")
	textEsc := MenuHintTextStyle.Render(" to exit menu")
	sb.WriteString(fmt.Sprintf("  %s%s", keyEsc, textEsc))

	return sb.String()
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing svyaz-cli..."
	}

	header := RenderHeader(m.headerInfo, m.width)
	footer := RenderFooter(m.headerInfo.CurrentModel, m.width)
	footerH := lipgloss.Height(footer)

	if m.state == StateMCPMenu {
		mcpView := RenderMCPView(m.mcpServers, m.mcpLocation, m.mcpIndex, m.mcpExpanded, m.width)
		mcpBox := lipgloss.NewStyle().Padding(1, 2).Render(mcpView)
		body := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			mcpBox,
		)
		usedLines := lipgloss.Height(body)
		remaining := m.height - usedLines - footerH
		if remaining < 0 {
			remaining = 0
		}
		return body + strings.Repeat("\n", remaining) + footer
	}

	if m.state == StateResumeMenu {
		resumeView := RenderResumeView(m.resumeSessions, m.resumeIndex, m.width)
		resumeBox := lipgloss.NewStyle().Padding(1, 2).Render(resumeView)
		body := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			resumeBox,
		)
		usedLines := lipgloss.Height(body)
		remaining := m.height - usedLines - footerH
		if remaining < 0 {
			remaining = 0
		}
		return body + strings.Repeat("\n", remaining) + footer
	}

	inputArea := m.renderInputArea()
	menuStr := m.renderAutocompleteMenu()

	headerH := lipgloss.Height(header)
	inputH := lipgloss.Height(inputArea)
	menuH := 0
	if menuStr != "" {
		menuH = lipgloss.Height(menuStr)
	}

	maxChatH := m.height - headerH - inputH - menuH - footerH
	if maxChatH < 2 {
		maxChatH = 2
	}

	inputWithMenu := inputArea
	if menuStr != "" {
		inputWithMenu = inputArea + "\n" + menuStr
	}

	contentStr := m.renderContent()
	trimmedContent := strings.TrimRight(contentStr, "\n")

	// Si no hay mensajes ni proceso en curso: header arriba, inputArea (+ menú) inmediatamente después
	if trimmedContent == "" {
		body := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			inputWithMenu,
		)
		usedLines := lipgloss.Height(body)
		remaining := m.height - usedLines - footerH
		if remaining < 0 {
			remaining = 0
		}
		return body + strings.Repeat("\n", remaining) + footer
	}

	contentH := lipgloss.Height(trimmedContent)

	// Mientras el chat quepa en pantalla: el banner queda arriba y el chat se alarga hacia abajo
	if contentH <= maxChatH {
		contentView := lipgloss.NewStyle().Padding(0, 1).Render(trimmedContent)
		body := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			contentView,
			inputWithMenu,
		)
		usedLines := lipgloss.Height(body)
		remaining := m.height - usedLines - footerH
		if remaining < 0 {
			remaining = 0
		}
		return body + strings.Repeat("\n", remaining) + footer
	}

	// Cuando el chat supera el espacio disponible: el banner permanece arriba y el viewport scrollea
	m.viewport.Height = maxChatH
	vpContent := lipgloss.NewStyle().Padding(0, 1).Render(m.viewport.View())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		vpContent,
		inputWithMenu,
		footer,
	)
}

func (m *Model) saveCurrentSession() {
	if len(m.messages) == 0 {
		return
	}
	var sessMsgs []session.SessionMessage
	for _, item := range m.messages {
		sMsg := session.SessionMessage{
			Role:    item.Role,
			Content: item.Content,
		}
		if item.Badge != nil {
			sMsg.ToolName = item.Badge.ToolName
			sMsg.Action = item.Badge.Action
			sMsg.Details = item.Badge.Details
			sMsg.Status = actionStatusToString(item.Badge.Status)
		}
		if item.Patch != nil {
			sMsg.DiffFile = item.Patch.FilePath
			sMsg.DiffText = item.Patch.DiffText
			sMsg.AddCount = item.Patch.AddCount
			sMsg.DelCount = item.Patch.DelCount
		}
		if item.Thought != nil {
			sMsg.Duration = item.Thought.Duration
			sMsg.Tokens = item.Thought.Tokens
			sMsg.Summary = item.Thought.Summary
		}
		sessMsgs = append(sessMsgs, sMsg)
	}

	if m.currentSessionID == "" {
		m.currentSessionID = session.NewSessionID()
	}

	sess := &session.Session{
		ID:       m.currentSessionID,
		Messages: sessMsgs,
	}
	_ = session.SaveSession(sess)
}

func actionStatusToString(s ActionStatus) string {
	switch s {
	case ActionPending:
		return "pending"
	case ActionWarning:
		return "warning"
	case ActionError:
		return "error"
	default:
		return "success"
	}
}

func loadSessionToMessages(sess *session.Session) []MessageItem {
	var items []MessageItem
	for _, sm := range sess.Messages {
		item := MessageItem{
			Role:    sm.Role,
			Content: sm.Content,
		}
		if sm.ToolName != "" || sm.Details != "" {
			item.Badge = &ToolBadge{
				ToolName: sm.ToolName,
				Action:   sm.Action,
				Details:  sm.Details,
				Status:   toActionStatus(sm.Status),
			}
		}
		if sm.DiffFile != "" || sm.DiffText != "" {
			item.Patch = &DiffPatch{
				FilePath: sm.DiffFile,
				DiffText: sm.DiffText,
				AddCount: sm.AddCount,
				DelCount: sm.DelCount,
			}
		}
		if sm.Summary != "" || sm.Tokens > 0 {
			item.Thought = &ThoughtInfo{
				Duration: sm.Duration,
				Tokens:   sm.Tokens,
				Summary:  sm.Summary,
			}
		}
		items = append(items, item)
	}
	return items
}
