package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jyablonski/actions/internal/templates"
)

type Sender interface {
	Send(context.Context, templates.Message) error
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	webhookURL string
	httpClient HTTPDoer
}

func NewClient(webhookURL string) Sender {
	return &Client{webhookURL: webhookURL, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func NewClientWithHTTPClient(webhookURL string, httpClient HTTPDoer) *Client {
	return &Client{webhookURL: webhookURL, httpClient: httpClient}
}

func (c *Client) Send(ctx context.Context, message templates.Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("encode Slack payload: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Slack request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("perform Slack request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4<<10))
	if err != nil {
		return fmt.Errorf("read Slack response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("slack returned HTTP %d: %s", response.StatusCode, string(responseBody))
	}
	return nil
}
