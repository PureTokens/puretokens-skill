#!/bin/sh

set -eu

current_skills="puretokens-balance puretokens-connection puretokens-models puretokens-image puretokens-video puretokens-update"
retired_skills="puretokens_media puretokens_balance puretokens_connection puretokens_models puretokens_image puretokens_video puretokens_update puretokens_get_balance puretokens_get_model_price puretokens_workbuddy_router"

usage() {
  printf '%s\n' "Usage: puretokens-skill-install.sh <check|sync> (--host <claude-code|codex|workbuddy|gemini-cli|grok-build|opencode|trae> | --target <absolute-skill-directory>) [--source <absolute-official-source-directory>]"
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

target_for_host() {
  host=$1
  [ -n "${HOME:-}" ] || fail "cannot resolve a host Skill directory because HOME is unavailable"
  case "$host" in
    claude-code) printf '%s\n' "$HOME/.claude/skills" ;;
    codex) printf '%s\n' "$HOME/.agents/skills" ;;
    workbuddy) printf '%s\n' "$HOME/.workbuddy/skills" ;;
    gemini-cli) printf '%s\n' "$HOME/.gemini/skills" ;;
    grok-build) printf '%s\n' "$HOME/.grok/skills" ;;
    opencode) printf '%s\n' "$HOME/.config/opencode/skills" ;;
    trae) printf '%s\n' "$HOME/.trae/skills" ;;
    *) fail "unsupported host: $host" ;;
  esac
}

migrate_legacy_codex_plugin() {
  target_root=$1
  [ -n "${HOME:-}" ] || return 0
  home_root=$(cd "$HOME" && pwd -P) || return 0
  [ "$target_root" = "$home_root/.agents/skills" ] || return 0
  if ! command -v codex >/dev/null 2>&1; then
    fail "cannot verify removal of legacy Codex plugin puretokens-media because the Codex CLI is unavailable; remove it in Codex Plugins, then run this installer again"
  fi
  plugin_list=$(codex plugin list --json 2>/dev/null) || {
    fail "cannot verify removal of legacy Codex plugin puretokens-media because Codex Plugins could not be inspected; remove it in Codex Plugins, then run this installer again"
  }
  if ! printf '%s' "$plugin_list" | grep -Eq '"name"[[:space:]]*:[[:space:]]*"puretokens-media"'; then
    return 0
  fi
  plugin_selectors=$(printf '%s' "$plugin_list" | tr '{},' '\n' | sed -n 's/.*"pluginId"[[:space:]]*:[[:space:]]*"\(puretokens-media@[^"]*\)".*/\1/p' | sort -u)
  [ -n "$plugin_selectors" ] || plugin_selectors="puretokens-media"
  for plugin_selector in $plugin_selectors; do
    codex plugin remove "$plugin_selector" --json >/dev/null 2>&1 ||
      fail "could not remove legacy Codex plugin $plugin_selector; remove it in Codex Plugins, then run this installer again"
  done
  plugin_list=$(codex plugin list --json 2>/dev/null) ||
    fail "could not verify removal of legacy Codex plugin puretokens-media; reopen Codex Plugins, remove it if present, then run this installer again"
  if printf '%s' "$plugin_list" | grep -Eq '"name"[[:space:]]*:[[:space:]]*"puretokens-media"'; then
    fail "legacy Codex plugin puretokens-media is still installed; remove it in Codex Plugins, then run this installer again"
  fi
  printf '%s\n' "Removed and verified legacy Codex plugin puretokens-media. Fully restart Codex before testing the new Skills."
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

  migrate_legacy_codex_plugin "$target_root"

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
  printf '%s\n' "Pure Tokens Skills $release_version synchronized at $target_root"
}

command_name=${1:-}
[ -n "$command_name" ] || { usage; exit 2; }
shift
target=
host=
source=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --target)
      [ "$#" -ge 2 ] || fail "--target requires an absolute directory"
      target=$2
      shift 2
      ;;
    --host)
      [ "$#" -ge 2 ] || fail "--host requires a supported host ID"
      host=$2
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
[ -z "$target" ] || [ -z "$host" ] || fail "use either --host or --target, not both"
[ -n "$target" ] || [ -n "$host" ] || fail "--host or --target is required"
[ -z "$host" ] || target=$(target_for_host "$host")
[ "${target#/}" != "$target" ] || fail "--target must be an absolute Skill directory"
[ -z "$source" ] || [ "${source#/}" != "$source" ] || fail "--source must be an absolute official source directory"
case "$command_name" in check|sync) ;; *) usage; exit 2 ;; esac

require_command mktemp
if [ -n "$source" ]; then
  [ -d "$source" ] || fail "--source does not exist: $source"
  source_root=$(cd "$source" && pwd -P) || fail "could not resolve --source"
  validate_source "$source_root"
elif bundled_source_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." 2>/dev/null && pwd -P) && [ -f "$bundled_source_root/README.md" ] && [ -f "$bundled_source_root/runtime/runtime.json" ] && [ -d "$bundled_source_root/skills" ]; then
  source_root=$bundled_source_root
  validate_source "$source_root"
else
  fail "run this source-only sync script from a fresh official Pure Tokens Skills main checkout or pass --source <absolute-checkout-directory>"
fi
if [ "$command_name" = "check" ]; then
  printf '%s\n' "Official Pure Tokens Skill source passed native static validation."
else
  sync_target "$source_root" "$target"
fi
