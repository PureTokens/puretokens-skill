#!/bin/sh

set -eu

current_skills="puretokens-balance puretokens-connection puretokens-models puretokens-image puretokens-video puretokens-update"
retired_skills="puretokens_media puretokens_balance puretokens_connection puretokens_models puretokens_image puretokens_video puretokens_update puretokens_get_balance puretokens_get_model_price puretokens_workbuddy_router"

usage() {
  printf '%s\n' "Usage: puretokens-skill-install.sh <check|init|sync> (--host <claude-code|codex|workbuddy|gemini-cli|grok-build|opencode|trae> | --target <absolute-skill-directory>) [--source <absolute-official-source-directory>]"
}

fail() {
  printf '%s\n' "Pure Tokens Skill installer: $*" >&2
  exit 1
}

managed_skill() {
  directory=$1
  name=$2
  [ -f "$directory/SKILL.md" ] && [ -f "$directory/skill.json" ] &&
    grep -Fq "\"name\": \"$name\"" "$directory/skill.json"
}

legacy_node_runtime() {
  directory=$1
  [ -f "$directory/runtime.json" ] && [ -f "$directory/puretokens-direct-api.mjs" ] &&
    grep -Eq '"name"[[:space:]]*:[[:space:]]*"puretokens-direct-api-runtime"' "$directory/runtime.json"
}

managed_executor() {
  directory=$1
  [ -f "$directory/runtime.json" ] && [ -f "$directory/puretokens-api" ] &&
    grep -Eq '"name"[[:space:]]*:[[:space:]]*"puretokens-api-executor"' "$directory/runtime.json"
}

executor_platform() {
  case "$(uname -s 2>/dev/null)-$(uname -m 2>/dev/null)" in
    Darwin-arm64) printf '%s\n' "darwin-arm64" ;;
    Darwin-x86_64) printf '%s\n' "darwin-amd64" ;;
    Linux-x86_64) printf '%s\n' "linux-amd64" ;;
    Linux-aarch64|Linux-arm64) printf '%s\n' "linux-arm64" ;;
    *) fail "no Pure Tokens executor is available for this operating system and CPU" ;;
  esac
}

sha256_file() {
  file=$1
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
  else
    fail "cannot verify the Pure Tokens executor because SHA-256 is unavailable"
  fi
}

executor_artifact() {
  source_root=$1
  platform=$(executor_platform)
  manifest="$source_root/runtime/executor/manifest.json"
  expected_path=$(sed -n "/\"$platform\"[[:space:]]*:/,/}/ s/.*\"path\"[[:space:]]*:[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" "$manifest" | sed -n '1p')
  expected_sha=$(sed -n "/\"$platform\"[[:space:]]*:/,/}/ s/.*\"sha256\"[[:space:]]*:[[:space:]]*\"\\([0-9a-f][0-9a-f]*\\)\".*/\\1/p" "$manifest" | sed -n '1p')
  [ -n "$expected_path" ] && [ -n "$expected_sha" ] || fail "official source is missing the executor artifact for $platform"
  artifact="$source_root/runtime/executor/$expected_path"
  [ -f "$artifact" ] || fail "official source is missing the executor binary for $platform"
  [ "$(sha256_file "$artifact")" = "$expected_sha" ] || fail "official source executor checksum mismatch for $platform"
  printf '%s\n' "$artifact"
}

source_release_version() {
  version=$(sed -n 's/^[[:space:]]*"version"[[:space:]]*:[[:space:]]*"\([0-9][0-9.]*\)".*$/\1/p' "$1/package.json" | sed -n '1p')
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
  plugin_list=$(codex plugin list --json 2>/dev/null) || fail "cannot inspect Codex Plugins; remove legacy puretokens-media in Codex Plugins, then run this installer again"
  if ! printf '%s' "$plugin_list" | grep -Eq '"name"[[:space:]]*:[[:space:]]*"puretokens-media"'; then
    return 0
  fi
  plugin_selectors=$(printf '%s' "$plugin_list" | tr '{},' '\n' | sed -n 's/.*"pluginId"[[:space:]]*:[[:space:]]*"\(puretokens-media@[^"]*\)".*/\1/p' | sort -u)
  [ -n "$plugin_selectors" ] || plugin_selectors="puretokens-media"
  for plugin_selector in $plugin_selectors; do
    codex plugin remove "$plugin_selector" --json >/dev/null 2>&1 || fail "could not remove legacy Codex plugin $plugin_selector; remove it in Codex Plugins, then run this installer again"
  done
  plugin_list=$(codex plugin list --json 2>/dev/null) || fail "could not verify removal of legacy Codex plugin puretokens-media"
  if printf '%s' "$plugin_list" | grep -Eq '"name"[[:space:]]*:[[:space:]]*"puretokens-media"'; then
    fail "legacy Codex plugin puretokens-media is still installed; remove it in Codex Plugins, then run this installer again"
  fi
  printf '%s\n' "Removed and verified legacy Codex plugin puretokens-media. Fully restart Codex before testing the new Skills."
}

