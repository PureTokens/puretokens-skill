#!/bin/sh

set -eu

payload_archive_primary_url="https://api.github.com/repos/PureTokens/puretokens-skill/contents/dist/puretokens-skill-install-payload.zip?ref=main"
payload_archive_fallback_url="https://raw.githubusercontent.com/PureTokens/puretokens-skill/main/dist/puretokens-skill-install-payload.zip"
payload_download_attempt_deadline_seconds=20
current_skills="puretokens-balance puretokens-connection puretokens-models puretokens-image puretokens-video puretokens-update"
retired_skills="puretokens_media puretokens_balance puretokens_connection puretokens_models puretokens_image puretokens_video puretokens_update puretokens_get_balance puretokens_get_model_price puretokens_workbuddy_router"

usage() {
  printf '%s\n' "Usage: puretokens-skill-install.sh <check|sync> --target <absolute-skill-directory> [--source <absolute-official-source-directory>]"
}

fail() {
  printf '%s\n' "Pure Tokens Skill installer: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "required system command is unavailable: $1"
}

managed_skill() {
  directory=$1
  name=$2
  [ -f "$directory/SKILL.md" ] && [ -f "$directory/skill.json" ] &&
    grep -Fq "\"name\": \"$name\"" "$directory/skill.json"
}

managed_runtime() {
  directory=$1
  [ -f "$directory/runtime.json" ] && [ -f "$directory/puretokens-direct-api.mjs" ] &&
    grep -Fq '"name": "puretokens-direct-api-runtime"' "$directory/runtime.json"
}

managed_release_version() {
  version=$(sed -n 's/^[[:space:]]*"version"[[:space:]]*:[[:space:]]*"\([0-9][0-9.]*\)".*$/\1/p' "$1/runtime/runtime.json" | sed -n '1p')
  case "$version" in
    *.*.*) printf '%s\n' "$version" ;;
    *) return 1 ;;
  esac
}

migrate_legacy_codex_plugin() {
  target_root=$1
  [ -n "${HOME:-}" ] || return 0
  [ "$target_root" = "$HOME/.agents/skills" ] || return 0
  if ! command -v codex >/dev/null 2>&1; then
    printf '%s\n' "Codex legacy-plugin migration was skipped because the Codex CLI is unavailable. If Puretokens Media is installed, remove it in Codex Plugins before opening a new conversation."
    return 0
  fi
  plugin_list=$(codex plugin list --json 2>/dev/null) || {
    printf '%s\n' "Codex legacy-plugin migration could not inspect installed plugins. If Puretokens Media is installed, remove it in Codex Plugins before opening a new conversation."
    return 0
  }
  if ! printf '%s' "$plugin_list" | grep -Eq '"name"[[:space:]]*:[[:space:]]*"puretokens-media"'; then
    return 0
  fi
  if codex plugin remove puretokens-media --json >/dev/null 2>&1; then
    printf '%s\n' "Removed legacy Codex plugin puretokens-media"
  else
    printf '%s\n' "Codex could not remove the legacy Puretokens Media plugin. Remove it in Codex Plugins, or ask the workspace administrator if it is managed, before opening a new conversation."
  fi
}

validate_source() {
  source_root=$1
  [ -f "$source_root/README.md" ] || fail "official source is missing README.md"
  [ -f "$source_root/runtime/runtime.json" ] || fail "official source is missing the managed runtime manifest"
  managed_runtime "$source_root/runtime" || fail "official source has an invalid managed runtime"
  managed_release_version "$source_root" >/dev/null || fail "official source has an invalid managed runtime version"
  [ -f "$source_root/runtime/puretokens-skill-install.sh" ] || fail "official source is missing the native installer"
  [ -f "$source_root/runtime/puretokens-skill-install.ps1" ] || fail "official source is missing the Windows native installer"
  for name in $current_skills; do
    managed_skill "$source_root/skills/$name" "$name" || fail "official source has an invalid Skill: $name"
  done
}

download_source() {
  workspace=$1
  archive="$workspace/puretokens-skill-install-payload.zip"
  unpacked="$workspace/unpacked"
  mkdir -p "$unpacked"
  printf '%s\n' "Downloading the compact official Pure Tokens Skill install payload through the official GitHub API." >&2
  if ! curl --fail --location --proto '=https' --proto-redir '=https' --tlsv1.2 --connect-timeout 7 --max-time "$payload_download_attempt_deadline_seconds" \
    --header 'Accept: application/vnd.github.raw+json' --header 'X-GitHub-Api-Version: 2022-11-28' \
    "$payload_archive_primary_url" --output "$archive"; then
    printf '%s\n' "The official GitHub API download path was unavailable; trying GitHub raw content once." >&2
    curl --fail --location --proto '=https' --proto-redir '=https' --tlsv1.2 --connect-timeout 7 --max-time "$payload_download_attempt_deadline_seconds" \
      "$payload_archive_fallback_url" --output "$archive" ||
      fail "could not download the compact official Pure Tokens Skill install payload from either official GitHub path within two 20-second attempts"
  fi
  unzip -q "$archive" -d "$unpacked" || fail "could not unpack the official Pure Tokens Skill source"
  source_root="$unpacked/puretokens-skill-main"
  [ -d "$source_root" ] || fail "official source archive has an unexpected layout"
  validate_source "$source_root"
  printf '%s\n' "$source_root"
}

