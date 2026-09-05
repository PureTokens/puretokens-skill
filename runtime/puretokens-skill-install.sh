#!/bin/sh
# puretokens-locate-v1

set -eu

current_skills="puretokens-balance puretokens-connection puretokens-models puretokens-image puretokens-video puretokens-update"
retired_skills="puretokens_media puretokens_balance puretokens_connection puretokens_models puretokens_image puretokens_video puretokens_update puretokens_get_balance puretokens_get_model_price puretokens_workbuddy_router"

usage() {
  printf '%s\n' "Usage: puretokens-skill-install.sh <check|init|sync|locate> (--host <claude-code|codex|workbuddy|gemini-cli|grok-build|opencode|trae> | --target <absolute-skill-directory>) [--source <absolute-official-source-directory>]"
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
    claude-code) printf '%s\n' "${CLAUDE_CONFIG_DIR:-$HOME/.claude}/skills" ;;
    codex) printf '%s\n' "$HOME/.agents/skills" ;;
    workbuddy) printf '%s\n' "${WORKBUDDY_CONFIG_DIR:-${CODEBUDDY_CONFIG_DIR:-$HOME/.workbuddy}}/skills" ;;
    gemini-cli)
      # Gemini resolves each .agents Skill before the same .gemini Skill.
      for skill in $current_skills; do
        if [ -e "$HOME/.agents/skills/$skill" ] || [ -L "$HOME/.agents/skills/$skill" ]; then
          printf '%s\n' "$HOME/.agents/skills"
          return
        fi
      done
      printf '%s\n' "$HOME/.gemini/skills"
      ;;
    grok-build) printf '%s\n' "$HOME/.grok/skills" ;;
    opencode) printf '%s\n' "$HOME/.config/opencode/skills" ;;
    trae) printf '%s\n' "$HOME/.trae/skills" ;;
    *) fail "unsupported host: $host" ;;
  esac
}

check_gemini_duplicates() {
  [ "$host" = gemini-cli ] && [ -n "${HOME:-}" ] || return 0
  shared_root="$HOME/.agents/skills"
  if [ -d "$shared_root" ]; then shared_root=$(cd "$shared_root" && pwd -P); fi
  for skill in $current_skills; do
    if [ "$target_root" != "$shared_root" ] && { [ -e "$shared_root/$skill" ] || [ -L "$shared_root/$skill" ]; }; then
      fail "Gemini would load $skill from its higher-priority .agents/skills directory. Run sync --host gemini-cli without a custom target; existing directories were left untouched"
    fi
    if [ "$target_root" = "$shared_root" ] && managed_skill "$HOME/.gemini/skills/$skill" "$skill"; then
      printf '%s\n' "Managed duplicate detected: Gemini will use the updated shared $skill; its lower-priority .gemini copy was left untouched."
    fi
  done
}

