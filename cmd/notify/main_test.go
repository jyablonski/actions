package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jyablonski/actions/internal/slack"
	"github.com/jyablonski/actions/internal/templates"
)

type fakeSender struct {
	message templates.Message
	err     error
}

func (s *fakeSender) Send(_ context.Context, message templates.Message) error {
	s.message = message
	return s.err
}

func TestRunSelectsPipelineFailureTemplate(t *testing.T) {
	t.Parallel()

	sender := &fakeSender{}
	err := run(
		[]string{"--template", "pipeline-failure", "--slack-webhook", "https://hooks.slack.com/services/test", "--failed-jobs", "test, deploy"},
		func(key string) string {
			return map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_REPOSITORY": "jyablonski/actions",
				"GITHUB_REF":        "refs/heads/main",
				"GITHUB_SHA":        "abcdef1234567890",
				"GITHUB_ACTOR":      "jacob",
				"GITHUB_WORKFLOW":   "CI",
				"GITHUB_RUN_ID":     "123",
				"GITHUB_SERVER_URL": "https://github.com",
			}[key]
		},
		func(string) ([]byte, error) { return nil, errors.New("not used") },
		func(string) slack.Sender { return sender },
		io.Discard,
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(sender.message.Text, "Main pipeline failed") {
		t.Fatalf("message text = %q", sender.message.Text)
	}
}

func TestRunReportsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		getenv   func(string) string
		readFile fileReader
		sender   *fakeSender
		want     string
	}{
		{
			name:   "invalid configuration",
			args:   []string{"--template", "unknown", "--slack-webhook", "https://hooks.slack.com/services/test"},
			getenv: func(string) string { return "" },
			want:   "unknown template",
		},
		{
			name: "GitHub context error",
			args: []string{"--template", "pipeline-failure", "--slack-webhook", "https://hooks.slack.com/services/test"},
			getenv: func(key string) string {
				if key == "GITHUB_REPOSITORY" {
					return ""
				}
				return ""
			},
			want: "GITHUB_REPOSITORY is required",
		},
		{
			name:   "template error",
			args:   []string{"--template", "pr-opened", "--slack-webhook", "https://hooks.slack.com/services/test"},
			getenv: baseEnvironment,
			want:   "pr-opened requires a pull_request event payload",
		},
		{
			name:   "Slack send error",
			args:   []string{"--template", "pipeline-failure", "--slack-webhook", "https://hooks.slack.com/services/test"},
			getenv: baseEnvironment,
			sender: &fakeSender{err: errors.New("Slack unavailable")},
			want:   "send Slack notification: Slack unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			sender := test.sender
			if sender == nil {
				sender = &fakeSender{}
			}
			err := run(test.args, test.getenv, test.readFile, func(string) slack.Sender { return sender }, io.Discard)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("run() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRunWritesSuccess(t *testing.T) {
	t.Parallel()

	sender := &fakeSender{}
	var output strings.Builder
	err := run(
		[]string{"--template", "deployment-success", "--slack-webhook", "https://hooks.slack.com/services/test"},
		baseEnvironment,
		func(string) ([]byte, error) { return nil, errors.New("not used") },
		func(string) slack.Sender { return sender },
		&output,
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got := output.String(); got != "Slack notification sent\n" {
		t.Fatalf("output = %q", got)
	}
	if !strings.Contains(sender.message.Text, "Deployment succeeded") {
		t.Fatalf("message text = %q", sender.message.Text)
	}
}

func TestRunDeliversMessageToLocalWebhook(t *testing.T) {
	t.Parallel()

	messages := make(chan templates.Message, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() { _ = request.Body.Close() }()
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want %s", request.Method, http.MethodPost)
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", request.Header.Get("Content-Type"))
		}
		var message templates.Message
		if err := json.NewDecoder(request.Body).Decode(&message); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		messages <- message
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := run(
		[]string{"--template", "pipeline-failure", "--slack-webhook", server.URL, "--failed-jobs", "test"},
		baseEnvironment,
		func(string) ([]byte, error) { return nil, errors.New("not used") },
		func(webhookURL string) slack.Sender {
			return slack.NewClientWithHTTPClient(webhookURL, server.Client())
		},
		io.Discard,
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	message := <-messages
	if !strings.Contains(message.Text, "Main pipeline failed") {
		t.Fatalf("message text = %q", message.Text)
	}
}

func baseEnvironment(key string) string {
	return map[string]string{
		"GITHUB_EVENT_NAME": "push",
		"GITHUB_REPOSITORY": "jyablonski/actions",
		"GITHUB_REF":        "refs/heads/main",
		"GITHUB_SHA":        "abcdef1234567890",
		"GITHUB_ACTOR":      "jacob",
		"GITHUB_WORKFLOW":   "CI",
		"GITHUB_RUN_ID":     "123",
		"GITHUB_SERVER_URL": "https://github.com",
	}[key]
}
