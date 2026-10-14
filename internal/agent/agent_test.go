package agent

import (
	"strings"
	"testing"
)

func TestExtractDiff_MarkdownBlock(t *testing.T) {
	raw := "Aquí está la solución:\n```diff\n--- a/main.go\n+++ b/main.go\n@@ -5,2 +5,3 @@\n func main() {\n+    println(\"hello\")\n }\n```\nEspero que sirva."

	filePath, diffText, addCount, delCount := ExtractDiff(raw)

	if filePath != "main.go" {
		t.Errorf("expected filePath 'main.go', got '%s'", filePath)
	}
	if !strings.Contains(diffText, "+    println(\"hello\")") {
		t.Errorf("expected diff content in diffText: %s", diffText)
	}
	if addCount != 1 {
		t.Errorf("expected addCount 1, got %d", addCount)
	}
	if delCount != 0 {
		t.Errorf("expected delCount 0, got %d", delCount)
	}
}

func TestParsePlanContent(t *testing.T) {
	raw := `### Resumen del Plan
Agregar endpoint /health para monitoreo básico.

### Archivos y Símbolos a Intervenir
- cmd/root.go: Agregar comando

### Pasos de Ejecución
1. Definir función HealthCheck
2. Registrar ruta en el router
`

	plan := parsePlanContent(raw)
	if !strings.Contains(plan.Summary, "Agregar endpoint /health") {
		t.Errorf("expected summary in plan, got '%s'", plan.Summary)
	}
	if len(plan.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(plan.Steps))
	}
}

func TestParsePlannerResult(t *testing.T) {
	chatRaw := "[CHAT] ¡Hola! ¿En qué puedo ayudarte hoy?"
	chatRes := parsePlannerResult(chatRaw)
	if chatRes.Intent != IntentChat {
		t.Errorf("expected IntentChat, got %s", chatRes.Intent)
	}
	if !strings.Contains(chatRes.Content, "¡Hola!") {
		t.Errorf("expected greeting in content, got '%s'", chatRes.Content)
	}
	if chatRes.Plan != nil {
		t.Errorf("expected nil Plan for IntentChat")
	}

	codeRaw := `[CODE_CHANGE]
### Resumen del Plan
Crear función Suma.

### Archivos y Símbolos a Intervenir
- calc.go: func Suma

### Pasos de Ejecución
1. Escribir función
`
	codeRes := parsePlannerResult(codeRaw)
	if codeRes.Intent != IntentCodeChange {
		t.Errorf("expected IntentCodeChange, got %s", codeRes.Intent)
	}
	if codeRes.Plan == nil || !strings.Contains(codeRes.Plan.Summary, "Crear función Suma") {
		t.Errorf("expected parsed Plan, got %+v", codeRes.Plan)
	}

	englishPlanRaw := `[CODE_CHANGE]
### Plan Summary
Add health check endpoint.

### Files and Symbols
- cmd/root.go: Add command
- internal/ui/tui.go: Update view

### Execution Steps
1. Define HealthCheck
2. Register route
`
	engRes := parsePlannerResult(englishPlanRaw)
	if engRes.Intent != IntentCodeChange {
		t.Errorf("expected IntentCodeChange, got %s", engRes.Intent)
	}
	if engRes.Plan == nil || len(engRes.Plan.Files) != 2 {
		t.Errorf("expected 2 files extracted in Plan, got %+v", engRes.Plan)
	}
	if engRes.Plan.Files[0] != "cmd/root.go" || engRes.Plan.Files[1] != "internal/ui/tui.go" {
		t.Errorf("expected [cmd/root.go internal/ui/tui.go], got %v", engRes.Plan.Files)
	}
}

func TestOrchestrator_GreetingIntent(t *testing.T) {
	orch := NewOrchestrator(nil, ".")
	res, err := orch.ProcessInput(nil, "Hola", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Intent != IntentChat {
		t.Errorf("expected IntentChat for 'Hola', got %s", res.Intent)
	}
	if res.Patch != nil {
		t.Errorf("expected nil Patch for greeting, got %v", res.Patch)
	}
	if !strings.Contains(res.Response, "¡Hola!") {
		t.Errorf("expected greeting response, got '%s'", res.Response)
	}

	// English greeting
	engRes, err := orch.ProcessInput(nil, "Hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(engRes.Response, "Hello!") {
		t.Errorf("expected English greeting, got '%s'", engRes.Response)
	}
}

