package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIError represents an HTTP or JSON error returned by the LLM provider
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("llm api error (HTTP %d): %s", e.StatusCode, e.Message)
}

// FallbackEvent notifies listeners that a model was rotated due to rate limits or failure
type FallbackEvent struct {
	FailedModel string
	NextModel   string
	Reason      string
	StatusCode  int
}

// Client interacts with OpenAI-compatible endpoints with SSE streaming and automatic model fallback
type Client struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	pool         *ModelPool
	maxRetries   int
	OnFallback   func(event FallbackEvent)
}

// NewClient creates an LLM Client
func NewClient(baseURL, apiKey string, timeout time.Duration, pool *ModelPool) *Client {
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	cleanURL := strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL: cleanURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		pool:       pool,
		maxRetries: 4,
	}
}

// FetchModels loads available models from GET /v1/models and syncs them into the ModelPool
func (c *Client) FetchModels(ctx context.Context) ([]ModelMetadata, error) {
	url := fmt.Sprintf("%s/models", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for /models: %w", err)
	}

	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching /models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var list ModelsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decoding /models response: %w", err)
	}

	if c.pool != nil {
		c.pool.UpdateAvailableModels(list.Data)
	}

	return list.Data, nil
}

// CreateChatCompletion sends a single-turn request to /v1/chat/completions (non-streaming)
func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling chat request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building chat completion request: %w", err)
	}

	c.setHeaders(httpReq)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling /chat/completions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var completion ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return nil, fmt.Errorf("decoding chat completion response: %w", err)
	}

	return &completion, nil
}

// CreateChatCompletionStream streams responses using Server-Sent Events (SSE)
func (c *Client) CreateChatCompletionStream(
	ctx context.Context,
	req ChatCompletionRequest,
	onDelta func(delta ChatChunkDelta) error,
) (*ChatCompletionResponse, error) {
	req.Stream = true
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling stream request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building stream request: %w", err)
	}

	c.setHeaders(httpReq)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling stream /chat/completions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	scanner := bufio.NewScanner(resp.Body)
	// Buffer up to 1MB per line if needed
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var assembledContent strings.Builder
	var assembledToolCalls []ToolCall
	var responseModel string
	var responseID string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue // comment or keep-alive
		}

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		var chunk ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if responseID == "" {
			responseID = chunk.ID
		}
		if responseModel == "" {
			responseModel = chunk.Model
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				assembledContent.WriteString(delta.Content)
			}
			if len(delta.ToolCalls) > 0 {
				assembledToolCalls = append(assembledToolCalls, delta.ToolCalls...)
			}
			if onDelta != nil {
				if err := onDelta(delta); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stream chunks: %w", err)
	}

	assembled := &ChatCompletionResponse{
		ID:    responseID,
		Model: responseModel,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:      RoleAssistant,
					Content:   assembledContent.String(),
					ToolCalls: assembledToolCalls,
				},
				FinishReason: "stop",
			},
		},
	}

	return assembled, nil
}

// ExecuteWithFallback runs a chat completion, automatically rotating models upon encountering HTTP 429 or 5xx
func (c *Client) ExecuteWithFallback(
	ctx context.Context,
	req ChatCompletionRequest,
	onChunk func(token string),
) (*ChatCompletionResponse, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("model pool is nil")
	}

	requiresTools := len(req.Tools) > 0
	attempts := 0
	maxAttempts := c.maxRetries
	if maxAttempts < len(c.pool.candidates) {
		maxAttempts = len(c.pool.candidates)
	}

	for attempts < maxAttempts {
		attempts++

		currentModel, err := c.pool.GetNextAvailableModel(requiresTools, 0)
		if err != nil {
			return nil, fmt.Errorf("exhausted available models in pool: %w", err)
		}

		req.Model = currentModel

		var resp *ChatCompletionResponse
		var callErr error

		if onChunk != nil {
			resp, callErr = c.CreateChatCompletionStream(ctx, req, func(delta ChatChunkDelta) error {
				if delta.Content != "" {
					onChunk(delta.Content)
				}
				return nil
			})
		} else {
			resp, callErr = c.CreateChatCompletion(ctx, req)
		}

		if callErr == nil {
			c.pool.ReportSuccess(currentModel)
			return resp, nil
		}

		// Handle error and circuit breaker
		var statusCode int
		if apiErr, ok := callErr.(*APIError); ok {
			statusCode = apiErr.StatusCode
		}

		c.pool.ReportFailure(currentModel, statusCode)

		// Determine next candidate to notify listener
		nextModel, _ := c.pool.GetNextAvailableModel(requiresTools, 0)
		if c.OnFallback != nil {
			c.OnFallback(FallbackEvent{
				FailedModel: currentModel,
				NextModel:   nextModel,
				Reason:      callErr.Error(),
				StatusCode:  statusCode,
			})
		}
	}

	return nil, fmt.Errorf("failed to complete chat request after %d model fallback attempts", attempts)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}
}
