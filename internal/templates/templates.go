package templates

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jyablonski/actions/internal/githubctx"
)

const (
	PROpened          = "pr-opened"
	PipelineFailure   = "pipeline-failure"
	DeploymentSuccess = "deployment-success"
)

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Block struct {
	Type   string `json:"type"`
	Text   *Text  `json:"text,omitempty"`
	Fields []Text `json:"fields,omitempty"`
}

type Attachment struct {
	Color  string  `json:"color"`
	Blocks []Block `json:"blocks"`
}

type Message struct {
	Text        string       `json:"text"`
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Params struct {
	Mention       string
	DeploymentURL string
	FailedJobs    string
	Summary       string
}

func Known(name string) bool {
	switch name {
	case PROpened, PipelineFailure, DeploymentSuccess:
		return true
	default:
		return false
	}
}

func Names() string {
	return strings.Join([]string{PROpened, PipelineFailure, DeploymentSuccess}, ", ")
}

func Render(name string, ctx githubctx.Context, params Params) (Message, error) {
	switch name {
	case PROpened:
		return renderPROpened(ctx, params)
	case PipelineFailure:
		return renderPipelineFailure(ctx, params), nil
	case DeploymentSuccess:
		return renderDeploymentSuccess(ctx, params), nil
	default:
		return Message{}, fmt.Errorf("unknown template %q; supported templates: %s", name, Names())
	}
}

func renderPROpened(ctx githubctx.Context, params Params) (Message, error) {
	pr := ctx.PullRequest
	if pr.Number == 0 || pr.Title == "" || pr.URL == "" {
		return Message{}, errors.New("pr-opened requires a pull_request event payload")
	}
	message := Message{
		Text: fmt.Sprintf("Pull request opened: #%d %s", pr.Number, pr.Title),
		Blocks: []Block{
			header(fmt.Sprintf("🔀 Pull request opened: #%d", pr.Number)),
			section(fmt.Sprintf("*%s*\n%s", escape(pr.Title), pr.URL)),
			fields(
				field("Repository", ctx.Repository),
				field("Branch", fmt.Sprintf("%s → %s", pr.HeadBranch, pr.BaseBranch)),
				field("Opened by", first(pr.Author, ctx.Actor)),
				field("Draft", yesNo(pr.Draft)),
			),
		},
	}
	return withRunLink(message, ctx, params.Mention), nil
}

func renderPipelineFailure(ctx githubctx.Context, params Params) Message {
	failedJobs := first(params.FailedJobs, "See workflow run")
	message := Message{
		Text: fmt.Sprintf("Main pipeline failed: %s", ctx.Repository),
		Blocks: []Block{
			header("🚨 Main pipeline failed"),
			section(fmt.Sprintf("*%s* · %s", escape(ctx.Repository), escape(ctx.Workflow))),
			fields(
				field("Commit", shortSHA(ctx.SHA)),
				field("Ref", ctx.Ref),
				field("Triggered by", ctx.Actor),
				field("Failed jobs", failedJobs),
			),
		},
	}
	if params.Summary != "" {
		message.Blocks = append(message.Blocks, section(escape(params.Summary)))
	}
	return withColor(withRunLink(message, ctx, params.Mention), "danger")
}

func renderDeploymentSuccess(ctx githubctx.Context, params Params) Message {
	message := Message{
		Text: fmt.Sprintf("Deployment succeeded: %s", ctx.Repository),
		Blocks: []Block{
			header("🚀 Deployment succeeded"),
			section(fmt.Sprintf("*%s* · %s", escape(ctx.Repository), escape(ctx.Workflow))),
			fields(
				field("Version", shortSHA(ctx.SHA)),
				field("Ref", ctx.Ref),
				field("Triggered by", ctx.Actor),
			),
		},
	}
	if params.DeploymentURL != "" {
		message.Blocks = append(message.Blocks, section(fmt.Sprintf("<%s|View deployment>", params.DeploymentURL)))
	}
	return withColor(withRunLink(message, ctx, params.Mention), "good")
}

func withRunLink(message Message, ctx githubctx.Context, mention string) Message {
	if mention != "" {
		message.Blocks = append(message.Blocks, section(mention))
	}
	if ctx.RunURL != "" {
		message.Blocks = append(message.Blocks, section(fmt.Sprintf("<%s|View workflow run>", ctx.RunURL)))
	}
	return message
}

func withColor(message Message, color string) Message {
	message.Attachments = []Attachment{{Color: color, Blocks: message.Blocks}}
	message.Blocks = nil
	return message
}

func header(value string) Block {
	return Block{Type: "header", Text: &Text{Type: "plain_text", Text: value}}
}
func section(value string) Block {
	return Block{Type: "section", Text: &Text{Type: "mrkdwn", Text: value}}
}
func fields(values ...Text) Block { return Block{Type: "section", Fields: values} }
func field(label, value string) Text {
	return Text{Type: "mrkdwn", Text: fmt.Sprintf("*%s*\n%s", label, escape(value))}
}
func first(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
func shortSHA(value string) string {
	if len(value) > 7 {
		return value[:7]
	}
	return value
}
func escape(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	return strings.ReplaceAll(value, ">", "&gt;")
}
