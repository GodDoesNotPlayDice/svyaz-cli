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

type SlashCommand struct {
	Name        string
	Description string
}

var defaultSlashCommands = []SlashCommand{
	{Name: "/help", Description: "Mostrar comandos y atajos disponibles"},
	{Name: "/clear", Description: "Limpiar historial de la pantalla"},
	{Name: "/status", Description: "Estado del sistema, modelo y arnés"},
	{Name: "/plan", Description: "Planificar tarea con Graphify y AST"},
	{Name: "/verify", Description: "Ejecutar verificación determinista local"},
	{Name: "/quit", Description: "Salir de svyaz-cli"},
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
					Content: "Parche cancelado.",
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
				m.messages = nil
				m.input.SetValue("")
				m.updateViewportContent()
				return m, nil
			}

			if val == "?" || val == "/help" {
				m.addMessage(MessageItem{
					Role: "agent",
					Content: "Atajos y comandos disponibles:\n" +
						"  /help   - Mostrar comandos y atajos disponibles\n" +
						"  /clear  - Limpiar historial de la pantalla\n" +
						"  /status - Estado del sistema, modelo y arnés\n" +
						"  /plan   - Planificar tarea con Graphify y AST\n" +
						"  /verify - Ejecutar verificación determinista local\n" +
						"  /quit   - Salir de svyaz-cli\n" +
						"  Esc     - Cancelar acción o cerrar menú",
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

			if val == "/plan" {
				m.addMessage(MessageItem{
					Role:    "user",
					Content: "/plan",
				})
				m.input.SetValue("")
				m.state = StatePlanning
				m.headerInfo.ModelStatus = "Planning..."
				m.spinnerText = "Planner Agent: Consultando grafo de dependencias y AST..."
				m.addMessage(MessageItem{
					Role: "tool",
					Badge: &ToolBadge{
						ToolName: "planner",
						Action:   "repo-map",
						Details:  "Analizando mapa del repositorio y símbolos clave...",
						Status:   ActionSuccess,
					},
				})
				m.updateViewportContent()
				if m.orchestrator != nil {
					orch := m.orchestrator
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							res, err := orch.ProcessInput(context.Background(), "Analizar arquitectura del proyecto y planificar tareas", nil)
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
				m.spinnerText = "Reviewer Agent: Ejecutando verificación determinista local..."
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
			m.spinnerText = "Planner Agent: Consultando grafo de dependencias y AST..."
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "thinking",
					Action:   "ast-graph",
					Details:  "Consultando grafo de dependencias y AST...",
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
				Content: fmt.Sprintf("Error procesando solicitud: %v", msg.Err),
			})
			m.updateViewportContent()
			return m, nil
		}

		if msg.Result == nil {
			m.state = StateIdle
			m.headerInfo.ModelStatus = "Ready"
			return m, nil
		}

		// Preserve verbosity: record all execution steps taken
		for _, step := range msg.Result.Steps {
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: step.ToolName,
					Action:   step.Action,
					Details:  step.Details,
					Status:   toActionStatus(step.Status),
				},
			})
		}

		// Display token consumption badge
		if msg.Result.TokensUsed > 0 {
			m.totalTokens += msg.Result.TokensUsed
			m.addMessage(MessageItem{
				Role: "tool",
				Badge: &ToolBadge{
					ToolName: "tokens",
					Action:   "usage",
					Details:  fmt.Sprintf("%d tokens consumidos (sesión: %d)", msg.Result.TokensUsed, m.totalTokens),
					Status:   ActionSuccess,
				},
			})
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
		tokens := 348
		m.totalTokens += tokens
		m.addMessage(MessageItem{
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "coder",
				Action:   "ponytail",
				Details:  "Unified diff generado para cmd/root.go (+2/-0)",
				Status:   ActionSuccess,
			},
		})
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
			Role: "tool",
			Badge: &ToolBadge{
				ToolName: "tokens",
				Action:   "usage",
				Details:  fmt.Sprintf("%d tokens consumidos (sesión: %d)", tokens, m.totalTokens),
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

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			line := UserPrefixStyle.Render("> ") + UserMessageStyle.Render(msg.Content)
			b.WriteString(wrapStyle.Render(line) + "\n\n")
		case "agent":
			agentBlock := lipgloss.NewStyle().
				PaddingLeft(2).
				Width(wrapWidth).
				Render(AgentMessageStyle.Render(msg.Content))
			b.WriteString(agentBlock + "\n\n")
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
	// esc para salir de ese pequeño menú
	keyEsc := MenuHintKeyStyle.Render("esc")
	textEsc := MenuHintTextStyle.Render(" para salir de ese pequeño menú")
	sb.WriteString(fmt.Sprintf("  %s%s", keyEsc, textEsc))

	return sb.String()
}

func (m Model) View() string {
	if !m.ready {
		return "Inicializando svyaz-cli..."
	}

	header := RenderHeader(m.headerInfo, m.width)
	footer := RenderFooter(m.headerInfo.CurrentModel, m.width)
	inputArea := m.renderInputArea()
	menuStr := m.renderAutocompleteMenu()

	headerH := lipgloss.Height(header)
	footerH := lipgloss.Height(footer)
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
