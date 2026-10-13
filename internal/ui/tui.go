package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	width        int
	height       int
	ready        bool
}

func NewModel(info HeaderInfo) Model {
	ti := textinput.New()
	ti.Placeholder = "Describe tu tarea o haz una pregunta (/help para comandos)..."
	ti.Prompt = PromptPromptStyle.Render("❯ ")
	ti.PromptStyle = PromptPromptStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTextPrimary))
	ti.PlaceholderStyle = PromptPlaceholderStyle
	ti.Focus()
	ti.CharLimit = 1000

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBlue))

	return Model{
		state:       StateIdle,
		headerInfo:  info,
		input:       ti,
		spinner:     sp,
		spinnerText: "Analizando proyecto con Graphify y AST...",
		messages: []MessageItem{
			{
				Role:    "agent",
				Content: "svyaz inicializado con arnés determinista de 4 capas y optimización FreeLLMAPI.\nEscribe tu requerimiento para comenzar.",
			},
		},
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

			if val == "/help" {
				m.addMessage(MessageItem{
					Role: "agent",
					Content: "Comandos disponibles:\n" +
						"  /help     - Muestra esta lista de comandos\n" +
						"  /clear    - Limpia el historial de la pantalla\n" +
						"  /status   - Muestra el estado del arnés y MCPs\n" +
						"  /quit     - Salir de svyaz-cli",
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
						"  Tokens usados: %d / %d\n"+
						"  Arnés: 4 Capas Activas (Grounding, MCP Gate, Dry-Run, Linter Loop)\n"+
						"  Skills: ponytail, ast-navigator",
						m.headerInfo.CurrentModel, m.headerInfo.TokensUsed, m.headerInfo.TokenLimit),
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

			// Start simulated Planner/Coder pipeline demo
			m.state = StatePlanning
			m.spinnerText = "Planner Agent: Consultando grafo de dependencias y AST..."
			m.updateViewportContent()

			return m, tea.Batch(
				m.spinner.Tick,
				// Simulate planning response
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
				m.spinnerText = "Reviewer Agent: Ejecutando verificación determinista local..."
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Parche aprobado. Aplicando en disco y validando...",
				})
				m.updateViewportContent()
				return m, func() tea.Msg {
					return reviewerFinishedMsg{}
				}
			case "d", "D", "n", "N":
				m.state = StateIdle
				m.currentPatch = nil
				m.addMessage(MessageItem{
					Role:    "agent",
					Content: "Parche descartado por el usuario.",
				})
				m.updateViewportContent()
				return m, nil
			}
		}

	case plannerFinishedMsg:
		m.state = StateCoding
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
		headerHeight := 2
		inputHeight := 3
		statusBarHeight := 2
		vpHeight := msg.Height - headerHeight - inputHeight - statusBarHeight
		if vpHeight < 5 {
			vpHeight = 5
		}

		if !m.ready {
			m.viewport = viewport.New(msg.Width, vpHeight)
			m.viewport.SetContent(m.renderContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = vpHeight
			m.updateViewportContent()
		}
		m.input.Width = msg.Width - 6
	}

	m.input, tiCmd = m.input.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

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

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			b.WriteString(UserPrefixStyle.Render("❯ ") + UserMessageStyle.Render(msg.Content) + "\n\n")
		case "agent":
			b.WriteString(AgentPrefixStyle.Render("svyaz: ") + AgentMessageStyle.Render(msg.Content) + "\n\n")
		case "tool":
			if msg.Badge != nil {
				b.WriteString(RenderToolBadge(*msg.Badge) + "\n\n")
			}
		case "diff":
			if msg.Patch != nil {
				b.WriteString(RenderDiffView(*msg.Patch, m.width) + "\n\n")
			}
		}
	}

	if m.state != StateIdle && m.state != StateConfirmPatch {
		sp := m.spinner.View() + " " + ActionTextStyle.Render(m.spinnerText)
		b.WriteString("  " + sp + "\n\n")
	}

	return b.String()
}

func (m Model) View() string {
	if !m.ready {
		return "Inicializando svyaz-cli..."
	}

	header := RenderHeader(m.headerInfo, m.width)
	vp := m.viewport.View()
	prompt := PromptContainerStyle.Width(m.width - 2).Render(m.input.View())

	var hints []KeyHint
	if m.state == StateConfirmPatch {
		hints = []KeyHint{
			{Key: "y", Desc: "aplicar"},
			{Key: "d", Desc: "descartar"},
			{Key: "Esc", Desc: "cancelar"},
		}
	} else {
		hints = []KeyHint{
			{Key: "Enter", Desc: "enviar"},
			{Key: "Esc", Desc: "limpiar"},
			{Key: "/help", Desc: "comandos"},
			{Key: "Ctrl+C", Desc: "salir"},
		}
	}
	statusBar := RenderStatusBar(hints, m.width)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		vp,
		prompt,
		statusBar,
	)
}