migrate_legacy_codex_plugin() {
  target_root=$1
  [ -n "${HOME:-}" ] || return 0
  home_root=$(cd "$HOME" && pwd -P) || return 0
  [ "$target_root" = "$home_root/.agents/skills" ] || return 0
  if ! command -v codex >/dev/null 2>&1; then
    printf '%s\n' "Legacy Codex plugin check unavailable. If an old Puretokens Media plugin appears, remove that plugin in Codex Plugins and restart Codex."
    return 0
  fi
  plugin_list=$(codex plugin list --json 2>/dev/null) || { printf '%s\n' "Legacy Codex plugin inspection unavailable; check Codex Plugins only if old Media instructions remain."; return 0; }
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
  [ -f "$source_root/runtime/puretokens-skill-fetch.sh" ] || fail "official source is missing the download wrapper"
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

restore_transaction() (
  recovery_root=$1
  recovery_stage=$2
  [ -f "$recovery_stage/plan" ] || return 0
  while read -r action entry; do
    case "$entry" in puretokens-balance|puretokens-connection|puretokens-models|puretokens-image|puretokens-video|puretokens-update|.puretokens-executor) ;; *) return 1 ;; esac
    case "$entry" in */*|*..*) return 1 ;; esac
    if [ "$action" = replace ] && [ -e "$recovery_stage/backup/$entry" ]; then
      [ ! -e "$recovery_root/$entry" ] || rm -rf -- "$recovery_root/$entry" || return 1
      mv "$recovery_stage/backup/$entry" "$recovery_root/$entry" || return 1
    elif [ "$action" = create ] && [ ! -e "$recovery_stage/$entry" ]; then
      [ ! -e "$recovery_root/$entry" ] || rm -rf -- "$recovery_root/$entry" || return 1
    fi
  done < "$recovery_stage/plan"
)

finish_transaction() {
  code=$?
  trap - EXIT HUP INT TERM
  if [ -n "${stage_root:-}" ] && [ -d "$stage_root" ]; then
    if [ -f "$stage_root/committed" ] || restore_transaction "$target_root" "$stage_root"; then
      rm -rf -- "$stage_root"
    else
      printf '%s\n' "Update recovery is incomplete; retained the managed recovery stage. Run sync again after resolving file access." >&2
      code=1
    fi
  fi
  release_update_lock
  exit "$code"
}

read_lock_owner() {
  lock_record=$(cat "$1" 2>/dev/null) || return 1
  case "$lock_record" in
    *" "*) lock_owner=${lock_record%% *}; lock_token=${lock_record#* } ;;
    *) lock_owner=$lock_record; lock_token="legacy-$lock_owner" ;;
  esac
  case "$lock_owner" in ''|*[!0-9]*) return 1 ;; esac
  case "$lock_token" in ''|*[!A-Za-z0-9._-]*) return 1 ;; esac
}

owns_update_lock() (
  lock_slot="$lock_root/pid"
  lock_seen=" "
  while read_lock_owner "$lock_slot"; do
    case "$lock_seen" in *" $lock_token "*) return 1 ;; esac
    lock_seen="$lock_seen$lock_token "
    if [ "$lock_owner" = "$$" ] && [ "$lock_token" = "${lock_candidate##*/}" ]; then return 0; fi
    lock_slot="$lock_root/next.$lock_token"
  done
  return 1
)

release_update_lock() {
  if owns_update_lock; then rm -rf -- "$lock_root"; fi
}

claim_update_lock() (
  ln "$lock_candidate" "$lock_slot" 2>/dev/null && return 0
  [ ! -e "$lock_slot" ] && [ ! -L "$lock_slot" ] || return 1
  # Filesystems without hard links still support exclusive creation. Never
  # replace a competing record; an interrupted partial write stops recovery.
  set -C
  cat "$lock_candidate" > "$lock_slot"
)

acquire_update_lock() {
  lock_root="$target_root/.puretokens-install-lock"
  mkdir "$lock_root" 2>/dev/null || [ -d "$lock_root" ] || fail "cannot create the update lock"
  for lock_entry in "$lock_root"/* "$lock_root"/.[!.]* "$lock_root"/..?*; do
    [ -e "$lock_entry" ] || [ -L "$lock_entry" ] || continue
    case "${lock_entry##*/}" in
      pid|owner.*|next.*) [ -f "$lock_entry" ] && [ ! -L "$lock_entry" ] || fail "unrecognized update lock contents; left untouched" ;;
      *) fail "unrecognized update lock contents; left untouched" ;;
    esac
  done
  lock_candidate=$(mktemp "$lock_root/owner.$$.XXXXXX") || fail "cannot prepare the update lock"
  printf '%s %s\n' "$$" "${lock_candidate##*/}" > "$lock_candidate"
  lock_slot="$lock_root/pid"
  lock_seen=" "
  # Publish complete, immutable owner records with an atomic hard link. A dead
  # owner has exactly one successor, so simultaneous recovery cannot replace a
  # live lock. An interrupted takeover is simply another dead owner to follow.
  while ! claim_update_lock 2>/dev/null; do
    if ! read_lock_owner "$lock_slot"; then
      rm -f -- "$lock_candidate"
      fail "another installation is starting or the update lock needs inspection"
    fi
    case "$lock_seen" in *" $lock_token "*) rm -f -- "$lock_candidate"; fail "invalid update lock history; left untouched" ;; esac
    lock_seen="$lock_seen$lock_token "
    if kill -0 "$lock_owner" 2>/dev/null; then
      rm -f -- "$lock_candidate"
      fail "another installation is in progress; wait for it to finish"
    fi
    lock_slot="$lock_root/next.$lock_token"
  done
  # A previous owner may have finished and removed its directory while this
  # process was reading it. Only a claim reachable from the current root owns it.
  if ! owns_update_lock; then
    rm -f -- "$lock_candidate"
    fail "another installation acquired the update lock"
  fi
  stage_root=
  trap finish_transaction EXIT
  trap 'exit 130' INT
  trap 'exit 129' HUP
  trap 'exit 143' TERM
  for previous_stage in "$target_root"/.puretokens-skill-stage.*; do
    [ -d "$previous_stage" ] || continue
    [ -f "$previous_stage/transaction-v1" ] || fail "an unrecognized staging directory needs inspection; left untouched"
    if [ -f "$previous_stage/committed" ] || restore_transaction "$target_root" "$previous_stage"; then
      rm -rf -- "$previous_stage"
    else
      fail "previous update recovery failed; retained its backup"
    fi
  done
}

