# svyaz-cli (Связь)

> **CLI agéntico ligero y de alto rendimiento compilado en Go**, diseñado para programar con asistencia agéntica, soporte nativo de **Model Context Protocol (MCP)**, **Skills modulares** y optimizado para operar sin fricción con **FreeLLMAPI**.

---

## ⚡ Características Principales

- **Cero dependencias de runtime:** Compilado como un binario estático único en Go.
- **TUI Sobria (Antigravity-inspired):** Interfaz limpia, minimalista y libre de estridencias construida con Bubble Tea y Lip Gloss.
- **Arnés Determinista de 4 Capas:**
  1. **Grounding:** Repo Map y dependencias declaradas.
  2. **MCP Gate:** Validación estricta de JSON Schema en Go para llamadas a herramientas.
  3. **Dry-Run Gate:** `git apply --check` en memoria antes de tocar disco.
  4. **Linter Loop:** Bucle de compilación determinista (`go vet`, `tsc`, `ruff`).
- **Arquitectura de 3 Sub-Agentes:** Planner (exploración AST y grafos, 0 código), Coder (parches con regla *ponytail*), y Reviewer (aprobación humana y verificación).
- **Economía de Tokens para FreeLLMAPI:** Integración con Token Savior, Graphify y Engram.

---

## 🛠 Instalación y Compilación

```bash
# Compilar binario nativo
go build -o svyaz .

# Modo interactivo TUI
./svyaz chat

# Modo headless / batch
./svyaz run "Corrige el cálculo de métricas"

# Inspección de MCPs
./svyaz mcp list
```
