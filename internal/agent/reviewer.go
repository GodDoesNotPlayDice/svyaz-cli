package agent

import (
	"fmt"

	"github.com/xaxa/svyaz-cli/internal/harness"
)

type ReviewerAgent struct {
	patcher  *harness.Patcher
	verifier *harness.Verifier
}

func NewReviewerAgent(workDir string) *ReviewerAgent {
	return &ReviewerAgent{
		patcher:  harness.NewPatcher(workDir),
		verifier: harness.NewVerifier(workDir),
	}
}

// DryRun checks the patch in memory without writing to disk (Capa 3 Arnés)
func (r *ReviewerAgent) DryRun(patch *Patch) error {
	if patch == nil || patch.Diff == "" {
		return fmt.Errorf("patch is empty")
	}
	return r.patcher.DryRun(patch.Diff)
}

// Apply writes the patch to disk
func (r *ReviewerAgent) Apply(patch *Patch) error {
	if patch == nil || patch.Diff == "" {
		return fmt.Errorf("patch is empty")
	}
	return r.patcher.Apply(patch.Diff)
}

// Verify runs the deterministic compiler/linter check (Capa 4 Arnés)
func (r *ReviewerAgent) Verify() (*harness.VerificationResult, error) {
	return r.verifier.RunVerification()
}
