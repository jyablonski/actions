package slack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jyablonski/actions/internal/templates"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) Do(request *http.Request) (*http.Response, error) { return f(request) }

func TestClientSend(t *testing.T) {
	t.Parallel()

	client := NewClientWithHTTPClient("https://hooks.slack.com/services/example", roundTripper(func(request *http.Request) (*http.Response, error) {
		if got := request.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}
		var payload templates.Message
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if len(payload.Attachments) != 1 || payload.Attachments[0].Color != templates.SuccessColor {
			t.Fatalf("attachments = %#v", payload.Attachments)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	}))
	if err := client.Send(context.Background(), templates.Message{Text: "hello", Attachments: []templates.Attachment{{Color: templates.SuccessColor, Blocks: []templates.Block{{Type: "section"}}}}}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	client, ok := NewClient("https://hooks.slack.com/services/example").(*Client)
	if !ok {
		t.Fatal("NewClient() did not return a *Client")
	}
	if client.webhookURL != "https://hooks.slack.com/services/example" || client.httpClient == nil {
		t.Fatalf("Client = %#v", client)
	}
}

func TestClientSendReturnsSlackError(t *testing.T) {
	t.Parallel()

	client := NewClientWithHTTPClient("https://hooks.slack.com/services/example", roundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("invalid_payload"))}, nil
	}))
	if err := client.Send(context.Background(), templates.Message{Text: "hello"}); err == nil || !strings.Contains(err.Error(), "invalid_payload") {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestClientSendReturnsTransportErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		doer HTTPDoer
		want string
	}{
		{
			name: "invalid request URL",
			url:  "://invalid",
			doer: roundTripper(func(*http.Request) (*http.Response, error) {
				t.Fatal("Do should not be called")
				return nil, nil
			}),
			want: "create Slack request",
		},
		{
			name: "transport failure",
			url:  "https://hooks.slack.com/services/example",
			doer: roundTripper(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("network unavailable")
			}),
			want: "perform Slack request: network unavailable",
		},
		{
			name: "response read failure",
			url:  "https://hooks.slack.com/services/example",
			doer: roundTripper(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: errReadCloser{}}, nil
			}),
			want: "read Slack response: read failure",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			client := NewClientWithHTTPClient(test.url, test.doer)
			err := client.Send(context.Background(), templates.Message{Text: "hello"})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Send() error = %v, want %q", err, test.want)
			}
		})
	}
}

type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) { return 0, errors.New("read failure") }
func (errReadCloser) Close() error             { return nil }
