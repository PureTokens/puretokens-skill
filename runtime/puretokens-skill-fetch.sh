#!/bin/sh
# Native download/update entry point. Installation writes belong to sync only.
set -eu
umask 077

fail() { printf '%s\n' "Pure Tokens Skill download: $*" >&2; exit 1; }
command_name=${1:-}
case "$command_name" in check-update|install|update) ;; *) fail "use check-update, install or update with --host or --target" ;; esac
shift
host=
target=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --host) [ "$#" -ge 2 ] || fail "--host requires a value"; host=$2; shift 2 ;;
    --target) [ "$#" -ge 2 ] || fail "--target requires a value"; target=$2; shift 2 ;;
    *) fail "unsupported option" ;;
  esac
done
bootstrap_root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
[ -n "$host" ] || [ -n "$target" ] || fail "--host or --target is required"
command -v curl >/dev/null 2>&1 || fail "curl is unavailable; use a host session with native HTTPS download support"
download_root=$(mktemp -d "${TMPDIR:-/tmp}/puretokens-download.XXXXXX") || fail "cannot create a private download directory"
trap 'rm -rf -- "$download_root"' EXIT
trap 'exit 130' INT
trap 'exit 129' HUP
trap 'exit 143' TERM

download() {
  # URLs are constructed below from fixed official origins and validated IDs.
  download_status=$(curl --silent --show-error --location --proto '=https' --proto-redir '=https' --tlsv1.2 --connect-timeout 15 --max-time 180 \
    --header 'Accept: application/vnd.github+json' --user-agent 'puretokens-skill-installer' --output "$2" --write-out '%{http_code}' "$1" 2>/dev/null) ||
    fail "the official source download failed; no installed files were changed"
}
json_string() { sed -n "s/.*\"$2\"[[:space:]]*:[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" "$1" | sed -n '1p'; }
version_valid() { printf '%s\n' "$1" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; }
sha256_file() {
  if command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
  else fail "SHA-256 verification is unavailable"; fi
}

download 'https://api.github.com/repos/PureTokens/puretokens-skill/commits/main' "$download_root/commit.json"
[ "$download_status" = 200 ] || fail "the official main revision could not be resolved"
source_commit=$(json_string "$download_root/commit.json" sha)
printf '%s\n' "$source_commit" | grep -Eq '^[0-9a-f]{40}$' || fail "the official revision is invalid"
download "https://raw.githubusercontent.com/PureTokens/puretokens-skill/$source_commit/package.json" "$download_root/package.json"
[ "$download_status" = 200 ] || fail "the pinned official version could not be read"
available_version=$(json_string "$download_root/package.json" version)
version_valid "$available_version" || fail "the official version is invalid"
location_installer="$bootstrap_root/puretokens-skill-install.sh"
if [ ! -f "$location_installer" ] || ! grep -Fqx '# puretokens-locate-v1' "$location_installer"; then
  location_installer="$download_root/puretokens-skill-install.sh"
  download "https://raw.githubusercontent.com/PureTokens/puretokens-skill/$source_commit/runtime/puretokens-skill-install.sh" "$location_installer"
  [ "$download_status" = 200 ] || fail "the pinned directory selector is unavailable"
fi
set -- locate
[ -z "$host" ] || set -- "$@" --host "$host"
[ -z "$target" ] || set -- "$@" --target "$target"
target=$(sh "$location_installer" "$@") || exit 1
installed_version=not_installed
if [ -f "$target/.puretokens-executor/runtime.json" ]; then
  if [ "$(json_string "$target/.puretokens-executor/runtime.json" name)" = puretokens-api-executor ]; then
    installed_version=$(json_string "$target/.puretokens-executor/runtime.json" version)
    version_valid "$installed_version" || installed_version=unverified
  fi
