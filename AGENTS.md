# Repository Guide

## Purpose

This repository publishes reusable GitHub Actions under `jyablonski/actions`.

Go-backed actions share the root Go module and one prebuilt GHCR image; simple runner-native actions remain Bash composite actions.

## Layout

- `cmd/notify` is the single Go entrypoint and selects behavior with `--template`.
- `internal/config`, `internal/githubctx`, `internal/slack`, and `internal/templates` contain private Go code; do not add `pkg/` APIs.
- `pr-notification` and `deploy-notification` are thin Docker action adapters that use the shared image.
- `.github/workflows/ci.yml` runs pull-request validation, and `release.yml` publishes the container on version tags.

## Go conventions

- Use Go `1.26.5`; keep the version aligned in `go.mod`, CI, and release workflows.
- Keep notification behavior behind named templates such as `pr-opened`, `pipeline-failure`, and `deployment-success`.
- Keep external HTTP calls behind interfaces so they can be unit tested without network access.
- Return contextual errors and never log or embed webhook URLs, tokens, or other secrets.
- Add focused unit tests for normal behavior and error paths with every Go behavior change.

## Action conventions

- Keep `action.yml` files thin: declare inputs and forward them as arguments to the shared command.
- Docker actions must reference `ghcr.io/jyablonski/actions-go` by immutable digest; never build Go source while a consumer workflow runs.
- Use the root action directories in consumer references, for example `jyablonski/actions/pr-notification@v1`.
- Composite actions run scripts directly on the runner and should use `bash` with strict error handling.
- Do not commit binaries, generated coverage files, secrets, or real Slack webhook URLs.

## Validation

- Run `make ci` before handing off Go, workflow, or script changes.
- `make test` uses gotestsum and must remain at or above 90% total statement coverage.
- `make deadcode` and `make lint` use the same pinned tools as CI.
- Run `make help` to see the local build and validation commands.

## Releases and versioning

- Tags matching `v*` publish the Linux amd64 image to GHCR and report its digest.
- After a release, update Docker action metadata to the published digest while consumers continue using the moving major action tag such as `@v1`.
- Do not run the Makefile `release` or `sync-v1` targets, create tags, push branches, or change package visibility unless the user explicitly requests it.

## Documentation style

- Do not hard-wrap prose: keep each paragraph and list item on one logical line.
- Update README examples whenever a public action directory, input, or consumer workflow changes.
