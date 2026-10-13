package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/xaxa/svyaz-cli/internal/llm"
)

type CoderAgent struct {
	client *llm.Client
}

func NewCoderAgent(client *llm.Client) *CoderAgent {
	return &CoderAgent{client: client}
}

const coderSystemPrompt = "Eres el Coder Agent de svyaz-cli (Связь).\n" +
	"Tu misión es implementar el código necesario cumpliendo estrictamente con la directiva 'ponytail':\n" +
	"1. YAGNI & CONCISIÓN EXTREMA: Prohibido sobre-estructurar, crear abstracciones innecesarias o paquetes redundantes.\n" +
	"2. FORMATO UNIFICADO: Emite ÚNICAMENTE un bloque de unified diff válido para git (```diff ... ```).\n" +
	"3. NO transcribas archivos completos. Solo genera el delta/diff con contexto mínimo suficiente para 'git apply'.\n" +
	"4. Formato de diff esperado:\n" +
	"```diff\n" +
	"--- a/ruta/del/archivo.go\n" +
	"+++ b/ruta/del/archivo.go\n" +
	"@@ -10,4 +10,6 @@\n" +
	" linea existente\n" +
	"-linea eliminada\n" +
	"+linea agregada\n" +
	" linea existente\n" +
	"```\n"

// GeneratePatch receives the task and file context, producing a clean unified diff Patch
func (c *CoderAgent) GeneratePatch(
	ctx context.Context,
	task string,
	plan *Plan,
	fileContexts map[string]string,
	correctionFeedback string,
	onStream func(token string),
) (*Patch, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== PLAN APROBADO ===\n%s\n\n", plan.Summary))

	if len(fileContexts) > 0 {
		sb.WriteString("=== CONTENIDO DE ARCHIVOS AFECTADOS ===\n")
		for path, content := range fileContexts {
			sb.WriteString(fmt.Sprintf("--- Archivo: %s ---\n%s\n\n", path, content))
		}
	}

	if correctionFeedback != "" {
		sb.WriteString(fmt.Sprintf("=== CORRECCIÓN DE ERROR PREVIO (Dry-Run falló) ===\n%s\nCorrige el diff asegurando que las líneas de contexto coincidan exactamente.\n\n", correctionFeedback))
	}

	sb.WriteString(fmt.Sprintf("=== TAREA ESPECÍFICA ===\n%s\nGenera el unified diff ahora:\n", task))

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: llm.RoleSystem, Content: coderSystemPrompt},
			{Role: llm.RoleUser, Content: sb.String()},
		},
	}

	resp, err := c.client.ExecuteWithFallback(ctx, req, onStream)
	if err != nil {
		return nil, fmt.Errorf("coder execution failed: %w", err)
	}

	content := resp.Choices[0].Message.Content
	filePath, diffText, addCount, delCount := ExtractDiff(content)

	if diffText == "" {
		return nil, fmt.Errorf("el modelo no produjo un bloque de diff unificado válido")
	}

	return &Patch{
		FilePath: filePath,
		Diff:     diffText,
		AddCount: addCount,
		DelCount: delCount,
	}, nil
}

// ExtractDiff extracts unified diff block from markdown or raw response
func ExtractDiff(raw string) (filePath, diffText string, addCount, delCount int) {
	// Look inside ```diff or ```patch block first
	lines := strings.Split(raw, "\n")
	var diffLines []string
	inDiffBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```diff") || strings.HasPrefix(trimmed, "```patch") {
			inDiffBlock = true
			continue
		}
		if inDiffBlock {
			if strings.HasPrefix(trimmed, "```") {
				break
			}
			diffLines = append(diffLines, line)
		}
	}

	if len(diffLines) == 0 {
		// Fallback: look for lines starting with '--- a/' or 'diff --git'
		recording := false
		for _, line := range lines {
			if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "--- a/") || strings.HasPrefix(line, "--- /dev/null") {
				recording = true
			}
			if recording {
				if strings.HasPrefix(line, "```") {
					break
				}
				diffLines = append(diffLines, line)
			}
		}
	}

	if len(diffLines) == 0 {
		return "", "", 0, 0
	}

	// Calculate stats and detect file path
	for _, l := range diffLines {
		if strings.HasPrefix(l, "+++ b/") {
			filePath = strings.TrimPrefix(l, "+++ b/")
		} else if strings.HasPrefix(l, "--- a/") && filePath == "" {
			filePath = strings.TrimPrefix(l, "--- a/")
		} else if strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++") {
			addCount++
		} else if strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "---") {
			delCount++
		}
	}

	diffText = strings.Join(diffLines, "\n") + "\n"
	return filePath, diffText, addCount, delCount
}
