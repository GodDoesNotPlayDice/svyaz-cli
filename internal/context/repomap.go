package context

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// RepoMap generates a concise ground-truth symbol map of the repository
type RepoMap struct {
	rootDir   string
	maxTokens int
}

// NewRepoMap creates a RepoMap builder for the specified directory
func NewRepoMap(rootDir string) *RepoMap {
	if rootDir == "" {
		rootDir = "."
	}
	return &RepoMap{
		rootDir:   rootDir,
		maxTokens: 2000,
	}
}

// Generate scans the repository and produces a compact Markdown summary of symbols and dependencies
func (rm *RepoMap) Generate() (string, error) {
	var sb strings.Builder
	sb.WriteString("=== REPOSITORY GROUNDING MAP (Capa 1 Arnés) ===\n")

	// 1. Read dependencies
	deps := rm.extractDependencies()
	if deps != "" {
		sb.WriteString("### Dependencias Reales Detectadas:\n")
		sb.WriteString(deps)
		sb.WriteString("\n\n")
	}

	// 2. Scan Go symbols using AST parser
	symbols := rm.extractGoSymbols()
	if symbols != "" {
		sb.WriteString("### Símbolos y Signaturas del Proyecto:\n")
		sb.WriteString(symbols)
		sb.WriteString("\n")
	}

	sb.WriteString("===============================================")
	return sb.String(), nil
}

func (rm *RepoMap) extractDependencies() string {
	goModPath := filepath.Join(rm.rootDir, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		lines := strings.Split(string(data), "\n")
		var out []string
		inRequire := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "module ") {
				out = append(out, fmt.Sprintf("- %s", trimmed))
			}
			if strings.HasPrefix(trimmed, "require (") {
				inRequire = true
				continue
			}
			if inRequire {
				if trimmed == ")" {
					inRequire = false
					continue
				}
				if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
					parts := strings.Fields(trimmed)
					if len(parts) >= 2 {
						out = append(out, fmt.Sprintf("  • %s (%s)", parts[0], parts[1]))
					}
				}
			}
		}
		if len(out) > 0 {
			return strings.Join(out, "\n")
		}
	}
	return ""
}

func (rm *RepoMap) extractGoSymbols() string {
	fset := token.NewFileSet()
	var fileSummaries []string

	err := filepath.Walk(rm.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden, vendors, node_modules, tests
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return nil
		}

		relPath, _ := filepath.Rel(rm.rootDir, path)
		var symbols []string

		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if ast.IsExported(x.Name.Name) {
					switch x.Type.(type) {
					case *ast.StructType:
						symbols = append(symbols, fmt.Sprintf("    type %s struct", x.Name.Name))
					case *ast.InterfaceType:
						symbols = append(symbols, fmt.Sprintf("    type %s interface", x.Name.Name))
					default:
						symbols = append(symbols, fmt.Sprintf("    type %s", x.Name.Name))
					}
				}
			case *ast.FuncDecl:
				if ast.IsExported(x.Name.Name) {
					var recv string
					if x.Recv != nil && len(x.Recv.List) > 0 {
						recv = "(recv) "
					}
					symbols = append(symbols, fmt.Sprintf("    func %s%s(...)", recv, x.Name.Name))
				}
			}
			return true
		})

		if len(symbols) > 0 {
			fileSummaries = append(fileSummaries, fmt.Sprintf("• %s:\n%s", relPath, strings.Join(symbols, "\n")))
		}

		return nil
	})

	if err != nil || len(fileSummaries) == 0 {
		return ""
	}

	return strings.Join(fileSummaries, "\n")
}
