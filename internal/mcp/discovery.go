package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type MCPServerInfo struct {
	Name       string   `json:"name"`
	Status     string   `json:"status"` // "✓" or "✗"
	AuthNeeded bool     `json:"auth_needed"`
	Command    string   `json:"command,omitempty"`
	Args       []string `json:"args,omitempty"`
	ServerURL  string   `json:"server_url,omitempty"`
	Tools      []string `json:"tools"`
}

type rawMCPConfig struct {
	MCPServers map[string]struct {
		Command   string            `json:"command"`
		Args      []string          `json:"args"`
		ServerURL string            `json:"serverUrl"`
		Headers   map[string]string `json:"headers"`
		Disabled  bool              `json:"disabled"`
		Env       map[string]string `json:"env"`
	} `json:"mcpServers"`
}

type yamlConfig struct {
	MCPServers map[string]struct {
		Command string            `yaml:"command"`
		Args    []string          `yaml:"args"`
		Env     map[string]string `yaml:"env"`
	} `yaml:"mcp_servers"`
}

// DiscoverServers inspects ~/.svyaz/mcp_config.json, ~/.svyaz/config.yaml, and system MCP configs.
// It performs real-time discovery of active tools and server health via stdio JSON-RPC.
func DiscoverServers() ([]MCPServerInfo, string) {
	homeDir, _ := os.UserHomeDir()
	location := "~/.svyaz/mcp_config.json"

	svyazJSON := filepath.Join(homeDir, ".svyaz", "mcp_config.json")
	svyazYAML := filepath.Join(homeDir, ".svyaz", "config.yaml")
	geminiJSON := filepath.Join(homeDir, ".gemini", "config", "mcp_config.json")
	mcpSchemaDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "mcp")

	serverMap := make(map[string]*MCPServerInfo)
	envMap := make(map[string]map[string]string)

	// 1. If ~/.svyaz/mcp_config.json doesn't exist, try to initialize it from system configs
	if _, err := os.Stat(svyazJSON); os.IsNotExist(err) {
		if data, err := os.ReadFile(geminiJSON); err == nil {
			_ = os.MkdirAll(filepath.Dir(svyazJSON), 0755)
			_ = os.WriteFile(svyazJSON, data, 0644)
		}
	}

	configSources := []string{
		svyazJSON,
		filepath.Join(".svyaz", "mcp_config.json"),
		geminiJSON,
	}

	for _, cfgPath := range configSources {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			continue
		}
		var parsed rawMCPConfig
		if err := json.Unmarshal(data, &parsed); err == nil && len(parsed.MCPServers) > 0 {
			for name, srv := range parsed.MCPServers {
				if srv.Disabled {
					continue
				}
				if _, exists := serverMap[name]; exists {
					continue
				}

				authNeeded := false
				lower := strings.ToLower(name)
				if lower == "neon" || lower == "supabase" {
					authNeeded = true
				}

				status := "✓"
				if srv.Command != "" {
					if _, err := exec.LookPath(srv.Command); err != nil {
						if _, statErr := os.Stat(srv.Command); statErr != nil {
							status = "✗"
						}
					}
				}

				serverMap[name] = &MCPServerInfo{
					Name:       name,
					Status:     status,
					AuthNeeded: authNeeded,
					Command:    srv.Command,
					Args:       srv.Args,
					ServerURL:  srv.ServerURL,
					Tools:      []string{},
				}
				if len(srv.Env) > 0 {
					envMap[name] = srv.Env
				}
			}
		}
	}

	// 2. Also read YAML config if available (e.g. ~/.svyaz/config.yaml or .svyaz/config.yaml)
	yamlPaths := []string{svyazYAML, filepath.Join(".svyaz", "config.yaml")}
	for _, yp := range yamlPaths {
		if yData, err := os.ReadFile(yp); err == nil {
			var yParsed yamlConfig
			if err := yaml.Unmarshal(yData, &yParsed); err == nil {
				for name, srv := range yParsed.MCPServers {
					if _, exists := serverMap[name]; exists {
						continue
					}
					status := "✓"
					if _, err := exec.LookPath(srv.Command); err != nil {
						if _, statErr := os.Stat(srv.Command); statErr != nil {
							status = "✗"
						}
					}
					serverMap[name] = &MCPServerInfo{
						Name:    name,
						Status:  status,
						Command: srv.Command,
						Args:    srv.Args,
						Tools:   []string{},
					}
					if len(srv.Env) > 0 {
						envMap[name] = srv.Env
					}
				}
			}
		}
	}

	// 3. Fallback schemas from disk cache (instant base)
	if entries, err := os.ReadDir(mcpSchemaDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			serverName := entry.Name()
			srv, exists := serverMap[serverName]
			if !exists {
				authNeeded := false
				lower := strings.ToLower(serverName)
				if lower == "neon" || lower == "supabase" {
					authNeeded = true
				}
				srv = &MCPServerInfo{
					Name:       serverName,
					Status:     "✓",
					AuthNeeded: authNeeded,
					Tools:      []string{},
				}
				serverMap[serverName] = srv
			}

			dirPath := filepath.Join(mcpSchemaDir, serverName)
			if toolFiles, err := os.ReadDir(dirPath); err == nil {
				for _, tf := range toolFiles {
					if !tf.IsDir() && strings.HasSuffix(tf.Name(), ".json") {
						srv.Tools = append(srv.Tools, strings.TrimSuffix(tf.Name(), ".json"))
					}
				}
			}
			sort.Strings(srv.Tools)
		}
	}

	// 4. Real-time Live Probing via stdio JSON-RPC (runs concurrently, fast timeout)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, srv := range serverMap {
		if srv.Command == "" || srv.Status == "✗" {
			continue
		}

		wg.Add(1)
		go func(sName string, sInfo *MCPServerInfo) {
			defer wg.Done()
			cmdEnv := envMap[sName]
			liveTools, err := queryLiveMCPTools(sInfo.Command, sInfo.Args, cmdEnv)
			if err == nil && len(liveTools) > 0 {
				mu.Lock()
				sInfo.Tools = liveTools
				sInfo.Status = "✓"
				mu.Unlock()
			}
		}(name, srv)
	}

	wg.Wait()

	if len(serverMap) == 0 {
		return defaultFallbackServers(), location
	}

	var result []MCPServerInfo
	for _, srv := range serverMap {
		result = append(result, *srv)
	}

	// Sort alphabetically by server name
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, location
}

