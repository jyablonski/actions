# GitHub Actions

Reusable GitHub Actions maintained by `jyablonski`. Go-backed actions share one Go module and a prebuilt GHCR image; simple actions remain Bash composite actions.

## Go-backed notification actions

Use a moving major tag for the action template:

```yaml
- uses: jyablonski/actions/pr-notification@v1
  with:
    slack-webhook: ${{ secrets.SLACK_WEBHOOK_URL }}
```

The action template is versioned by the moving `v1` major tag, while its internal `docker://` image reference is pinned to an immutable digest. After the first image release, repoint the clearly marked `<PLACEHOLDER_DIGEST>` values in the Go-backed action metadata to the digest emitted by the release workflow.

### Pull request opened

`pr-notification` posts a concise pull request summary. It reads pull request details from the standard GitHub event payload.

See [examples/pr-notification.yaml](examples/pr-notification.yaml) for a complete workflow.

### Post-merge pipeline

`deploy-notification` accepts `template: pipeline-failure` or `template: deployment-success`. Run a final notification job with `if: always()` for failures. Send a deployment-success message only when the deploy job reports that it actually released a deployment.

See [examples/deploy-notification.yaml](examples/deploy-notification.yaml) for failure and conditional deployment-success notifications.

## Development

```sh
make ci
```

Run `make help` to list the available local build and validation commands.

Install the local CI checks with `pre-commit install`. The hooks run `go vet`, gotestsum with a required 90% coverage threshold, deadcode, and the same pinned golangci-lint version used in CI.
