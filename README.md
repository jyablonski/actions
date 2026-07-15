# GitHub Actions

Reusable GitHub Actions maintained by `jyablonski`. Go-backed actions share one Go module and a prebuilt GHCR image; simple actions remain Bash composite actions.

## Go-backed notification actions

| Action                                      | Notification template | Required inputs             | Optional inputs                           | Purpose                                    |
| ------------------------------------------- | --------------------- | --------------------------- | ----------------------------------------- | ------------------------------------------ |
| `jyablonski/actions/pr-notification@v1`     | `pr-opened`           | `slack-webhook`             | `slack-mention`                           | Posts a summary when a pull request opens. |
| `jyablonski/actions/deploy-notification@v1` | `pipeline-failure`    | `slack-webhook`, `template` | `failed-jobs`, `slack-mention`, `summary` | Posts when a pipeline job fails.           |
| `jyablonski/actions/deploy-notification@v1` | `deployment-success`  | `slack-webhook`, `template` | `deployment-url`, `slack-mention`         | Posts after a deployment is released.      |

For example, use the pull request notification action in a pull request workflow:

```yaml
- uses: jyablonski/actions/pr-notification@v1
  with:
    slack-webhook: ${{ secrets.SLACK_WEBHOOK_URL }}
    slack-mention: "<@SLACK_USER_ID>"
```

Set `slack-mention` to an individual user ID such as `<@U01234567>` or a user-group ID such as `<!subteam^S01234567>`.

The action template is versioned by the moving `v1` major tag, while its internal `docker://` image reference is pinned to an immutable digest. After the first image release, replace the `<PLACEHOLDER_DIGEST>` values in the Go-backed action metadata with the digest emitted by the release workflow.

## Development

```sh
make test
```

Run `make help` to list the available local build and validation commands.

Install the local CI checks with `pre-commit install`. The hooks run `go vet`, gotestsum with a required 90% coverage threshold, deadcode, and the same pinned golangci-lint version used in CI.
