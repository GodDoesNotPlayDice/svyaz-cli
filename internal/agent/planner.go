package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/xaxa/svyaz-cli/internal/llm"
)

type PlannerAgent struct {
	client *llm.Client
}

func NewPlannerAgent(client *llm.Client) *PlannerAgent {
	return &PlannerAgent{client: client}
}

const plannerSystemPrompt = `Eres svyaz-cli (Связь), un asistente de programación agéntico soberbio y minimalista en Go.
Tu primera responsabilidad es determinar con inteligencia la INTENCIÓN del usuario:

1. [CHAT]:
- Aplica cuando el usuario saluda ("hola", "buenas", "hey"), agradece, hace preguntas conceptuales o teóricas, consulta sobre qué hace el proyecto, pide explicaciones de código o comandos, o simplemente conversa.
- NO requiere modificar, crear ni eliminar ningún archivo en disco.
- REGLA: Comienza tu respuesta EXACTAMENTE con "[CHAT]" y responde de manera directa, natural, concisa y útil al usuario. NUNCA inventes archivos ficticios (como hola.go) para un saludo o consulta.

2. [CODE_CHANGE]:
- Aplica ÚNICAMENTE cuando el usuario pide explícitamente crear una nueva funcionalidad, escribir código, arreglar un bug, refactorizar o modificar archivos reales en el repositorio.
- REGLAS DE CONDUCTA (Arnés Capa 1):
  1. CERO CÓDIGO DIRECTO: Solo planifica, no generes la implementación completa aquí.
  2. ANCLAJE DE REALIDAD: Prohibido inventar dependencias o librerías ajenas al Repo Map.
  3. MINIMALISMO (YAGNI): Identifica únicamente los archivos y símbolos exactos a intervenir.
- REGLA: Comienza tu respuesta EXACTAMENTE con "[CODE_CHANGE]" seguido del plan estructurado:
[CODE_CHANGE]
### Resumen del Plan
[1-2 líneas]

### Archivos y Símbolos a Intervenir
- ruta/archivo.go: [descripción]

### Pasos de Ejecución
1. [Paso 1]
2. [Paso 2]
`

// PlanTask analyzes user request against repo map and determines intent (chat vs code change plan)
func (p *PlannerAgent) PlanTask(
	ctx context.Context,
	task string,
	repoMap string,
	onStream func(token string),
) (*PlannerResult, error) {
	prompt := fmt.Sprintf("%s\n\n=== MENSAJE DEL USUARIO ===\n%s\n", repoMap, task)

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: llm.RoleSystem, Content: plannerSystemPrompt},
			{Role: llm.RoleUser, Content: prompt},
		},
	}

	resp, err := p.client.ExecuteWithFallback(ctx, req, onStream)
	if err != nil {
		return nil, fmt.Errorf("planner execution failed: %w", err)
	}

	content := resp.Choices[0].Message.Content
	return parsePlannerResult(content), nil
}

func parsePlannerResult(content string) *PlannerResult {
	trimmed := strings.TrimSpace(content)

	if strings.HasPrefix(trimmed, "[CHAT]") {
		chatContent := strings.TrimSpace(strings.TrimPrefix(trimmed, "[CHAT]"))
		return &PlannerResult{
			Intent:  IntentChat,
			Content: chatContent,
			Plan:    nil,
		}
	}

	if strings.HasPrefix(trimmed, "[CODE_CHANGE]") {
		planContent := strings.TrimSpace(strings.TrimPrefix(trimmed, "[CODE_CHANGE]"))
		return &PlannerResult{
			Intent:  IntentCodeChange,
			Content: planContent,
			Plan:    parsePlanContent(planContent),
		}
	}

	// Fallback heuristic: check if content contains plan sections
	if strings.Contains(trimmed, "### Archivos") || strings.Contains(trimmed, "### Pasos") {
		return &PlannerResult{
			Intent:  IntentCodeChange,
			Content: trimmed,
			Plan:    parsePlanContent(trimmed),
		}
	}

	// Default to chat if no code changes planned
	return &PlannerResult{
		Intent:  IntentChat,
		Content: trimmed,
		Plan:    nil,
	}
}

func parsePlanContent(content string) *Plan {
	lines := strings.Split(content, "\n")
	plan := &Plan{
		Summary: "",
		Steps:   []PlanStep{},
	}

	var currentSection string
	var summaryLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### Resumen") {
			currentSection = "summary"
			continue
		} else if strings.HasPrefix(trimmed, "### Archivos") {
			currentSection = "files"
			continue
		} else if strings.HasPrefix(trimmed, "### Pasos") {
			currentSection = "steps"
			continue
		}

		switch currentSection {
		case "summary":
			if trimmed != "" {
				summaryLines = append(summaryLines, trimmed)
			}
		case "steps":
			if strings.HasPrefix(trimmed, "1.") || strings.HasPrefix(trimmed, "2.") ||
				strings.HasPrefix(trimmed, "3.") || strings.HasPrefix(trimmed, "-") {
				plan.Steps = append(plan.Steps, PlanStep{
					Description: strings.TrimLeft(trimmed, "0123456789.- "),
				})
			}
		}
	}

	if len(summaryLines) > 0 {
		plan.Summary = strings.Join(summaryLines, " ")
	} else {
		plan.Summary = strings.TrimSpace(content)
	}

	return plan
}
