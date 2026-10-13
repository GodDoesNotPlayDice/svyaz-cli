package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HeaderInfo struct {
	Version      string
	RepoName     string
	GitBranch    string
	CurrentModel string
	TokensUsed   int
	TokenLimit   int
}

// GetCurrentGitBranch retrieves the active git branch name cleanly
func GetCurrentGitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetCurrentDirName gets the basename of current directory
func GetCurrentDirName() string {
	wd, err := os.Getwd()
	if err != nil {
		return "repo"
	}
	parts := strings.Split(strings.TrimRight(wd, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "repo"
}

// RenderHeader generates the sober Antigravity-like top bar
func RenderHeader(info HeaderInfo, width int) string {
	brand := BrandStyle.Render("svyaz") + " " + VersionStyle.Render(info.Version)

	repoText := info.RepoName
	if info.GitBranch != "" {
		repoText = fmt.Sprintf("%s (%s)", info.RepoName, info.GitBranch)
	}
	repoPart := HeaderItemStyle.Render("repo: ") + HeaderItemValueStyle.Render(repoText)

	modelPart := HeaderItemStyle.Render("model: ") + HeaderAccentValueStyle.Render(info.CurrentModel)

	tokenPart := HeaderItemStyle.Render("tokens: ") + HeaderItemValueStyle.Render(
		fmt.Sprintf("%.1fk/%.0fk", float64(info.TokensUsed)/1000.0, float64(info.TokenLimit)/1000.0),
	)

	sep := HeaderSeparatorStyle.String()
	content := lipgloss.JoinHorizontal(
		lipgloss.Center,
		brand,
		sep,
		repoPart,
		sep,
		modelPart,
		sep,
		tokenPart,
	)

	if width > 0 {
		return HeaderContainerStyle.Width(width).Render(content)
	}
	return HeaderContainerStyle.Render(content)
}
