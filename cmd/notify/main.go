package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/jyablonski/actions/internal/config"
	"github.com/jyablonski/actions/internal/githubctx"
	"github.com/jyablonski/actions/internal/slack"
	"github.com/jyablonski/actions/internal/templates"
)

func main() {
	if err := run(os.Args[1:], os.Getenv, os.ReadFile, slack.NewClient, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "notify:", err)
		os.Exit(1)
	}
}

type fileReader func(string) ([]byte, error)
type clientFactory func(string) slack.Sender

func run(args []string, getenv func(string) string, readFile fileReader, newClient clientFactory, output io.Writer) error {
	cfg, err := config.Parse(args, getenv)
	if err != nil {
		return err
	}

	ctx, err := githubctx.Load(getenv, readFile)
	if err != nil {
		return err
	}

	message, err := templates.Render(cfg.Template, ctx, templates.Params{
		Mention:       cfg.Mention,
		DeploymentURL: cfg.DeploymentURL,
		FailedJobs:    cfg.FailedJobs,
		Summary:       cfg.Summary,
	})
	if err != nil {
		return err
	}

	if err := newClient(cfg.SlackWebhook).Send(context.Background(), message); err != nil {
		return fmt.Errorf("send Slack notification: %w", err)
	}

	_, _ = fmt.Fprintln(output, "Slack notification sent")
	return nil
}