fi
if [ "$installed_version" = "$available_version" ]; then update_state=current; else update_state=version_differs; fi
printf '%s\n' "Pure Tokens Skills update check: installed=$installed_version available=$available_version status=$update_state source_commit=$source_commit"
[ "$command_name" != check-update ] || exit 0
if version_valid "$installed_version" && awk -v current="$installed_version" -v available="$available_version" 'BEGIN {split(current,c,"."); split(available,a,"."); for(i=1;i<=3;i++){if(c[i]+0>a[i]+0)exit 0;if(c[i]+0<a[i]+0)exit 1}exit 1}'; then
  fail "the installed version is newer than official main; automatic downgrade was stopped"
fi

case "$(uname -s)-$(uname -m)" in
  Darwin-arm64) platform=darwin-arm64 ;;
  Darwin-x86_64) platform=darwin-amd64 ;;
  Linux-x86_64) platform=linux-amd64 ;;
  Linux-aarch64|Linux-arm64) platform=linux-arm64 ;;
  *) fail "this operating system and CPU has no platform executor" ;;
esac
command -v unzip >/dev/null 2>&1 || fail "ZIP extraction is unavailable; use a host session with native ZIP support"
release_origin="https://github.com/PureTokens/puretokens-skill/releases/download/v$available_version"
download "$release_origin/release-manifest.json" "$download_root/release.json"
distribution=source
if [ "$download_status" = 200 ]; then
  release_version=$(json_string "$download_root/release.json" version)
  release_commit=$(json_string "$download_root/release.json" sourceCommit)
  if [ "$release_version" = "$available_version" ] && [ "$release_commit" = "$source_commit" ]; then
    sed -n "/\"$platform\"[[:space:]]*:/,/}/p" "$download_root/release.json" > "$download_root/platform.json"
    filename=$(json_string "$download_root/platform.json" filename)
    checksum=$(json_string "$download_root/platform.json" sha256)
    [ "$filename" = "puretokens-skill-$available_version-$platform.zip" ] || fail "published platform metadata is invalid"
    printf '%s\n' "$checksum" | grep -Eq '^[0-9a-f]{64}$' || fail "published platform checksum is invalid"
    download "$release_origin/$filename" "$download_root/source.zip"
    [ "$download_status" = 200 ] || fail "the published platform archive is unavailable"
    [ "$(sha256_file "$download_root/source.zip")" = "$checksum" ] || fail "published platform archive checksum mismatch"
    distribution=platform
    archive_root=puretokens-skill
  fi
elif [ "$download_status" != 404 ]; then
  fail "published platform metadata could not be checked"
fi
if [ "$distribution" = source ]; then
  printf '%s\n' "Matching published platform assets are unavailable; retrieving the pinned official source archive."
  download "https://codeload.github.com/PureTokens/puretokens-skill/zip/$source_commit" "$download_root/source.zip"
  [ "$download_status" = 200 ] || fail "the pinned official source archive is unavailable"
  archive_root="puretokens-skill-$source_commit"
fi

# Reject escaping paths, unexpected archive roots and symlinks before extraction.
unzip -Z1 "$download_root/source.zip" > "$download_root/entries" 2>/dev/null || fail "the archive is unreadable"
awk -v root="$archive_root/" '
  index($0,root)!=1 || $0 ~ /\\/ || $0 ~ /(^|\/)\.\.?($|\/)/ {exit 1}
  END {if(NR==0)exit 1}
' "$download_root/entries" || fail "the archive contains an unsafe path"
unzip -Z -l "$download_root/source.zip" > "$download_root/archive-info" 2>/dev/null || fail "the archive cannot be inspected"
if awk '$1 ~ /^l/ {found=1} END {exit !found}' "$download_root/archive-info"; then fail "archive symlinks are unsupported"; fi
mkdir "$download_root/unpacked"
unzip -q "$download_root/source.zip" -d "$download_root/unpacked" || fail "the archive could not be extracted"
source_root="$download_root/unpacked/$archive_root"
[ "$(json_string "$source_root/package.json" version)" = "$available_version" ] || fail "the archive version does not match the pinned source"
set -- sync --source "$source_root" --target "$target"
[ -z "$host" ] || set -- "$@" --host "$host"
sh "$source_root/runtime/puretokens-skill-install.sh" "$@"
