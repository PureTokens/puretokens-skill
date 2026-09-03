#!/bin/sh

set -eu

repository_archive_url="https://github.com/PureTokens/puretokens-skill/archive/refs/heads/main.zip"
current_skills="puretokens-balance puretokens-connection puretokens-models puretokens-image puretokens-video puretokens-update"
retired_skills="puretokens_media puretokens_balance puretokens_connection puretokens_models puretokens_image puretokens_video puretokens_update puretokens_get_balance puretokens_get_model_price puretokens_workbuddy_router"

usage() {
  printf '%s\n' "Usage: puretokens-skill-install.sh <check|sync> --target <absolute-skill-directory>"
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

validate_source() {
  source_root=$1
  [ -f "$source_root/README.md" ] || fail "official source is missing README.md"
  [ -f "$source_root/runtime/runtime.json" ] || fail "official source is missing the managed runtime manifest"
  managed_runtime "$source_root/runtime" || fail "official source has an invalid managed runtime"
  [ -f "$source_root/runtime/puretokens-skill-install.sh" ] || fail "official source is missing the native installer"
  [ -f "$source_root/runtime/puretokens-skill-install.ps1" ] || fail "official source is missing the Windows native installer"
  for name in $current_skills; do
    managed_skill "$source_root/skills/$name" "$name" || fail "official source has an invalid Skill: $name"
  done
}

download_source() {
  workspace=$1
  archive="$workspace/puretokens-skill-main.zip"
  unpacked="$workspace/unpacked"
  mkdir -p "$unpacked"
  curl --fail --location --proto '=https' --proto-redir '=https' --tlsv1.2 \
    "$repository_archive_url" --output "$archive" || fail "could not download the official Pure Tokens Skill source"
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
  [ -d "$target_root" ] || mkdir -p "$target_root"
  target_root=$(cd "$target_root" && pwd -P)

  for name in $retired_skills; do
    destination="$target_root/$name"
    [ ! -e "$destination" ] || managed_skill "$destination" "$name" ||
      fail "unmanaged retired Skill conflicts: $destination"
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
    destination="$target_root/$name"
    [ ! -e "$destination" ] && continue
    backup="$target_root/.${name}.retired-$(date +%s)-$$"
    mv "$destination" "$backup" || fail "could not archive retired Skill: $name"
    printf '%s\n' "Archived retired $name to $backup"
  done
  rm -rf -- "$stage_root"
  printf '%s\n' "Installed or upgraded official Pure Tokens Skills at $target_root"
}

command_name=${1:-}
[ -n "$command_name" ] || { usage; exit 2; }
shift
target=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --target)
      [ "$#" -ge 2 ] || fail "--target requires an absolute directory"
      target=$2
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done
[ -n "$target" ] && [ "${target#/}" != "$target" ] || fail "--target must be an absolute Skill directory"
case "$command_name" in check|sync) ;; *) usage; exit 2 ;; esac

require_command curl
require_command unzip
require_command mktemp
workspace=$(mktemp -d "${TMPDIR:-/tmp}/puretokens-skill.XXXXXX") || fail "could not create a private temporary directory"
cleanup_workspace() { rm -rf -- "$workspace"; }
trap cleanup_workspace EXIT HUP INT TERM
source_root=$(download_source "$workspace")
if [ "$command_name" = "check" ]; then
  printf '%s\n' "Official Pure Tokens Skill source passed native static validation."
else
  sync_target "$source_root" "$target"
fi