restore_target() {
  target_root=$1
  stage_root=$2
  replaced_names=$3
  created_names=$4
  for name in $replaced_names; do
    destination="$target_root/$name"
    backup="$stage_root/backup/$name"
    [ -e "$backup" ] || continue
    [ ! -e "$destination" ] || rm -rf -- "$destination"
    mv "$backup" "$destination" || true
  done
  for name in $created_names; do
    [ ! -e "$target_root/$name" ] || rm -rf -- "$target_root/$name"
  done
}

sync_target() {
  source_root=$1
  target_root=$2
  release_version=$(managed_release_version "$source_root") || fail "official source has an invalid managed runtime version"
  [ -d "$target_root" ] || mkdir -p "$target_root"
  target_root=$(cd "$target_root" && pwd -P)

  for name in $retired_skills; do
    for destination in "$target_root/$name" "$target_root/.$name.retired-"*; do
      [ ! -e "$destination" ] || managed_skill "$destination" "$name" ||
        fail "unmanaged retired Skill conflicts: $destination"
    done
  done
  [ ! -e "$target_root/.puretokens-runtime" ] || managed_runtime "$target_root/.puretokens-runtime" ||
    fail "unmanaged Pure Tokens runtime conflicts: $target_root/.puretokens-runtime"
  for name in $current_skills; do
    destination="$target_root/$name"
    [ ! -e "$destination" ] || managed_skill "$destination" "$name" ||
      fail "unmanaged Skill conflicts: $destination"
  done

  stage_root=$(mktemp -d "$target_root/.puretokens-skill-stage.XXXXXX") || fail "could not create a private update staging directory"
  mkdir -p "$stage_root/backup"
  cp -R "$source_root/runtime" "$stage_root/.puretokens-runtime"
  for name in $current_skills; do cp -R "$source_root/skills/$name" "$stage_root/$name"; done

  replaced_names=
  created_names=
  if [ -e "$target_root/.puretokens-runtime" ]; then
    mv "$target_root/.puretokens-runtime" "$stage_root/backup/.puretokens-runtime"
    replaced_names="$replaced_names .puretokens-runtime"
  else
    created_names="$created_names .puretokens-runtime"
  fi
  if ! mv "$stage_root/.puretokens-runtime" "$target_root/.puretokens-runtime"; then
    restore_target "$target_root" "$stage_root" "$replaced_names" "$created_names"
    rm -rf -- "$stage_root"
    fail "could not install the managed runtime"
  fi
  for name in $current_skills; do
    destination="$target_root/$name"
    if [ -e "$destination" ]; then
      mv "$destination" "$stage_root/backup/$name"
      replaced_names="$replaced_names $name"
    else
      created_names="$created_names $name"
    fi
    if ! mv "$stage_root/$name" "$destination"; then
      restore_target "$target_root" "$stage_root" "$replaced_names" "$created_names"
      rm -rf -- "$stage_root"
      fail "could not install Skill: $name"
    fi
  done
  for name in $retired_skills; do
    for destination in "$target_root/$name" "$target_root/.$name.retired-"*; do
      [ ! -e "$destination" ] && continue
      rm -rf -- "$destination" || fail "could not remove retired Skill: $name"
      printf '%s\n' "Removed retired managed $name from $destination"
    done
  done
  rm -rf -- "$stage_root"
  migrate_legacy_codex_plugin "$target_root"
  printf '%s\n' "Pure Tokens Skills $release_version synchronized at $target_root"
}

command_name=${1:-}
[ -n "$command_name" ] || { usage; exit 2; }
shift
target=
source=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --target)
      [ "$#" -ge 2 ] || fail "--target requires an absolute directory"
      target=$2
      shift 2
      ;;
    --source)
      [ "$#" -ge 2 ] || fail "--source requires an absolute official source directory"
      source=$2
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done
[ -n "$target" ] && [ "${target#/}" != "$target" ] || fail "--target must be an absolute Skill directory"
[ -z "$source" ] || [ "${source#/}" != "$source" ] || fail "--source must be an absolute official source directory"
case "$command_name" in check|sync) ;; *) usage; exit 2 ;; esac

require_command mktemp
if [ -n "$source" ]; then
  [ -d "$source" ] || fail "--source does not exist: $source"
  source_root=$(cd "$source" && pwd -P) || fail "could not resolve --source"
  validate_source "$source_root"
else
  require_command curl
  require_command unzip
  workspace=$(mktemp -d "${TMPDIR:-/tmp}/puretokens-skill.XXXXXX") || fail "could not create a private temporary directory"
  cleanup_workspace() { rm -rf -- "$workspace"; }
  trap cleanup_workspace EXIT HUP INT TERM
  source_root=$(download_source "$workspace")
fi
if [ "$command_name" = "check" ]; then
  printf '%s\n' "Official Pure Tokens Skill source passed native static validation."
else
  sync_target "$source_root" "$target"
fi
