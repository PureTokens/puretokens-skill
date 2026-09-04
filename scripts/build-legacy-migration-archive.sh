#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
archive_path="$repository_root/dist/puretokens-skill-install.zip"
staging_root=$(mktemp -d "${TMPDIR:-/tmp}/puretokens-skill-legacy-migration.XXXXXX")

cleanup() {
  rm -rf -- "$staging_root"
}

trap cleanup EXIT HUP INT TERM

source_root="$staging_root/puretokens-skill-main"
mkdir -p "$source_root"
cp "$repository_root/README.md" "$source_root/README.md"
cp "$repository_root/package.json" "$source_root/package.json"
cp -R "$repository_root/runtime" "$source_root/runtime"
cp -R "$repository_root/skills" "$source_root/skills"

test ! -e "$source_root/runtime/runtime.json" || { printf '%s\n' "legacy migration archive source must not contain runtime/runtime.json" >&2; exit 1; }
test ! -e "$source_root/runtime/puretokens-direct-api.mjs" || { printf '%s\n' "legacy migration archive source must not contain a direct API executor" >&2; exit 1; }
test ! -e "$source_root/skills/puretokens-image/references/model-selection.json" || { printf '%s\n' "legacy migration archive source must not contain retired image model selection" >&2; exit 1; }
test ! -e "$source_root/skills/puretokens-video/references/model-selection.json" || { printf '%s\n' "legacy migration archive source must not contain retired video model selection" >&2; exit 1; }

temporary_archive="$staging_root/puretokens-skill-install.zip"
(cd "$staging_root" && zip -qr "$temporary_archive" puretokens-skill-main)
mv "$temporary_archive" "$archive_path"
