# DevDesk Deployment Docs

This document covers building, tagging, and publishing DevDesk releases.

## Release Assets

DevDesk releases publish macOS binaries for:

- Apple Silicon: `darwin/arm64`
- Intel Mac: `darwin/amd64`

The installer expects release assets named:

```text
devdesk_<version>_darwin_arm64.tar.gz
devdesk_<version>_darwin_amd64.tar.gz
checksums.txt
```

For example:

```text
devdesk_v0.1.0_darwin_arm64.tar.gz
devdesk_v0.1.0_darwin_amd64.tar.gz
checksums.txt
```

## Versioning

Versions are computed from existing Git tags matching `vX.Y.Z`.

Check the next patch version:

```sh
make version
```

Check a minor or major bump:

```sh
make version BUMP=minor
make version BUMP=major
```

Override the version manually when needed:

```sh
make package-mac VERSION=v1.2.3
```

## Package Locally

Build and package both macOS architectures:

```sh
make package-mac
```

This writes release files into `dist/`:

```text
dist/devdesk_<version>_darwin_arm64.tar.gz
dist/devdesk_<version>_darwin_amd64.tar.gz
dist/checksums.txt
```

## Tag a Release

Create an annotated Git tag for the computed version:

```sh
make tag
```

Use a different automatic bump:

```sh
make tag BUMP=minor
```

`make tag` runs tests first and requires a clean working tree. If the tag is created locally, push it with:

```sh
git push origin <tag>
```

## Publish a GitHub Release

Use one command for the full release flow:

```sh
make publish
```

This resolves one version, runs tests, builds both macOS binaries, packages archives, creates or reuses the matching tag on `HEAD`, pushes the tag, and creates the GitHub release.

Use a specific version:

```sh
make publish VERSION=v0.2.2
```

Use an automatic bump when `HEAD` does not already have a version tag:

```sh
make publish BUMP=minor
```

Do not run `make tag` and then `make release` as separate commands for normal publishing. Creating a tag first changes what the next automatic patch version is.

`make publish` requires:

- Authenticated `gh`
- A clean working tree
- Permission to create releases in `jonace-mpelule/devdesk`

After publishing, verify the public installer against the release:

```sh
DEVDESK_VERSION=<tag> DEVDESK_INSTALL_DIR=/tmp/devdesk-install-test sh ./install.sh
/tmp/devdesk-install-test/devdesk version
```

The installed binary version should match the release tag.

## Installer Contract

The public install command downloads the latest GitHub release:

```sh
curl -fsSL https://raw.githubusercontent.com/jonace-mpelule/devdesk/main/install.sh | sh
```

The installer supports these environment variables:

- `DEVDESK_REPO`: GitHub repository, defaults to `jonace-mpelule/devdesk`.
- `DEVDESK_VERSION`: release tag, defaults to latest.
- `DEVDESK_INSTALL_DIR`: install directory, defaults to `~/.local/bin`.

Example:

```sh
DEVDESK_VERSION=v0.1.0 DEVDESK_INSTALL_DIR=/usr/local/bin sh ./install.sh
```
