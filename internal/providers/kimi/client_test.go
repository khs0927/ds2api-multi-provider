package kimi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestK3PreservesHistoryToolsAndProviderParameters(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":null,"reasoning_content":"think","tool_calls":[{"id":"call_1"}]}}]}`)
	}))
	defer server.Close()

	client := New(server.URL+"/v1", "test-key")
	assistant := json.RawMessage(`{"role":"assistant","content":null,"reasoning_content":"old","tool_calls":[{"id":"old_call"}]}`)
	result, err := client.Chat(context.Background(), Request{
		Model: "kimi-k3",
		Messages: []json.RawMessage{
			json.RawMessage(`{"role":"user","content":"hello"}`),
			assistant,
			json.RawMessage(`{"role":"tool","tool_call_id":"old_call","content":"done"}`),
		},
		Tools:               json.RawMessage(`[{"type":"function","function":{"name":"search"}}]`),
		ToolChoice:          json.RawMessage(`"auto"`),
		ReasoningEffort:     "high",
		MaxCompletionTokens: 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	if received["temperature"] != 1.0 || received["top_p"] != 0.95 {
		t.Fatalf("K3 sampling policy missing: %#v", received)
	}
	if received["reasoning_effort"] != "high" || received["max_completion_tokens"] != float64(123) {
		t.Fatalf("K3 reasoning/token policy missing: %#v", received)
	}
	messages, ok := received["messages"].([]any)
	if !ok || len(messages) != 3 {
		t.Fatalf("history was not preserved: %#v", received["messages"])
	}
	if !strings.Contains(string(result.Raw), "reasoning_content") || !strings.Contains(string(result.Raw), "tool_calls") {
		t.Fatalf("raw assistant fields were not preserved: %s", result.Raw)
	}
}

func TestK3RejectsNonExactModelIDs(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()
	_, err := New(server.URL+"/v1", "test-key").Chat(context.Background(), Request{
		Model:    " kimi-k3 ",
		Messages: []json.RawMessage{json.RawMessage(`{"role":"user","content":"x"}`)},
	})
	if err == nil || !strings.Contains(err.Error(), "only accepts") {
		t.Fatalf("expected exact model rejection, got %v", err)
	}
	if called {
		t.Fatal("non-exact model must not reach the upstream")
	}
}

func TestK3StreamKeepsRawDataFrames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"think\"}}]}\n\n")
		io.WriteString(w, "event: delta\n")
		io.WriteString(w, "data: {\"choices\":[\n")
		io.WriteString(w, "data: {\"delta\":{\"tool_calls\":[{\"id\":\"c1\"}]}}\n\n")
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()
	result, err := New(server.URL+"/v1", "").Chat(context.Background(), Request{
		Model: "kimi-k3", Messages: []json.RawMessage{json.RawMessage(`{"role":"user","content":"x"}`)}, Stream: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Stream || result.Body == nil {
		t.Fatalf("expected live stream body, got %#v", result)
	}
	var frames [][]byte
	if err := result.ReadStream(func(frame []byte) error {
		frames = append(frames, append([]byte(nil), frame...))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 {
		t.Fatalf("stream frames = %#v", frames)
	}
	if !strings.Contains(string(frames[0]), "reasoning_content") || !strings.Contains(string(frames[1]), "tool_calls") {
		t.Fatalf("stream fields lost: %#v", frames)
	}
	if !strings.Contains(string(frames[1]), "\n") {
		t.Fatalf("multi-line SSE data was not preserved: %#v", frames[1])
	}
}

func TestK3RedactsHTTPErrorBodies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "password=DO_NOT_LEAK", http.StatusUnauthorized)
	}))
	defer server.Close()
	_, err := New(server.URL+"/v1", "test-key").Chat(context.Background(), Request{
		Model:    "kimi-k3",
		Messages: []json.RawMessage{json.RawMessage(`{"role":"user","content":"x"}`)},
	})
	if err == nil {
		t.Fatal("expected HTTP error")
	}
	if !strings.Contains(err.Error(), "K3 API HTTP 401") || strings.Contains(err.Error(), "DO_NOT_LEAK") {
		t.Fatalf("unsafe HTTP error: %v", err)
	}
	var providerErr *Error
	if !errors.As(err, &providerErr) || providerErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected typed HTTP status error, got %T %v", err, err)
	}
}
