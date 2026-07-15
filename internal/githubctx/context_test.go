package githubctx

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadParsesPullRequestEvent(t *testing.T) {
	t.Parallel()

	ctx, err := Load(func(key string) string {
		return map[string]string{
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_REPOSITORY": "jyablonski/actions",
			"GITHUB_EVENT_PATH": "/event.json",
			"GITHUB_SERVER_URL": "https://github.com",
			"GITHUB_RUN_ID":     "42",
		}[key]
	}, func(string) ([]byte, error) {
		return []byte(`{"number":12,"pull_request":{"title":"Add notifications","html_url":"https://github.com/jyablonski/actions/pull/12","draft":false,"user":{"login":"jacob"},"head":{"ref":"feature/notifications"},"base":{"ref":"main"}}}`), nil
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if ctx.PullRequest.Number != 12 || ctx.PullRequest.HeadBranch != "feature/notifications" {
		t.Fatalf("PullRequest = %#v", ctx.PullRequest)
	}
	if ctx.RunURL != "https://github.com/jyablonski/actions/actions/runs/42" {
		t.Fatalf("RunURL = %q", ctx.RunURL)
	}
}

func TestLoadWithoutEventPayload(t *testing.T) {
	t.Parallel()

	ctx, err := Load(func(key string) string {
		return map[string]string{
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_REPOSITORY": "jyablonski/actions",
			"GITHUB_REF":        "refs/heads/main",
			"GITHUB_SHA":        "abcdef123",
			"GITHUB_ACTOR":      "jacob",
			"GITHUB_WORKFLOW":   "CI",
		}[key]
	}, func(string) ([]byte, error) {
		t.Fatal("readFile should not be called without GITHUB_EVENT_PATH")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if ctx.RunURL != "" || ctx.Ref != "refs/heads/main" || ctx.Workflow != "CI" {
		t.Fatalf("Context = %#v", ctx)
	}
}

func TestLoadRejectsInvalidContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		getenv   func(string) string
		readFile func(string) ([]byte, error)
		want     string
	}{
		{
			name:     "missing repository",
			getenv:   func(string) string { return "" },
			readFile: func(string) ([]byte, error) { return nil, nil },
			want:     "GITHUB_REPOSITORY is required",
		},
		{
			name: "event read error",
			getenv: func(key string) string {
				if key == "GITHUB_REPOSITORY" {
					return "jyablonski/actions"
				}
				if key == "GITHUB_EVENT_PATH" {
					return "/event.json"
				}
				return ""
			},
			readFile: func(string) ([]byte, error) { return nil, errors.New("not found") },
			want:     "read GITHUB_EVENT_PATH: not found",
		},
		{
			name: "invalid event JSON",
			getenv: func(key string) string {
				if key == "GITHUB_REPOSITORY" {
					return "jyablonski/actions"
				}
				if key == "GITHUB_EVENT_PATH" {
					return "/event.json"
				}
				return ""
			},
			readFile: func(string) ([]byte, error) { return []byte("{"), nil },
			want:     "parse GitHub event payload",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(test.getenv, test.readFile)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want %q", err, test.want)
			}
		})
	}
}
