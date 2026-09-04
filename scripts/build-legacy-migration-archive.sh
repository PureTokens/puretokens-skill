#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
archive_path="$repository_root/dist/puretokens-skill-install.zip"
legacy_bootstrap_archive_path="$repository_root/dist/puretokens-skill-install-payload.zip"
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

# Published 0.13.x updaters are hard-coded to request this exact archive and
# validate that it has a legacy runtime marker before copying the current
# Skills. It is not linked from current installation docs and is never used at
# runtime. Keeping this bridge lets those users reach the source-only release.
legacy_source_root="$staging_root/legacy-bootstrap/puretokens-skill-main"
mkdir -p "$legacy_source_root"
cp "$repository_root/README.md" "$legacy_source_root/README.md"
cp "$repository_root/package.json" "$legacy_source_root/package.json"
cp -R "$repository_root/runtime" "$legacy_source_root/runtime"
cp -R "$repository_root/skills" "$legacy_source_root/skills"
cp "$repository_root/scripts/legacy-bootstrap/runtime.json" "$legacy_source_root/runtime/runtime.json"
cp "$repository_root/scripts/legacy-bootstrap/puretokens-direct-api.mjs" "$legacy_source_root/runtime/puretokens-direct-api.mjs"

grep -Fq '"name": "puretokens-direct-api-runtime"' "$legacy_source_root/runtime/runtime.json" || {
  printf '%s\n' "legacy bootstrap archive is missing its compatibility marker" >&2
  exit 1
}
grep -Fq 'legacyBootstrapOnly' "$legacy_source_root/runtime/runtime.json" || {
  printf '%s\n' "legacy bootstrap archive runtime must be marker-only" >&2
  exit 1
}

legacy_temporary_archive="$staging_root/puretokens-skill-install-payload.zip"
(cd "$staging_root/legacy-bootstrap" && zip -qr "$legacy_temporary_archive" puretokens-skill-main)
mv "$legacy_temporary_archive" "$legacy_bootstrap_archive_path"
