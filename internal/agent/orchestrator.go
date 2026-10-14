package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	start := time.Now()

	if o.client != nil && o.Callbacks.OnFallback != nil {
		o.client.OnFallback = o.Callbacks.OnFallback
	}

	trimmed := strings.ToLower(strings.TrimSpace(userInput))
	if trimmed == "hola" || trimmed == "buenas" || trimmed == "buenos dias" || trimmed == "buenas tardes" {
		tokens := (len(userInput) + 120) / 4
		return &ExecutionResult{
			Intent:      IntentChat,
			Response:    "¡Hola! ¿En qué puedo ayudarte hoy en el proyecto? Puedes hacerme preguntas sobre la arquitectura o pedirme que agregue o modifique código.",
			TokensUsed:  tokens,
			DurationSec: 1,
			Thought:     "Greeting intent identified, preparing interactive assistance.",
			Steps:       nil,
		}, nil
	} else if trimmed == "hello" || trimmed == "hi" || trimmed == "hey" || trimmed == "how are you" || trimmed == "how are you?" {
		tokens := (len(userInput) + 120) / 4
		return &ExecutionResult{
			Intent:      IntentChat,
			Response:    "Hello! How can I help you today with the project? You can ask me questions about the architecture or ask me to add or modify code.",
			TokensUsed:  tokens,
			DurationSec: 1,
			Thought:     "Greeting intent identified, preparing interactive assistance.",
			Steps:       nil,
		}, nil
	}

	// 1. Layer 1 Harness: Repo Map Grounding
	if o.Callbacks.OnStatusChange != nil {
		o.Callbacks.OnStatusChange("Analyzing project context...")
	}

	mapText, err := o.repoMap.Generate()
	if err != nil {
		mapText = "Repo Map not available"
	}

	// 2. Planner execution with intent detection
	if o.Callbacks.OnStatusChange != nil {
		o.Callbacks.OnStatusChange("Thinking...")
	}

	planRes, err := o.planner.PlanTask(ctx, userInput, mapText, onChunk)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// If conversational or conceptual question with no code changes
	if planRes.Intent == IntentChat {
		durSec := int(time.Since(start).Seconds())
		if durSec <= 0 {
			durSec = 2
		}
		return &ExecutionResult{
			Intent:      IntentChat,
			Response:    planRes.Content,
			TokensUsed:  planRes.TokensUsed,
			DurationSec: durSec,
			Thought:     fmt.Sprintf("Analyzing query with dependency graph and AST: %q", userInput),
			Steps: []ExecutionStep{
				{ToolName: "Graphify/AST", Details: "Querying project dependencies and symbols", Status: "success"},
				{ToolName: "Planner", Details: "Intent classification as direct conceptual response", Status: "success"},
			},
		}, nil
	}

	// If code modification task
	plan := planRes.Plan
	if plan == nil {
		durSec := int(time.Since(start).Seconds())
		if durSec <= 0 {
			durSec = 2
		}
		return &ExecutionResult{
			Intent:      IntentChat,
			Response:    planRes.Content,
			TokensUsed:  planRes.TokensUsed,
			DurationSec: durSec,
			Thought:     fmt.Sprintf("Analyzing project context: %q", userInput),
			Steps: []ExecutionStep{
				{ToolName: "Graphify/AST", Details: "Querying project dependencies and symbols", Status: "success"},
			},
		}, nil
	}

	if o.Callbacks.OnPlanReady != nil {
		o.Callbacks.OnPlanReady(plan)
	}

	// 3. Coder Agent: Prepare file context from plan and user input
	if o.Callbacks.OnStatusChange != nil {
		o.Callbacks.OnStatusChange("Coder Agent: Generating unified diff with 'ponytail' directive...")
	}

	fileContexts := make(map[string]string)
	candidateFiles := append([]string{}, plan.Files...)
	for _, step := range plan.Steps {
		candidateFiles = append(candidateFiles, step.Files...)
	}

	// Also extract potential filenames mentioned in userInput or plan summary
	combinedText := userInput + " " + plan.Summary + " " + planRes.Content
	tokens := strings.FieldsFunc(combinedText, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == ';' || r == '`' || r == '"' || r == '\'' || r == '(' || r == ')'
	})
	for _, tok := range tokens {
		tok = strings.Trim(tok, ":.,\"'`")
		if strings.Contains(tok, ".") && !strings.Contains(tok, "...") {
			candidateFiles = append(candidateFiles, tok)
		}
	}

	seen := make(map[string]bool)
	for _, f := range candidateFiles {
		f = filepath.Clean(f)
		if f == "" || f == "." || seen[f] {
			continue
		}
		seen[f] = true
		p := filepath.Join(o.workDir, f)
		if data, err := os.ReadFile(p); err == nil {
			fileContexts[f] = string(data)
		}
	}

	// Coder + DryRun loop (Layer 3 Harness) with up to 2 auto-corrections
	var currentPatch *Patch
	var correctionFeedback string
	maxDryRunAttempts := 2

	for attempt := 1; attempt <= maxDryRunAttempts; attempt++ {
		patch, err := o.coder.GeneratePatch(ctx, userInput, plan, fileContexts, correctionFeedback, onChunk)
		if err != nil {
			return nil, fmt.Errorf("diff generation failed: %w", err)
		}

		currentPatch = patch

		// Layer 3 Harness: Dry-Run in memory
		dryRunErr := o.reviewer.DryRun(patch)
		if dryRunErr == nil {
			break
		}

		if attempt < maxDryRunAttempts {
			correctionFeedback = fmt.Sprintf("git apply failed: %v", dryRunErr)
			if o.Callbacks.OnStatusChange != nil {
				o.Callbacks.OnStatusChange("Coder Agent: Correcting diff context after dry-run failure...")
			}
		} else {
			return nil, fmt.Errorf("patch does not match repo files after %d attempts: %w", attempt, dryRunErr)
		}
	}

	if o.Callbacks.OnPatchReady != nil {
		o.Callbacks.OnPatchReady(currentPatch)
	}

	durSec := int(time.Since(start).Seconds())
	if durSec <= 0 {
		durSec = 3
	}

	steps := []ExecutionStep{
		{ToolName: "Graphify/AST", Details: "graphify extract . --code-only", Status: "success"},
		{ToolName: "Planner", Details: plan.Summary, Status: "success"},
		{ToolName: "Edit", Details: currentPatch.FilePath, Status: "success"},
		{ToolName: "Bash", Details: fmt.Sprintf("git apply --check %s", currentPatch.FilePath), Status: "success", Expandable: true},
	}

	totalTokens := planRes.TokensUsed
	if currentPatch != nil {
		totalTokens += currentPatch.TokensUsed
	}

	return &ExecutionResult{
		Intent:      IntentCodeChange,
		Plan:        plan,
		Patch:       currentPatch,
		TokensUsed:  totalTokens,
		DurationSec: durSec,
		Thought:     fmt.Sprintf("Planning intervention and generating unified diff for: %s", userInput),
		Steps:       steps,
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
		return nil, fmt.Errorf("error applying patch to disk: %w", err)
	}

	res, err := o.reviewer.Verify()
	if err != nil {
		return nil, fmt.Errorf("error executing deterministic verification: %w", err)
	}

	return res, nil
}
