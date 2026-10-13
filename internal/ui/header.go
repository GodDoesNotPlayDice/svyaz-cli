package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HeaderInfo struct {
	Version      string
	WorkDir      string
	CurrentModel string
	ModelStatus  string
}

// GetActiveWorkDir returns the current working directory path
func GetActiveWorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// RenderHeader renders the Antigravity-style header with chain links logo and 3 metadata lines
func RenderHeader(info HeaderInfo, width int) string {
	// 1. Chain links logo (5 lines)
	logoLines := []string{
		LogoChainStyle.Render("┌────┐  ┌────┐"),
		LogoChainStyle.Render("│ ┌──┼──┼──┐ │"),
		LogoChainStyle.Render("│ └──┼──┼──┘ │"),
		LogoChainStyle.Render("└────┘  └────┘"),
		LogoTextStyle.Render("   S V Y A Z  "),
	}
	logoBlock := strings.Join(logoLines, "\n")

	// 2. Metadata (3 lines)
	version := info.Version
	if version == "" {
		version = "v1.0.0"
	}
	model := info.CurrentModel
	if model == "" {
		model = "deepseek-coder"
	}
	path := info.WorkDir
	if path == "" {
		path = GetActiveWorkDir()
	}

	line1 := MetaTitleStyle.Render("Svyaz CLI ") + MetaVersionStyle.Render(version)
	line2 := MetaLabelStyle.Render("Model: ") + MetaModelStyle.Render(model)
	line3 := MetaLabelStyle.Render("Path:  ") + MetaPathStyle.Render(path)

	metaBlock := lipgloss.NewStyle().
		PaddingLeft(3).
		Render(fmt.Sprintf("%s\n%s\n%s", line1, line2, line3))

	// Combine Logo + Metadata horizontally (aligned to top)
	headerContent := lipgloss.JoinHorizontal(lipgloss.Top, logoBlock, metaBlock)

	// Outer padding around header
	paddedHeader := lipgloss.NewStyle().
		Padding(1, 2, 1, 2).
		Render(headerContent)

	// Horizontal separator line spanning full width
	sepWidth := width
	if sepWidth <= 0 {
		sepWidth = 80
	}
	separator := SeparatorLineStyle.Render(strings.Repeat("─", sepWidth))

	return paddedHeader + "\n" + separator
}
