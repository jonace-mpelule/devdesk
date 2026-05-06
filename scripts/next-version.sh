#!/usr/bin/env sh
set -eu

bump="${1:-patch}"
latest="$(git tag --list 'v[0-9]*' --sort=-v:refname | head -n 1)"

if [ -z "$latest" ]; then
	printf '%s\n' "v0.1.0"
	exit 0
fi

version="${latest#v}"
major="${version%%.*}"
rest="${version#*.}"
minor="${rest%%.*}"
patch="${rest#*.}"
patch="${patch%%[-+]*}"

case "$bump" in
	major)
		major=$((major + 1))
		minor=0
		patch=0
		;;
	minor)
		minor=$((minor + 1))
		patch=0
		;;
	patch)
		patch=$((patch + 1))
		;;
	*)
		printf 'unknown bump "%s"; use major, minor, or patch\n' "$bump" >&2
		exit 1
		;;
esac

printf 'v%s.%s.%s\n' "$major" "$minor" "$patch"
