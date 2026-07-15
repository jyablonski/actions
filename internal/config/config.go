package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/jyablonski/actions/internal/templates"
)

type Config struct {
	Template      string
	SlackWebhook  string
	Mention       string
	DeploymentURL string
	FailedJobs    string
	Summary       string
}

func Parse(args []string, getenv func(string) string) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet("notify", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.Template, "template", "", "notification template")
	flags.StringVar(&cfg.SlackWebhook, "slack-webhook", getenv("SLACK_WEBHOOK"), "Slack incoming webhook URL")
	flags.StringVar(&cfg.Mention, "mention", "", "Slack mention to prepend")
	flags.StringVar(&cfg.DeploymentURL, "deployment-url", "", "deployment URL")
	flags.StringVar(&cfg.FailedJobs, "failed-jobs", "", "comma-separated failed jobs")
	flags.StringVar(&cfg.Summary, "summary", "", "optional notification summary")
	if err := flags.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}
	if flags.NArg() != 0 {
		return Config{}, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if cfg.Template == "" {
		return Config{}, errors.New("--template is required")
	}
	if !templates.Known(cfg.Template) {
		return Config{}, fmt.Errorf("unknown template %q; supported templates: %s", cfg.Template, templates.Names())
	}
	if cfg.SlackWebhook == "" {
		return Config{}, errors.New("--slack-webhook is required (or set SLACK_WEBHOOK)")
	}
	webhookURL, err := url.ParseRequestURI(cfg.SlackWebhook)
	if err != nil || webhookURL.Scheme != "https" || webhookURL.Host == "" {
		return Config{}, errors.New("--slack-webhook must be a valid HTTPS URL")
	}
	if cfg.DeploymentURL != "" {
		deploymentURL, err := url.ParseRequestURI(cfg.DeploymentURL)
		if err != nil || deploymentURL.Scheme != "https" || deploymentURL.Host == "" {
			return Config{}, errors.New("--deployment-url must be a valid HTTPS URL")
		}
	}
	return cfg, nil
}
