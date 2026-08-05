package kimi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Request is deliberately independent from DS2API's DeepSeek prompt model.
// Messages are RawMessage values so K3 assistant history is preserved byte-for-byte.
type Request struct {
	Model               string
	Messages            []json.RawMessage
	Tools               json.RawMessage
	ToolChoice          json.RawMessage
	ReasoningEffort     string
	MaxCompletionTokens int
	Stream              bool
}

// Response keeps the exact non-stream body or every raw SSE data frame.
// Higher layers decide how to interpret reasoning_content and tool_calls.
type Response struct {
	Raw    []byte
	Chunks [][]byte
	Stream bool
}

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string { return e.Message }

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: apiKey, HTTPClient: http.DefaultClient}
}

func (c *Client) Chat(ctx context.Context, req Request) (Response, error) {
	if strings.TrimSpace(req.Model) != "kimi-k3" && strings.TrimSpace(req.Model) != "moonshotai/Kimi-K3" {
		return Response{}, fmt.Errorf("kimi provider only accepts kimi-k3 or moonshotai/Kimi-K3")
	}
	if len(req.Messages) == 0 {
		return Response{}, fmt.Errorf("kimi request requires messages")
	}
	effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort))
	if effort == "" {
		effort = "max"
	}
	if effort != "low" && effort != "high" && effort != "max" {
		return Response{}, fmt.Errorf("unsupported K3 reasoning_effort %q", effort)
	}
	payload := map[string]any{
		"model":            req.Model,
		"messages":         req.Messages,
		"reasoning_effort": effort,
		"temperature":      1.0,
		"top_p":            0.95,
		"stream":           req.Stream,
	}
	if len(req.Tools) > 0 {
		payload["tools"] = req.Tools
	}
	if len(req.ToolChoice) > 0 {
		payload["tool_choice"] = req.ToolChoice
	}
	if req.MaxCompletionTokens > 0 {
		payload["max_completion_tokens"] = req.MaxCompletionTokens
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("marshal K3 request: %w", err)
	}
	endpoint, err := url.JoinPath(c.BaseURL, "chat/completions")
	if err != nil {
		return Response{}, fmt.Errorf("build K3 endpoint: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("build K3 request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if strings.TrimSpace(c.APIKey) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("K3 request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read K3 response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, &Error{StatusCode: resp.StatusCode, Message: fmt.Sprintf("K3 API HTTP %d", resp.StatusCode)}
	}
	if !req.Stream {
		return Response{Raw: raw}, nil
	}
	return Response{Raw: raw, Chunks: sseDataFrames(raw), Stream: true}, nil
}

func sseDataFrames(raw []byte) [][]byte {
	var frames [][]byte
	for _, line := range bytes.Split(raw, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("data:")) {
			value := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
			if len(value) > 0 && !bytes.Equal(value, []byte("[DONE]")) {
				frames = append(frames, append([]byte(nil), value...))
			}
		}
	}
	return frames
}