sync_target() {
  source_root=$1
  target_root=$2
  release_version=$(source_release_version "$source_root") || fail "official source has an invalid Skill version"
  [ -d "$target_root" ] || mkdir -p "$target_root"
  target_root=$(cd "$target_root" && pwd -P)
  check_gemini_duplicates
  acquire_update_lock
  # Downloads happen outside this lock. Re-read the installed version after
  # acquiring it so an older download cannot overwrite a newer completed sync.
  installed_manifest="$target_root/.puretokens-executor/runtime.json"
  if managed_executor "$target_root/.puretokens-executor"; then
    installed_version=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([0-9][0-9.]*\)".*/\1/p' "$installed_manifest" | sed -n '1p')
    printf '%s\n' "$installed_version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$' || fail "installed executor version is invalid; left untouched"
    if awk -v current="$installed_version" -v available="$release_version" 'BEGIN {split(current,c,"."); split(available,a,"."); for(i=1;i<=3;i++){if(c[i]+0>a[i]+0)exit 0;if(c[i]+0<a[i]+0)exit 1}exit 1}'; then
      fail "a newer executor version is already installed; downgrade was stopped under the update lock"
    fi
  fi

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
  touch "$stage_root/transaction-v1"
  mkdir -p "$stage_root/backup"
  for name in $current_skills; do cp -R "$source_root/skills/$name" "$stage_root/$name"; done
  executor_source=$(executor_artifact "$source_root")
  mkdir -p "$stage_root/.puretokens-executor"
  cp "$executor_source" "$stage_root/.puretokens-executor/puretokens-api"
  cp "$source_root/runtime/puretokens-skill-install.sh" "$source_root/runtime/puretokens-skill-fetch.sh" "$stage_root/.puretokens-executor/"
  chmod 700 "$stage_root/.puretokens-executor/puretokens-api"
  printf '{\n  "schemaVersion": 1,\n  "name": "puretokens-api-executor",\n  "version": "%s",\n  "platform": "%s"\n}\n' "$release_version" "$(executor_platform)" > "$stage_root/.puretokens-executor/runtime.json"

  migrate_legacy_codex_plugin "$target_root"

  : > "$stage_root/plan"
  for entry in $current_skills .puretokens-executor; do
    if [ -e "$target_root/$entry" ]; then action=replace; else action=create; fi
    printf '%s %s\n' "$action" "$entry" >> "$stage_root/plan"
  done
  while read -r action entry; do
    if [ "$action" = replace ]; then mv "$target_root/$entry" "$stage_root/backup/$entry"; fi
    mv "$stage_root/$entry" "$target_root/$entry"
  done < "$stage_root/plan"
  touch "$stage_root/committed"
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
  rm -rf -- "$stage_root"
  stage_root=
  release_update_lock
  trap - EXIT HUP INT TERM
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

[ -n "$target" ] || [ -n "$host" ] || fail "--host or --target is required"
case "$host" in ''|claude-code|codex|workbuddy|gemini-cli|grok-build|opencode|trae) ;; *) fail "unsupported host" ;; esac
[ -n "$target" ] || target=$(target_for_host "$host")
[ "${target#/}" != "$target" ] || fail "--target must be an absolute Skill directory"
[ -z "$source" ] || [ "${source#/}" != "$source" ] || fail "--source must be an absolute official source directory"
case "$command_name" in check|init|sync|locate) ;; *) usage; exit 2 ;; esac

if [ "$command_name" = locate ]; then
  printf '%s\n' "$target"
  exit 0
fi

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
