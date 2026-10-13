package harness

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type VerificationResult struct {
	Success bool
	Output  string
	Command string
}

type Verifier struct {
	workDir string
}

func NewVerifier(workDir string) *Verifier {
	if workDir == "" {
		workDir = "."
	}
	return &Verifier{workDir: workDir}
}

// DetectProjectCommand returns the appropriate compiler/linter command based on project files
func (v *Verifier) DetectProjectCommand() (string, []string) {
	// Check for Go
	if _, err := os.Stat(filepath.Join(v.workDir, "go.mod")); err == nil {
		return "go", []string{"vet", "./..."}
	}
	// Check for TypeScript/Node
	if _, err := os.Stat(filepath.Join(v.workDir, "tsconfig.json")); err == nil {
		return "npx", []string{"tsc", "--noEmit"}
	}
	// Check for Python
	if _, err := os.Stat(filepath.Join(v.workDir, "pyproject.toml")); err == nil {
		return "ruff", []string{"check", "."}
	}
	// Check for Rust
	if _, err := os.Stat(filepath.Join(v.workDir, "Cargo.toml")); err == nil {
		return "cargo", []string{"check"}
	}

	return "", nil
}

// RunVerification executes the deterministic verification command and trims redundant stacktraces
func (v *Verifier) RunVerification() (*VerificationResult, error) {
	binary, args := v.DetectProjectCommand()
	if binary == "" {
		return &VerificationResult{
			Success: true,
			Output:  "No specific compiler/linter detected; skipping verification.",
			Command: "none",
		}, nil
	}

	cmdStr := fmt.Sprintf("%s %s", binary, strings.Join(args, " "))
	cmd := exec.Command(binary, args...)
	cmd.Dir = v.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	combined := strings.TrimSpace(stdout.String() + "\n" + stderr.String())

	// Context compression: prune excessive output to avoid burning tokens
	prunedOutput := v.pruneOutput(combined, 50)

	return &VerificationResult{
		Success: err == nil,
		Output:  prunedOutput,
		Command: cmdStr,
	}, nil
}

// pruneOutput keeps only relevant failure lines and truncates huge dumps
func (v *Verifier) pruneOutput(out string, maxLines int) string {
	lines := strings.Split(out, "\n")
	if len(lines) <= maxLines {
		return out
	}
	// Keep the first few and last few lines
	half := maxLines / 2
	head := lines[:half]
	tail := lines[len(lines)-half:]

	var sb strings.Builder
	sb.WriteString(strings.Join(head, "\n"))
	sb.WriteString(fmt.Sprintf("\n... [%d lines truncated by svyaz context compressor] ...\n", len(lines)-maxLines))
	sb.WriteString(strings.Join(tail, "\n"))
	return sb.String()
}