validate_source() {
  source_root=$1
  [ -f "$source_root/README.md" ] || fail "official source is missing README.md"
  [ -f "$source_root/package.json" ] || fail "official source is missing package.json"
  source_release_version "$source_root" >/dev/null || fail "official source has an invalid Skill version"
  [ -f "$source_root/runtime/puretokens-skill-install.sh" ] || fail "official source is missing the macOS/Linux installer"
  [ -f "$source_root/runtime/puretokens-skill-install.ps1" ] || fail "official source is missing the Windows installer"
  [ -f "$source_root/runtime/executor/manifest.json" ] || fail "official source is missing the executor manifest"
  executor_artifact "$source_root" >/dev/null
  for name in $current_skills; do
    managed_skill "$source_root/skills/$name" "$name" || fail "official source has an invalid Skill: $name"
  done
}

validate_target() {
  target_root=$1
  [ -d "$target_root" ] || fail "target Skill directory does not exist: $target_root"
  for name in $current_skills; do
    managed_skill "$target_root/$name" "$name" || fail "target is missing the managed Skill: $name"
  done
  managed_executor "$target_root/.puretokens-executor" || fail "target is missing the managed native executor"
}

usage_guide() {
  target_root=$1
  guide="$target_root/puretokens-update/references/usage-guide.md"
  if [ -f "$guide" ]; then
    cat "$guide"
    return 0
  fi
  printf '%s\n' "Pure Tokens Skill usage guide is unavailable; update the Skills from the official repository."
}

init_json_string() {
  field=$1
  document=$2
  printf '%s' "$document" | sed -n "s/.*\"$field\":\"\([^\"]*\)\".*/\1/p" | sed -n '1p'
}

init_json_number() {
  field=$1
  document=$2
  printf '%s' "$document" | sed -n "s/.*\"$field\":\([0-9][0-9]*\).*/\1/p" | sed -n '1p'
}

