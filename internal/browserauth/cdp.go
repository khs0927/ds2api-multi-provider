package browserauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

type cdpClient struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	nextID int
}

type cdpResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *cdpError       `json:"error,omitempty"`
}

type cdpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func dialCDP(ctx context.Context, endpoint string) (*cdpClient, error) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || (u.Scheme != "ws" && u.Scheme != "wss") || u.Host == "" {
		return nil, errors.New("invalid browser websocket endpoint")
	}
	originScheme := "https"
	if u.Scheme == "ws" {
		originScheme = "http"
	}
	origin := originScheme + "://" + u.Host
	config, err := websocket.NewConfig(u.String(), origin)
	if err != nil {
		return nil, err
	}
	conn, err := websocket.DialConfig(config)
	if err != nil {
		return nil, fmt.Errorf("connect browser websocket: %w", err)
	}
	client := &cdpClient{conn: conn}
	go func() {
		<-ctx.Done()
		_ = client.Close()
	}()
	return client, nil
}

func (c *cdpClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *cdpClient) call(method string, params any, sessionID string, out any) error {
	if c == nil || c.conn == nil {
		return errors.New("browser websocket is not connected")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	id := c.nextID
	req := map[string]any{"id": id, "method": method}
	if params != nil {
		req["params"] = params
	}
	if strings.TrimSpace(sessionID) != "" {
		req["sessionId"] = sessionID
	}
	if err := websocket.JSON.Send(c.conn, req); err != nil {
		return err
	}
	for {
		var resp cdpResponse
		if err := websocket.JSON.Receive(c.conn, &resp); err != nil {
			return err
		}
		if resp.ID != id {
			continue
		}
		if resp.Error != nil {
			return fmt.Errorf("cdp %s failed: %s", method, strings.TrimSpace(resp.Error.Message))
		}
		if out == nil || len(resp.Result) == 0 {
			return nil
		}
		return json.Unmarshal(resp.Result, out)
	}
}

func (c *cdpClient) createPage(targetURL string) (string, error) {
	var created struct {
		TargetID string `json:"targetId"`
	}
	if err := c.call("Target.createTarget", map[string]any{"url": targetURL}, "", &created); err != nil {
		return "", err
	}
	if strings.TrimSpace(created.TargetID) == "" {
		return "", errors.New("browser did not create a page target")
	}
	return c.attach(created.TargetID)
}

func (c *cdpClient) attach(targetID string) (string, error) {
	var attached struct {
		SessionID string `json:"sessionId"`
	}
	if err := c.call("Target.attachToTarget", map[string]any{"targetId": targetID, "flatten": true}, "", &attached); err != nil {
		return "", err
	}
	if strings.TrimSpace(attached.SessionID) == "" {
		return "", errors.New("browser did not return a target session")
	}
	return attached.SessionID, nil
}

func (c *cdpClient) findPageSession() (string, error) {
	var targets struct {
		TargetInfos []struct {
			TargetID string `json:"targetId"`
			Type     string `json:"type"`
			URL      string `json:"url"`
		} `json:"targetInfos"`
	}
	if err := c.call("Target.getTargets", nil, "", &targets); err != nil {
		return "", err
	}
	var fallback string
	for _, target := range targets.TargetInfos {
		if target.Type != "page" || strings.TrimSpace(target.TargetID) == "" {
			continue
		}
		if fallback == "" {
			fallback = target.TargetID
		}
		if strings.Contains(target.URL, "deepseek.com") || strings.Contains(target.URL, "accounts.google.com") {
			return c.attach(target.TargetID)
		}
	}
	if fallback == "" {
		return "", errors.New("browser page target not found")
	}
	return c.attach(fallback)
}
