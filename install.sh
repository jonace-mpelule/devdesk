#!/usr/bin/env sh
set -eu

repo="${DEVDESK_REPO:-jonace-mpelule/devdesk}"
install_dir="${DEVDESK_INSTALL_DIR:-$HOME/.local/bin}"
version="${DEVDESK_VERSION:-latest}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
	darwin) os="darwin" ;;
	*)
		printf 'devdesk installer currently supports macOS only\n' >&2
		exit 1
		;;
esac

case "$arch" in
	arm64|aarch64) arch="arm64" ;;
	x86_64|amd64) arch="amd64" ;;
	*)
		printf 'unsupported architecture: %s\n' "$arch" >&2
		exit 1
		;;
esac

if [ "$version" = "latest" ]; then
	api_url="https://api.github.com/repos/$repo/releases/latest"
	version="$(curl -fsSL "$api_url" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
fi

if [ -z "$version" ]; then
	printf 'could not resolve devdesk release version\n' >&2
	exit 1
fi

asset="devdesk_${version}_${os}_${arch}.tar.gz"
url="https://github.com/$repo/releases/download/$version/$asset"
tmp_dir="$(mktemp -d)"

cleanup() {
	rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

printf 'Downloading %s\n' "$url"
curl -fsSL "$url" -o "$tmp_dir/$asset"
tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"
mkdir -p "$install_dir"
cp "$tmp_dir/devdesk_${version}_${os}_${arch}/devdesk" "$install_dir/devdesk"
chmod +x "$install_dir/devdesk"

printf 'Installed devdesk %s to %s/devdesk\n' "$version" "$install_dir"
printf 'Make sure %s is in your PATH.\n' "$install_dir"
