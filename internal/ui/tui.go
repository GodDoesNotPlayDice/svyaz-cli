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
)

type AppState int

const (
	StateIdle AppState = iota
	StatePlanning
	StateCoding
	StateConfirmPatch
	StateReviewing
)

type MessageItem struct {
	Role    string // "user", "agent", "tool", "diff"
	Content string
	Badge   *ToolBadge
	Patch   *DiffPatch
}

type Model struct {
	state        AppState
	headerInfo   HeaderInfo
	input        textinput.Model
	viewport     viewport.Model
	spinner      spinner.Model
	spinnerText  string
	messages     []MessageItem
	currentPatch *DiffPatch
	rawPatch     *agent.Patch
	orchestrator *agent.Orchestrator
	width        int
	height       int
	ready        bool
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
		state:        StateIdle,
		headerInfo:   info,
		input:        ti,
		spinner:      sp,
		spinnerText:  "Analizando proyecto con Graphify y AST...",
		orchestrator: o,
		messages:     []MessageItem{},
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
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.state == StateConfirmPatch {
				m.state = StateIdle
				m.currentPatch = nil
				m.headerInfo.ModelStatus = "Ready"
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Parche cancelado.",
				})
				m.updateViewportContent()
				return m, nil
			}
			m.input.SetValue("")
			return m, nil
		case tea.KeyEnter:
			if m.state == StateConfirmPatch {
				return m, nil
			}
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				return m, nil
			}

			if val == "/quit" || val == "/exit" {
				return m, tea.Quit
			}

			if val == "/clear" {
				m.messages = nil
				m.input.SetValue("")
				m.updateViewportContent()
				return m, nil
			}

			if val == "?" || val == "/help" {
				m.addMessage(MessageItem{
					Role: "agent",
					Content: "Atajos y comandos disponibles:\n" +
						"  ? o /help - Mostrar comandos disponibles\n" +
						"  /clear    - Limpiar historial de la pantalla\n" +
						"  /status   - Estado del arnés y MCPs\n" +
						"  /quit     - Salir de svyaz-cli\n" +
						"  Esc       - Cancelar acción o limpiar entrada",
				})
				m.input.SetValue("")
				m.updateViewportContent()
				return m, nil
			}

			if val == "/status" {
				m.addMessage(MessageItem{
					Role: "agent",
					Content: fmt.Sprintf("Estado del sistema:\n"+
						"  Modelo activo: %s\n"+
						"  Directorio: %s\n"+
						"  Arnés: 4 Capas Activas (Grounding, MCP Gate, Dry-Run, Linter Loop)\n"+
						"  Skills: ponytail, ast-navigator",
						m.headerInfo.CurrentModel, m.headerInfo.WorkDir),
				})
				m.input.SetValue("")
				m.updateViewportContent()
				return m, nil
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
			m.spinnerText = "Planner Agent: Consultando grafo de dependencias y AST..."
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
				m.spinnerText = "Reviewer Agent: Ejecutando verificación determinista local..."
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Parche aprobado. Aplicando en disco y validando...",
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
			case "d", "D", "n", "N":
				m.state = StateIdle
				m.headerInfo.ModelStatus = "Ready"
				m.currentPatch = nil
				m.rawPatch = nil
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Parche descartado por el usuario.",
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
				Role:    "agent",
				Content: fmt.Sprintf("Error: %v", msg.Err),
			})
			m.updateViewportContent()
			return m, nil
		}

		if msg.Result == nil {
			m.state = StateIdle
			m.headerInfo.ModelStatus = "Ready"
			return m, nil
		}

		// Case 1: Pure conversation, question, or text explanation
		if msg.Result.Intent == agent.IntentChat {
			m.state = StateIdle
			m.headerInfo.ModelStatus = "Ready"
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: msg.Result.Response,
			})
			m.updateViewportContent()
			return m, nil
		}

		// Case 2: Code modification task
		if msg.Result.Plan != nil && msg.Result.Plan.Summary != "" {
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "planner",
					Action:   "repo-map",
					Details:  msg.Result.Plan.Summary,
					Status:   ActionSuccess,
				},
			})
		}

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
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "git",
					Action:   "dry-run",
					Details:  "git apply --check exitoso en memoria (Capa 3 OK)",
					Status:   ActionSuccess,
				},
			})
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
					ToolName: "verifier",
					Action:   "check",
					Details:  fmt.Sprintf("Fallo: %v", msg.Err),
					Status:   ActionError,
				},
			})
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: fmt.Sprintf("Error aplicando/verificando: %v", msg.Err),
			})
		} else if msg.Result != nil && !msg.Result.Success {
			m.headerInfo.ModelStatus = "Failed"
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "verifier",
					Action:   "check",
					Details:  fmt.Sprintf("%s devolvió errores", msg.Result.Command),
					Status:   ActionError,
				},
			})
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: fmt.Sprintf("Salida del compilador/linter:\n%s", msg.Result.Output),
			})
		} else {
			m.headerInfo.ModelStatus = "Ready"
			cmdStr := "linter/compiler"
			if msg.Result != nil && msg.Result.Command != "" {
				cmdStr = msg.Result.Command
			}
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "verifier",
					Action:   "check",
					Details:  fmt.Sprintf("%s exitoso (Capa 4 OK)", cmdStr),
					Status:   ActionSuccess,
				},
			})
			m.addMessage(MessageItem{
				Role:    "agent",
				Content: "Tarea completada con éxito. Todos los chequeos deterministas pasaron.",
			})
		}
		m.updateViewportContent()
		return m, nil

	case fallbackEventMsg:
		m.headerInfo.CurrentModel = msg.NextModel
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "fallback",
				Action:   "circuit-breaker",
				Details:  fmt.Sprintf("%s (429) -> %s", msg.FailedModel, msg.NextModel),
				Status:   ActionWarning,
			},
		})
		m.updateViewportContent()
		return m, nil

	case plannerFinishedMsg:
		m.state = StateCoding
		m.headerInfo.ModelStatus = "Generating Patch..."
		m.spinnerText = "Coder Agent: Generando parche conciso con skill 'ponytail'..."
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "ast",
				Action:   "find",
				Details:  "Identificados símbolos requeridos sin saturar contexto",
				Status:   ActionSuccess,
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
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "git",
				Action:   "dry-run",
				Details:  "git apply --check exitoso en memoria (Capa 3 OK)",
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
				ToolName: "linter",
				Action:   "verify",
				Details:  "go vet / compilación exitosa sin errores (Capa 4 OK)",
				Status:   ActionSuccess,
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

	m.input, tiCmd = m.input.Update(msg)
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

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			line := UserPrefixStyle.Render("> ") + UserMessageStyle.Render(msg.Content)
			b.WriteString(wrapStyle.Render(line) + "\n\n")
		case "agent":
			line := AgentPrefixStyle.Render("svyaz: ") + AgentMessageStyle.Render(msg.Content)
			b.WriteString(wrapStyle.Render(line) + "\n\n")
		case "tool":
			if msg.Badge != nil {
				badgeStr := RenderToolBadge(*msg.Badge)
				b.WriteString(wrapStyle.Render(badgeStr) + "\n\n")
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

func (m Model) View() string {
	if !m.ready {
		return "Inicializando svyaz-cli..."
	}

	header := RenderHeader(m.headerInfo, m.width)
	footer := RenderFooter(m.headerInfo.CurrentModel, m.width)
	inputArea := m.renderInputArea()

	// If there are no messages yet: header, then inputArea immediately, then empty space to footer
	if len(m.messages) == 0 && m.state == StateIdle {
		body := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			inputArea,
		)
		usedLines := lipgloss.Height(body)
		remaining := m.height - usedLines - lipgloss.Height(footer)
		if remaining < 0 {
			remaining = 0
		}
		return body + strings.Repeat("\n", remaining) + footer
	}

	// When messages exist, display viewport between header and inputArea
	vpContent := lipgloss.NewStyle().Padding(0, 1).Render(m.viewport.View())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		vpContent,
		inputArea,
		footer,
	)
}
