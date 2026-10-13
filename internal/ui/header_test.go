package ui

import (
	"fmt"
	"testing"
)

func TestRenderHeaderAndFooter(t *testing.T) {
	info := HeaderInfo{
		Version:      "v1.0.0",
		WorkDir:      "/home/xaxa/Documents/Github/svyaz-cli",
		CurrentModel: "deepseek-coder",
		ModelStatus:  "Ready",
	}

	header := RenderHeader(info, 80)
	footer := RenderFooter(info.ModelStatus, 80)

	fmt.Println("--- HEADER PREVIEW ---")
	fmt.Println(header)
	fmt.Println("--- PROMPT PREVIEW ---")
	fmt.Println("> ")
	fmt.Println("--- FOOTER PREVIEW ---")
	fmt.Println(footer)
}
