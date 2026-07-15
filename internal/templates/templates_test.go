package templates

import (
	"strings"
	"testing"

	"github.com/jyablonski/actions/internal/githubctx"
)

func TestRenderPROpened(t *testing.T) {
	t.Parallel()

	message, err := Render(PROpened, githubctx.Context{
		Repository: "jyablonski/actions",
		RunURL:     "https://github.com/jyablonski/actions/actions/runs/1",
		PullRequest: githubctx.PullRequest{
			Number: 5, Title: "Ship <notifications>", URL: "https://github.com/jyablonski/actions/pull/5", HeadBranch: "feature", BaseBranch: "main", Author: "jacob",
		},
	}, Params{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(message.Text, "#5") {
		t.Fatalf("message text = %q", message.Text)
	}
	if !strings.Contains(message.Blocks[1].Text.Text, "&lt;notifications&gt;") {
		t.Fatalf("title block = %q", message.Blocks[1].Text.Text)
	}
}

func TestRenderPROpenedIncludesOptionalFields(t *testing.T) {
	t.Parallel()

	message, err := Render(PROpened, githubctx.Context{
		Repository: "jyablonski/actions",
		Actor:      "jacob",
		RunURL:     "https://github.com/jyablonski/actions/actions/runs/1",
		PullRequest: githubctx.PullRequest{
			Number: 5, Title: "Ship", URL: "https://github.com/jyablonski/actions/pull/5", HeadBranch: "feature", BaseBranch: "main", Draft: true,
		},
	}, Params{Mention: "<!channel>"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := message.Blocks[2].Fields[2].Text; !strings.Contains(got, "jacob") {
		t.Fatalf("opened by field = %q", got)
	}
	if got := message.Blocks[2].Fields[3].Text; !strings.Contains(got, "yes") {
		t.Fatalf("draft field = %q", got)
	}
	if got := message.Blocks[len(message.Blocks)-2].Text.Text; got != "<!channel>" {
		t.Fatalf("mention block = %q", got)
	}
}

func TestRenderPROpenedRequiresPullRequestPayload(t *testing.T) {
	t.Parallel()

	if _, err := Render(PROpened, githubctx.Context{}, Params{}); err == nil || !strings.Contains(err.Error(), "pull_request event") {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestRenderPipelineFailureIncludesFailedJobs(t *testing.T) {
	t.Parallel()

	message, err := Render(PipelineFailure, githubctx.Context{Repository: "jyablonski/actions", SHA: "abcdef123", Ref: "refs/heads/main", Actor: "jacob", Workflow: "CI"}, Params{FailedJobs: "test, deploy"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(message.Attachments) != 1 || message.Attachments[0].Color != "danger" {
		t.Fatalf("attachments = %#v", message.Attachments)
	}
	blocks := message.Attachments[0].Blocks
	if !strings.Contains(blocks[2].Fields[3].Text, "test, deploy") {
		t.Fatalf("failed jobs field = %q", blocks[2].Fields[3].Text)
	}
}

func TestRenderPipelineFailureIncludesDefaultsAndSummary(t *testing.T) {
	t.Parallel()

	message, err := Render(PipelineFailure, githubctx.Context{
		Repository: "jyablonski/actions",
		SHA:        "abcdef1",
		RunURL:     "https://github.com/jyablonski/actions/actions/runs/1",
	}, Params{Mention: "<!here>", Summary: "test <failed>"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	blocks := message.Attachments[0].Blocks
	if got := blocks[2].Fields[3].Text; !strings.Contains(got, "See workflow run") {
		t.Fatalf("failed jobs field = %q", got)
	}
	if got := blocks[3].Text.Text; got != "test &lt;failed&gt;" {
		t.Fatalf("summary block = %q", got)
	}
	if got := blocks[len(blocks)-1].Text.Text; !strings.Contains(got, "View workflow run") {
		t.Fatalf("workflow link = %q", got)
	}
}

func TestRenderDeploymentSuccess(t *testing.T) {
	t.Parallel()

	message, err := Render(DeploymentSuccess, githubctx.Context{
		Repository: "jyablonski/actions",
		SHA:        "abcdef123456",
		Ref:        "refs/heads/main",
		Actor:      "jacob",
		Workflow:   "deploy",
	}, Params{DeploymentURL: "https://example.com/deployments/42"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(message.Attachments) != 1 || message.Attachments[0].Color != "good" {
		t.Fatalf("attachments = %#v", message.Attachments)
	}
	blocks := message.Attachments[0].Blocks
	if got := blocks[2].Fields[0].Text; !strings.Contains(got, "abcdef1") {
		t.Fatalf("version field = %q", got)
	}
	if got := blocks[3].Text.Text; got != "<https://example.com/deployments/42|View deployment>" {
		t.Fatalf("deployment link = %q", got)
	}
}

func TestKnown(t *testing.T) {
	t.Parallel()

	for _, name := range []string{PROpened, PipelineFailure, DeploymentSuccess} {
		if !Known(name) {
			t.Fatalf("Known(%q) = false", name)
		}
	}
	if Known("unknown") {
		t.Fatal("Known(unknown) = true")
	}
}

func TestRenderRejectsUnknownTemplate(t *testing.T) {
	t.Parallel()

	if _, err := Render("unknown", githubctx.Context{}, Params{}); err == nil {
		t.Fatal("Render() error = nil")
	}
}