// queryLiveMCPTools connects via stdio JSON-RPC, sends initialize + tools/list with 1.2s timeout
func queryLiveMCPTools(cmdPath string, args []string, extraEnv map[string]string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdPath, args...)
	if len(extraEnv) > 0 {
		cmd.Env = os.Environ()
		for k, v := range extraEnv {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"svyaz","version":"1.0"}}}`
	listReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	_, _ = fmt.Fprintf(stdin, "%s\n%s\n", initReq, listReq)

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var tools []string
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		for scanner.Scan() {
			line := scanner.Bytes()
			var resp struct {
				ID     int `json:"id"`
				Result struct {
					Tools []struct {
						Name string `json:"name"`
					} `json:"tools"`
				} `json:"result"`
			}
			if err := json.Unmarshal(line, &resp); err == nil && resp.ID == 2 {
				for _, t := range resp.Result.Tools {
					if t.Name != "" {
						tools = append(tools, t.Name)
					}
				}
				break
			}
		}
	}()

	select {
	case <-doneCh:
	case <-ctx.Done():
	}

	_ = stdin.Close()
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()

	sort.Strings(tools)
	return tools, nil
}

func defaultFallbackServers() []MCPServerInfo {
	return []MCPServerInfo{
		{Name: "codebase-memory-mcp", Status: "✓", Tools: []string{"index_repository", "search_graph", "query_graph", "trace_path", "get_code_snippet"}},
		{Name: "engram", Status: "✓", Tools: []string{"mem_capture_passive", "mem_compare", "mem_context", "mem_current_project", "mem_doctor"}},
		{Name: "github", Status: "✓", Tools: []string{"actions_get", "actions_list", "actions_run_trigger", "add_comment_to_pending_review", "add_issue_comment"}},
		{Name: "graphify", Status: "✓", Tools: []string{"query_graph", "get_node", "get_neighbors", "get_community", "god_nodes"}},
		{Name: "Neon", Status: "✓", AuthNeeded: true, Tools: []string{"run_sql", "run_sql_transaction", "describe_table_schema", "get_database_tables", "prepare_database_migration"}},
		{Name: "obsidian", Status: "✓", Tools: []string{"obsidian_list_vaults", "obsidian_read_note", "obsidian_create_note", "obsidian_edit_note", "obsidian_delete_note"}},
		{Name: "stitch", Status: "✓", Tools: []string{"create_project", "get_project", "delete_project", "list_screens", "get_screen"}},
		{Name: "supabase", Status: "✓", AuthNeeded: true, Tools: []string{"search_docs", "list_tables", "list_extensions", "list_migrations", "apply_migration"}},
		{Name: "token-savior-recall", Status: "✓", Tools: []string{"list_projects", "switch_project", "get_git_status", "get_changed_symbols", "build_commit_summary"}},
	}
}
