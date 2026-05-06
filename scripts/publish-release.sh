#!/usr/bin/env sh
set -eu

version="${1:-}"
bump="${2:-patch}"

if [ -z "$version" ]; then
	version="$(./scripts/next-version.sh "$bump")"
fi

case "$version" in
	[0-9]*.[0-9]*.[0-9]*) version="v$version" ;;
	v[0-9]*.[0-9]*.[0-9]*) ;;
	*)
		printf 'version must look like vX.Y.Z, got "%s"\n' "$version" >&2
		exit 1
		;;
esac

printf 'Publishing DevDesk %s\n' "$version"

if ! git diff --quiet || ! git diff --cached --quiet; then
	printf 'working tree has uncommitted changes; commit them before publishing\n' >&2
	exit 1
fi

make test
make package-mac VERSION="$version"

if git rev-parse "$version" >/dev/null 2>&1; then
	tag_commit="$(git rev-list -n 1 "$version")"
	head_commit="$(git rev-parse HEAD)"
	if [ "$tag_commit" != "$head_commit" ]; then
		printf 'tag "%s" already exists on a different commit\n' "$version" >&2
		exit 1
	fi
	printf 'Using existing tag %s on HEAD\n' "$version"
else
	./scripts/tag-version.sh "$version"
fi

git push origin "$version"

if gh release view "$version" >/dev/null 2>&1; then
	printf 'GitHub release "%s" already exists\n' "$version" >&2
	exit 1
fi

gh release create "$version" dist/*.tar.gz dist/checksums.txt --title "$version" --notes "DevDesk $version"
printf 'Published DevDesk %s\n' "$version"
