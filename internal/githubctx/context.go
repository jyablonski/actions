package githubctx

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Context struct {
	EventName   string
	Repository  string
	Ref         string
	SHA         string
	Actor       string
	Workflow    string
	RunURL      string
	PullRequest PullRequest
}

type PullRequest struct {
	Number     int
	Title      string
	URL        string
	HeadBranch string
	BaseBranch string
	Draft      bool
	Author     string
}

type eventPayload struct {
	Number      int `json:"number"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		Draft   bool   `json:"draft"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
		Head struct {
			Ref string `json:"ref"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
}

func Load(getenv func(string) string, readFile func(string) ([]byte, error)) (Context, error) {
	ctx := Context{
		EventName:  getenv("GITHUB_EVENT_NAME"),
		Repository: getenv("GITHUB_REPOSITORY"),
		Ref:        getenv("GITHUB_REF"),
		SHA:        getenv("GITHUB_SHA"),
		Actor:      getenv("GITHUB_ACTOR"),
		Workflow:   getenv("GITHUB_WORKFLOW"),
	}
	if ctx.Repository == "" {
		return Context{}, errors.New("GITHUB_REPOSITORY is required")
	}
	if serverURL, runID := getenv("GITHUB_SERVER_URL"), getenv("GITHUB_RUN_ID"); serverURL != "" && runID != "" {
		ctx.RunURL = fmt.Sprintf("%s/%s/actions/runs/%s", strings.TrimRight(serverURL, "/"), ctx.Repository, runID)
	}

	eventPath := getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return ctx, nil
	}
	payloadBytes, err := readFile(eventPath)
	if err != nil {
		return Context{}, fmt.Errorf("read GITHUB_EVENT_PATH: %w", err)
	}
	var payload eventPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return Context{}, fmt.Errorf("parse GitHub event payload: %w", err)
	}
	ctx.PullRequest = PullRequest{
		Number:     payload.Number,
		Title:      payload.PullRequest.Title,
		URL:        payload.PullRequest.HTMLURL,
		HeadBranch: payload.PullRequest.Head.Ref,
		BaseBranch: payload.PullRequest.Base.Ref,
		Draft:      payload.PullRequest.Draft,
		Author:     payload.PullRequest.User.Login,
	}
	return ctx, nil
}
