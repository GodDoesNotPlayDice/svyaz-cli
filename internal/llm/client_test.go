package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_StreamAndFallback(t *testing.T) {
	// Create mock server
	var requestedModels []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			var req ChatCompletionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			requestedModels = append(requestedModels, req.Model)

			// Model 1 triggers rate limit 429
			if req.Model == "failing-model" {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":{"message":"Rate limit exceeded (429)"}}`))
				return
			}

			// Model 2 succeeds with SSE stream
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Fatalf("expected http.Flusher")
			}

			chunks := []string{"Hola", " desde", " fallback", "!"}
			for _, chunk := range chunks {
				data := fmt.Sprintf(`data: {"id":"c-1","model":"%s","choices":[{"delta":{"content":"%s"}}]}`, req.Model, chunk)
				fmt.Fprintf(w, "%s\n\n", data)
				flusher.Flush()
			}
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	pool := NewModelPool("failing-model", []string{"working-model"}, 1*time.Second)
	client := NewClient(server.URL, "test-api-key", 2*time.Second, pool)

	var fallbackEvents []FallbackEvent
	client.OnFallback = func(ev FallbackEvent) {
		fallbackEvents = append(fallbackEvents, ev)
	}

	var streamedOutput strings.Builder
	resp, err := client.ExecuteWithFallback(context.Background(), ChatCompletionRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "test prompt"},
		},
	}, func(token string) {
		streamedOutput.WriteString(token)
	})

	if err != nil {
		t.Fatalf("expected fallback to succeed, got error: %v", err)
	}

	if resp == nil {
		t.Fatalf("expected valid response")
	}

	expectedText := "Hola desde fallback!"
	if streamedOutput.String() != expectedText {
		t.Errorf("expected streamed text '%s', got '%s'", expectedText, streamedOutput.String())
	}

	if len(requestedModels) != 2 {
		t.Errorf("expected 2 model attempts, got %d: %v", len(requestedModels), requestedModels)
	}

	if len(fallbackEvents) != 1 {
		t.Errorf("expected 1 fallback event, got %d", len(fallbackEvents))
	} else {
		if fallbackEvents[0].FailedModel != "failing-model" || fallbackEvents[0].NextModel != "working-model" {
			t.Errorf("unexpected fallback event: %+v", fallbackEvents[0])
		}
	}

	if !pool.IsOnCooldown("failing-model") {
		t.Errorf("failing-model should be placed on cooldown")
	}
}
