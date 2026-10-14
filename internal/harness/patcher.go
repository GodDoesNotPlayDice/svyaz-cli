package harness

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Patcher manages memory-checked and applied git patches
type Patcher struct {
	workDir string
}

func NewPatcher(workDir string) *Patcher {
	if workDir == "" {
		workDir = "."
	}
	return &Patcher{workDir: workDir}
}

// DryRun tests the patch against the working tree in memory without modifying files
func (p *Patcher) DryRun(diffText string) error {
	cmd := exec.Command("git", "apply", "--recount", "--ignore-whitespace", "--check", "-")
	cmd.Dir = p.workDir
	cmd.Stdin = strings.NewReader(diffText)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("patch dry-run failed: %s (err: %w)", strings.TrimSpace(stderr.String()), err)
	}
	return nil
}

// Apply applies the patch directly to the working tree
func (p *Patcher) Apply(diffText string) error {
	// Always dry-run first
	if err := p.DryRun(diffText); err != nil {
		return fmt.Errorf("refusing to apply invalid patch: %w", err)
	}

	cmd := exec.Command("git", "apply", "--recount", "--ignore-whitespace", "-")
	cmd.Dir = p.workDir
	cmd.Stdin = strings.NewReader(diffText)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("applying patch failed: %s (err: %w)", strings.TrimSpace(stderr.String()), err)
	}
	return nil
}
