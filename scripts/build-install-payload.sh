#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
output="$repository_root/dist/puretokens-skill-install.zip"
legacy_output="$repository_root/dist/puretokens-skill-install-payload.zip"
stage=$(mktemp -d "${TMPDIR:-/tmp}/puretokens-skill-payload.XXXXXX")
cleanup() { rm -rf -- "$stage"; }
trap cleanup EXIT HUP INT TERM

mkdir -p "$stage/puretokens-skill-main"
cp -R "$repository_root/runtime" "$stage/puretokens-skill-main/runtime"
cp -R "$repository_root/skills" "$stage/puretokens-skill-main/skills"
cp "$repository_root/README.md" "$stage/puretokens-skill-main/README.md"
rm -f -- "$output" "$legacy_output"
(cd "$stage" && zip -X -q -r "$output" puretokens-skill-main)
cp "$output" "$legacy_output"
