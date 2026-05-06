# DevDesk Developer Docs

This document is for people working on the DevDesk codebase.

## Requirements

- macOS
- Go matching the version in `go.mod`
- Git
- GitHub CLI, only when publishing releases with `make release`

## Common Commands

Run the CLI during development:

```sh
make run
```

Run tests:

```sh
make test
```

Build a local binary:

```sh
make build
```

Install the locally built binary to `~/.local/bin/devdesk`:

```sh
make install-local
```

## Version Metadata

The binary includes version metadata injected at build time:

- `Version`
- `Commit`
- `Date`

Check it with:

```sh
devdesk version
```

When no version is passed explicitly, the Makefile computes the next semantic version from Git tags:

```sh
make version
```

By default, this uses `BUMP=patch`. You can request a different bump:

```sh
make version BUMP=minor
make version BUMP=major
```

If the repository has no `vX.Y.Z` tags yet, the computed version starts at `v0.1.0`.

## Makefile Targets

- `make run`: runs `go run ./cmd/devdesk`.
- `make test`: runs all tests with repo-local Go caches.
- `make build`: builds a local binary into `dist/devdesk`.
- `make build-mac`: builds macOS arm64 and amd64 binaries.
- `make package-mac`: builds macOS binaries, archives them, and writes checksums.
- `make tag`: runs tests and creates an annotated Git tag for the computed version.
- `make release`: packages macOS binaries, creates the tag, and publishes a GitHub release.
- `make install-local`: installs the local binary into `~/.local/bin/devdesk`.
- `make clean`: removes `dist/`.

## Local Caches

The Makefile uses repo-local Go caches:

```text
.gocache/
.gomodcache/
```

These paths are ignored by Git. They avoid permission issues on systems where Go cannot write to the default user cache paths.

## Project Layout

```text
cmd/devdesk/main.go        CLI entrypoint
internal/devdesk/          CLI behavior and macOS workspace logic
scripts/next-version.sh    Computes the next semver tag
scripts/tag-version.sh     Validates and creates release tags
install.sh                 End-user binary installer
```
