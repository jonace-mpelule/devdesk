#!/usr/bin/env sh
set -eu

version="${1:-}"

if [ -z "$version" ]; then
	printf 'usage: scripts/tag-version.sh vX.Y.Z\n' >&2
	exit 1
fi

case "$version" in
	v[0-9]*.[0-9]*.[0-9]*) ;;
	*)
		printf 'version must look like vX.Y.Z, got "%s"\n' "$version" >&2
		exit 1
		;;
esac

if ! git diff --quiet || ! git diff --cached --quiet; then
	printf 'working tree has uncommitted changes; commit them before tagging\n' >&2
	exit 1
fi

if git rev-parse "$version" >/dev/null 2>&1; then
	printf 'tag "%s" already exists\n' "$version" >&2
	exit 1
fi

git tag -a "$version" -m "Release $version"
printf 'created tag %s\n' "$version"
