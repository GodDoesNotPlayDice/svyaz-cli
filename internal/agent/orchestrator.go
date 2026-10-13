package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	svcontext "github.com/xaxa/svyaz-cli/internal/context"
	"github.com/xaxa/svyaz-cli/internal/harness"
	"github.com/xaxa/svyaz-cli/internal/llm"
)

type OrchestratorCallbacks struct {
	OnStatusChange func(status string)
	OnPlanReady    func(plan *Plan)
	OnPatchReady   func(patch *Patch)
	OnFallback     func(ev llm.FallbackEvent)
}

type Orchestrator struct {
	client    *llm.Client
	planner   *PlannerAgent
	coder     *CoderAgent
	reviewer  *ReviewerAgent
	workDir   string
	repoMap   *svcontext.RepoMap
	Callbacks OrchestratorCallbacks
}

func NewOrchestrator(client *llm.Client, workDir string) *Orchestrator {
	if workDir == "" {
		workDir = "."
	}

	return &Orchestrator{
		client:   client,
		planner:  NewPlannerAgent(client),
		coder:    NewCoderAgent(client),
		reviewer: NewReviewerAgent(workDir),
		workDir:  workDir,
		repoMap:  svcontext.NewRepoMap(workDir),
	}
}

// ProcessInput processes any user message, detecting if it is conversational/question or a code modification task
func (o *Orchestrator) ProcessInput(
	ctx context.Context,
	userInput string,
	onChunk func(token string),
) (*ExecutionResult, error) {
	if o.client != nil && o.Callbacks.OnFallback != nil {
		o.client.OnFallback = o.Callbacks.OnFallback
	}

	trimmed := strings.ToLower(strings.TrimSpace(userInput))
	if trimmed == "hola" || trimmed == "hello" || trimmed == "hi" || trimmed == "buenas" || trimmed == "buenos dias" || trimmed == "buenas tardes" || trimmed == "hey" {
		return &ExecutionResult{
			Intent:   IntentChat,
			Response: "¡Hola! ¿En qué puedo ayudarte hoy en el proyecto? Puedes hacerme preguntas sobre la arquitectura o pedirme que agregue o modifique código.",
		}, nil
	}

	// 1. Capa 1 Arnés: Repo Map Grounding
	if o.Callbacks.OnStatusChange != nil {
		o.Callbacks.OnStatusChange("Analizando contexto del proyecto...")
	}

	mapText, err := o.repoMap.Generate()
	if err != nil {
		mapText = "Repo Map no disponible"
	}

	// 2. Ejecución de Planner con detección de intención
	if o.Callbacks.OnStatusChange != nil {
		o.Callbacks.OnStatusChange("Pensando...")
	}

	planRes, err := o.planner.PlanTask(ctx, userInput, mapText, onChunk)
	if err != nil {
		return nil, fmt.Errorf("fallo en análisis: %w", err)
	}

	// Si es conversación o consulta sin cambios en archivos
	if planRes.Intent == IntentChat {
		return &ExecutionResult{
			Intent:   IntentChat,
			Response: planRes.Content,
		}, nil
	}

	// Si es una tarea de modificación de código
	plan := planRes.Plan
	if plan == nil {
		return &ExecutionResult{
			Intent:   IntentChat,
			Response: planRes.Content,
		}, nil
	}

	if o.Callbacks.OnPlanReady != nil {
		o.Callbacks.OnPlanReady(plan)
	}

	// 3. Coder Agent: Preparar contexto de archivos
	if o.Callbacks.OnStatusChange != nil {
		o.Callbacks.OnStatusChange("Coder Agent: Generando unified diff con directiva 'ponytail'...")
	}

	fileContexts := make(map[string]string)
	for _, step := range plan.Steps {
		for _, f := range step.Files {
			p := filepath.Join(o.workDir, f)
			if data, err := os.ReadFile(p); err == nil {
				fileContexts[f] = string(data)
			}
		}
	}

	// Loop de Coder + DryRun (Capa 3 Arnés) con hasta 2 auto-correcciones
	var currentPatch *Patch
	var correctionFeedback string
	maxDryRunAttempts := 2

	for attempt := 1; attempt <= maxDryRunAttempts; attempt++ {
		patch, err := o.coder.GeneratePatch(ctx, userInput, plan, fileContexts, correctionFeedback, onChunk)
		if err != nil {
			return nil, fmt.Errorf("fallo generando diff: %w", err)
		}

		currentPatch = patch

		// Capa 3 Arnés: Dry-Run en memoria
		dryRunErr := o.reviewer.DryRun(patch)
		if dryRunErr == nil {
			break
		}

		if attempt < maxDryRunAttempts {
			correctionFeedback = fmt.Sprintf("git apply --check falló con: %v", dryRunErr)
			if o.Callbacks.OnStatusChange != nil {
				o.Callbacks.OnStatusChange("Coder Agent: Corrigiendo contexto de diff tras fallo de dry-run...")
			}
		} else {
			return nil, fmt.Errorf("parche no coincide con los archivos del repo tras %d intentos: %w", attempt, dryRunErr)
		}
	}

	if o.Callbacks.OnPatchReady != nil {
		o.Callbacks.OnPatchReady(currentPatch)
	}

	return &ExecutionResult{
		Intent: IntentCodeChange,
		Plan:   plan,
		Patch:  currentPatch,
	}, nil
}

// PlanAndGeneratePatch runs Planner -> Coder -> DryRun loop (compatibility wrapper)
func (o *Orchestrator) PlanAndGeneratePatch(
	ctx context.Context,
	userTask string,
	onChunk func(token string),
) (*Patch, *Plan, error) {
	res, err := o.ProcessInput(ctx, userTask, onChunk)
	if err != nil {
		return nil, nil, err
	}
	return res.Patch, res.Plan, nil
}

// ApplyAndVerify applies the approved patch to disk and executes compiler/linter verification
func (o *Orchestrator) ApplyAndVerify(patch *Patch) (*harness.VerificationResult, error) {
	if err := o.reviewer.Apply(patch); err != nil {
		return nil, fmt.Errorf("error aplicando parche en disco: %w", err)
	}

	res, err := o.reviewer.Verify()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando verificación determinista: %w", err)
	}

	return res, nil
}
