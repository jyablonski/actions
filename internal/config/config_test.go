package config

import "testing"

func TestParseUsesEnvironmentWebhook(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{"--template", "pr-opened"}, func(key string) string {
		if key == "SLACK_WEBHOOK" {
			return "https://hooks.slack.com/services/example"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.SlackWebhook != "https://hooks.slack.com/services/example" {
		t.Fatalf("SlackWebhook = %q", cfg.SlackWebhook)
	}
}

func TestParseAcceptsAllInputs(t *testing.T) {
	t.Parallel()

	cfg, err := Parse([]string{
		"--template", "deployment-success",
		"--slack-webhook", "https://hooks.slack.com/services/example",
		"--mention", "<!channel>",
		"--deployment-url", "https://example.com/deployments/42",
		"--failed-jobs", "test",
		"--summary", "completed",
	}, func(string) string { return "" })
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.Template != "deployment-success" || cfg.Mention != "<!channel>" || cfg.DeploymentURL != "https://example.com/deployments/42" {
		t.Fatalf("Config = %#v", cfg)
	}
}

func TestParseRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "missing template", args: []string{"--slack-webhook", "https://hooks.slack.com/services/example"}},
		{name: "missing webhook", args: []string{"--template", "pr-opened"}},
		{name: "non HTTPS webhook", args: []string{"--template", "pr-opened", "--slack-webhook", "http://example.com"}},
		{name: "unknown template", args: []string{"--template", "unknown", "--slack-webhook", "https://hooks.slack.com/services/example"}},
		{name: "invalid deployment URL", args: []string{"--template", "deployment-success", "--slack-webhook", "https://hooks.slack.com/services/example", "--deployment-url", "http://example.com"}},
		{name: "removed environment flag", args: []string{"--template", "deployment-success", "--slack-webhook", "https://hooks.slack.com/services/example", "--environment", "production"}},
		{name: "unexpected argument", args: []string{"--template", "pr-opened", "--slack-webhook", "https://hooks.slack.com/services/example", "extra"}},
		{name: "invalid flag", args: []string{"--not-a-flag"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Parse(test.args, func(string) string { return "" }); err == nil {
				t.Fatal("Parse() error = nil")
			}
		})
	}
}
