# Contributing

## Development

You need the Go toolchain declared in [`.tool-versions`](.tool-versions) (`asdf install` or `mise install`).
Any Go release supported upstream works too: Go downloads the right toolchain automatically.

```
go run .                 # play
go test -race ./...      # test
gofmt -l . && go vet ./...
```

Pull requests run the same checks in CI, plus `go mod tidy -diff`, `govulncheck` and a multi-platform image build.

### Project layout

| Path | Content |
|------|---------|
| `main.go` | Flags, environment variables and wiring |
| `internal/stratagem` | Stratagem list and icons, embedded in the binary |
| `internal/keys` | Keyboard layouts |
| `internal/terminal` | Color detection and the kitty graphics protocol |
| `internal/ui` | Game screen: bubbletea model and lipgloss styles |
| `internal/buildinfo` | Version string |
| `cmd/update-stratagems` | Regenerates the stratagem list and icons from the Helldivers Wiki |

`main.go` stays at the root so that `go install github.com/paulden/libertea@latest` keeps working.

## Continuous integration

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| `CI` | pull requests, pushes to `main` | Checks, then builds the image. Pushes `ghcr.io/paulden/libertea:main`, `:latest` and `:sha-<commit>` from `main` only. |
| `Release` | `v*` tags | Checks, then publishes binaries and multi-platform images with GoReleaser. |
| `Weekly checks` | every Monday | Runs the checks to catch new vulnerabilities when nothing is pushed. |
| `Image cleanup` | every Monday, manual | Deletes untagged images and keeps the 10 newest `sha-*` images. Manual runs are dry runs by default. |

`Checks` and `Image` are reusable workflows called by the ones above.

## Releasing

Tag the commit on `main` and push the tag:

```
git tag -a v1.2.3 -m "v1.2.3"
git push origin v1.2.3
```

The image gets the tags `1.2.3`, `1.2` and `1`, `latest` follows `main`.
Pre-releases such as `v1.3.0-rc.1` only get their own image tag, and are published as GitHub pre-releases.

Binaries and images come with build provenance attestations stored by GitHub, see `gh attestation verify` in the README.

## Dependency updates

Updates are proposed by [Renovate](https://docs.renovatebot.com/), configured in [`renovate.json`](renovate.json).
The Renovate GitHub app must be installed on the repository.

### What is tracked

| Group | Files | Update |
|-------|-------|--------|
| Go toolchain | `go.mod` (`toolchain`), `Dockerfile`, `.tool-versions` | one pull request for the three files |
| Go modules | `go.mod`, `go.sum` | minor and patch grouped, majors one by one |
| GitHub Actions | `.github/workflows/*.yml` | pinned by commit digest |
| Base image | `Dockerfile`, `goreleaser.Dockerfile` | pinned by digest |

The `go` directive in `go.mod` is the oldest Go release still supported upstream.
It is not updated by Renovate: raise it by hand when a Go release reaches its end of life,
that is when a new major Go release comes out, every February and August.

### Schedule

- Regular updates: pull requests are opened on Monday mornings, see the Dependency Dashboard issue for the queue.
- Vulnerabilities: Renovate opens a pull request labelled `security` as soon as an advisory is published,
  and `govulncheck` fails the checks when vulnerable code is reachable.

### Reviewing an update

1. Read the changelog linked in the pull request, look for breaking changes and deprecations.
2. Wait for the checks and the image build to pass.
3. For TUI libraries (`charmbracelet/*`, `muesli/*`), run the game once: rendering changes are not covered by tests.
4. Merge with a merge commit.
