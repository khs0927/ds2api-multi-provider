package kimi

import (
	"bufio"
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
	Body   io.ReadCloser
}

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string { return e.Message }

// ReadStream consumes the response incrementally and invokes onFrame for each
// SSE data event. The caller keeps control of the stream lifetime through this
// method instead of receiving a fully buffered response from Chat.
func (r *Response) ReadStream(onFrame func([]byte) error) (err error) {
	if onFrame == nil {
		return fmt.Errorf("K3 stream callback is nil")
	}
	if r.Body == nil {
		for _, frame := range r.Chunks {
			if err := onFrame(frame); err != nil {
				return err
			}
		}
		return nil
	}
	defer func() {
		closeErr := r.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	reader := bufio.NewReader(r.Body)
	var dataLines []string
	flush := func() error {
		if len(dataLines) == 0 {
			return nil
		}
		frame := []byte(strings.Join(dataLines, "\n"))
		dataLines = nil
		if bytes.Equal(frame, []byte("[DONE]")) {
			return nil
		}
		r.Chunks = append(r.Chunks, append([]byte(nil), frame...))
		return onFrame(frame)
	}
	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		} else if line == "" {
			if flushErr := flush(); flushErr != nil {
				return flushErr
			}
		}
		if readErr == io.EOF {
			return flush()
		}
		if readErr != nil {
			return readErr
		}
	}
}

// Close releases a live stream without reading it to completion.
func (r *Response) Close() error {
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: apiKey, HTTPClient: http.DefaultClient}
}

func (c *Client) Chat(ctx context.Context, req Request) (Response, error) {
	if req.Model != "kimi-k3" && req.Model != "moonshotai/Kimi-K3" {
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return Response{}, &Error{StatusCode: resp.StatusCode, Message: fmt.Sprintf("K3 API HTTP %d", resp.StatusCode)}
	}
	if req.Stream {
		return Response{Stream: true, Body: resp.Body}, nil
	}
	raw, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		return Response{}, fmt.Errorf("read K3 response: %w", err)
	}
	if closeErr != nil {
		return Response{}, fmt.Errorf("close K3 response: %w", closeErr)
	}
	return Response{Raw: raw}, nil
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