init_target() {
  target_root=$1
  host_id=$2
  validate_target "$target_root"
  if [ -n "$host_id" ]; then
    executor="$target_root/.puretokens-executor/puretokens-api"
    init_output=
    init_status=0
    init_output=$("$executor" init --host "$host_id" 2>/dev/null) || init_status=$?
    if printf '%s' "$init_output" | grep -Fq '"configuration_status":"verified"'; then
      printf '%s\n' "Pure Tokens Skill init: fixed API identity verified for the current host."
    elif [ "$init_status" -ne 0 ] && [ -z "$init_output" ]; then
      printf '%s\n' "Pure Tokens Skill init: configuration could not be checked in this host session."
    else
      init_state=$(init_json_string "configuration_status" "$init_output")
      init_message=$(init_json_string "message" "$init_output")
      init_next=$(init_json_string "next_action" "$init_output")
      init_http_status=$(init_json_number "http_status" "$init_output")
      [ -n "$init_state" ] || init_state="unverified"
      [ -n "$init_message" ] || init_message="The fixed API identity is not verified in this host session."
      printf '%s\n' "Pure Tokens Skill init: $init_message [$init_state]"
      if [ -n "$init_http_status" ]; then
        printf '%s\n' "HTTP status: $init_http_status"
      fi
      if [ -n "$init_next" ]; then
        printf '%s\n' "Next: $init_next"
      fi
    fi
  else
    printf '%s\n' "Pure Tokens Skill init: host ID was not supplied, so the connection check was deferred."
  fi
  printf '%s\n' ""
  usage_guide "$target_root"
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
  release_version=$(source_release_version "$source_root") || fail "official source has an invalid Skill version"
  [ -d "$target_root" ] || mkdir -p "$target_root"
  target_root=$(cd "$target_root" && pwd -P)

  for name in $retired_skills; do
    for destination in "$target_root/$name" "$target_root/.$name.retired-"*; do
      [ ! -e "$destination" ] || managed_skill "$destination" "$name" || fail "unmanaged retired Skill conflicts: $destination"
    done
  done
  for name in $current_skills; do
    destination="$target_root/$name"
    [ ! -e "$destination" ] || managed_skill "$destination" "$name" || fail "unmanaged Skill conflicts: $destination"
  done
  executor_destination="$target_root/.puretokens-executor"
  [ ! -e "$executor_destination" ] || managed_executor "$executor_destination" || fail "unmanaged Pure Tokens executor conflicts: $executor_destination"

  stage_root=$(mktemp -d "$target_root/.puretokens-skill-stage.XXXXXX") || fail "could not create a private update staging directory"
  cleanup_stage() { [ ! -d "$stage_root" ] || rm -rf -- "$stage_root"; }
  trap cleanup_stage EXIT HUP INT TERM
  mkdir -p "$stage_root/backup"
  for name in $current_skills; do cp -R "$source_root/skills/$name" "$stage_root/$name"; done
  executor_source=$(executor_artifact "$source_root")
  mkdir -p "$stage_root/executor"
  cp "$executor_source" "$stage_root/executor/puretokens-api"
  chmod 700 "$stage_root/executor/puretokens-api"
  printf '{\n  "schemaVersion": 1,\n  "name": "puretokens-api-executor",\n  "version": "%s",\n  "platform": "%s"\n}\n' "$release_version" "$(executor_platform)" > "$stage_root/executor/runtime.json"

  migrate_legacy_codex_plugin "$target_root"

  replaced_names=
  created_names=
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
      fail "could not install Skill: $name"
    fi
  done
  executor_replaced=0
  if [ -e "$executor_destination" ]; then
    mv "$executor_destination" "$stage_root/backup/executor" || { restore_target "$target_root" "$stage_root" "$replaced_names" "$created_names"; fail "could not stage the current Pure Tokens executor"; }
    executor_replaced=1
  fi
  if ! mv "$stage_root/executor" "$executor_destination"; then
    [ "$executor_replaced" -eq 0 ] || mv "$stage_root/backup/executor" "$executor_destination" || true
    restore_target "$target_root" "$stage_root" "$replaced_names" "$created_names"
    fail "could not install the Pure Tokens executor"
  fi
  for name in $retired_skills; do
    for destination in "$target_root/$name" "$target_root/.$name.retired-"*; do
      [ ! -e "$destination" ] && continue
      rm -rf -- "$destination" || fail "could not remove retired Skill: $name"
      printf '%s\n' "Removed retired managed $name from $destination"
    done
  done
  legacy_runtime="$target_root/.puretokens-runtime"
  if [ -e "$legacy_runtime" ] && legacy_node_runtime "$legacy_runtime"; then
    rm -rf -- "$legacy_runtime" || fail "could not remove retired Node runtime"
    printf '%s\n' "Removed retired managed Node runtime from $legacy_runtime"
  fi
  printf '%s\n' "Pure Tokens Skills $release_version synchronized with the native API executor at $target_root"
  init_target "$target_root" "$host"
}

command_name=${1:-}
[ -n "$command_name" ] || { usage; exit 2; }
shift
target=
host=
source=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --target) [ "$#" -ge 2 ] || fail "--target requires an absolute directory"; target=$2; shift 2 ;;
    --host) [ "$#" -ge 2 ] || fail "--host requires a supported host ID"; host=$2; shift 2 ;;
    --source) [ "$#" -ge 2 ] || fail "--source requires an absolute official source directory"; source=$2; shift 2 ;;
    *) usage; exit 2 ;;
  esac
done
[ -z "$target" ] || [ -z "$host" ] || fail "use either --host or --target, not both"
[ -n "$target" ] || [ -n "$host" ] || fail "--host or --target is required"
[ -z "$host" ] || target=$(target_for_host "$host")
[ "${target#/}" != "$target" ] || fail "--target must be an absolute Skill directory"
[ -z "$source" ] || [ "${source#/}" != "$source" ] || fail "--source must be an absolute official source directory"
case "$command_name" in check|init|sync) ;; *) usage; exit 2 ;; esac

if [ "$command_name" = "init" ]; then
  [ -n "$target" ] || fail "--target or --host is required for init"
  [ "${target#/}" != "$target" ] || fail "--target must be an absolute Skill directory"
  target_root=$(cd "$target" 2>/dev/null && pwd -P) || fail "could not resolve --target"
  init_target "$target_root" "$host"
  exit 0
fi

if [ -n "$source" ]; then
  [ -d "$source" ] || fail "--source does not exist: $source"
  source_root=$(cd "$source" && pwd -P) || fail "could not resolve --source"
  validate_source "$source_root"
elif bundled_source_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." 2>/dev/null && pwd -P) && [ -f "$bundled_source_root/README.md" ] && [ -f "$bundled_source_root/package.json" ] && [ -d "$bundled_source_root/skills" ]; then
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
