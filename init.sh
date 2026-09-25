#!/usr/bin/env bash

set -euo pipefail

usage() {
  printf 'Usage: %s "Title" "Description"\n' "${0##*/}" >&2
}

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

realpath_portable() {
  if command -v realpath >/dev/null 2>&1; then
    realpath "$1"
  else
    local path=$1 directory basename
    directory=$(cd "$(dirname "$path")" && pwd -P)
    basename=${path##*/}
    printf '%s/%s\n' "$directory" "$basename"
  fi
}

sed_inplace() {
  if sed --version >/dev/null 2>&1; then
    sed -i "$@"
  else
    sed -i '' "$@"
  fi
}

if [[ $# -ne 2 || -z $1 || -z $2 ]]; then
  usage
  exit 1
fi

title=$1
description=$2

if [[ $title =~ [[:cntrl:]] || $description =~ [[:cntrl:]] ]]; then
  fail 'Title and description cannot contain control characters.'
fi

root=$(git rev-parse --show-toplevel 2>/dev/null) || fail 'The initializer must be run from a Git repository.'
root=$(realpath_portable "$root")
script_path=$(realpath_portable "$0")
if [[ $(pwd -P) != "$root" || $script_path != "$root/init.sh" ]]; then
  fail 'Run ./init.sh from the repository root.'
fi
if [[ ! -d $root/.git || -L $root/.git ]]; then
  fail 'Only repositories with a normal .git directory are supported.'
fi
if [[ $(git rev-parse --is-inside-work-tree 2>/dev/null) != true || $(git rev-parse --is-bare-repository 2>/dev/null) != false ]]; then
  fail 'The initializer requires a non-bare Git repository.'
fi

sparse_config=$(git config --bool core.sparseCheckout || true)
sparse_cone=$(git config --bool core.sparseCheckoutCone || true)
sparse_index=$(git config --bool index.sparse || true)
if [[ $sparse_config == true || $sparse_cone == true || $sparse_index == true ]]; then
  fail 'Sparse checkouts and sparse indexes are not supported.'
fi
if git sparse-checkout list >/dev/null 2>&1; then
  fail 'Sparse checkouts are not supported.'
fi
worktree_count=$(git worktree list --porcelain | grep -c '^worktree ' || true)
if [[ $worktree_count -ne 1 ]]; then
  fail 'Linked worktrees are not supported.'
fi

worktree_status=$(git status --porcelain=v1 --untracked-files=all) || fail 'Unable to inspect the Git worktree.'
if [[ -n $worktree_status ]]; then
  fail 'Commit or discard all existing worktree changes before running the initializer.'
fi

git_name=$(git config user.name || true)
git_email=$(git config user.email || true)
if [[ -z $git_name || -z $git_email ]]; then
  fail 'Configure Git user.name and user.email before running the initializer.'
fi
if [[ $git_name =~ [[:cntrl:]] || $git_email =~ [[:cntrl:]] ]]; then
  fail 'Git user.name and user.email cannot contain control characters.'
fi

slug=$(printf '%s' "$title" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')
if [[ ! $slug =~ ^[a-z][a-z0-9-]*$ ]]; then
  fail 'Title must produce a package name beginning with a letter.'
fi
version=0.0.1
year=$(date +%Y)

escape_pattern() {
  local value=$1 output='' char index
  for ((index = 0; index < ${#value}; index++)); do
    char=${value:index:1}
    case $char in
      "\\"|'.'|'^'|'$'|'*'|'['|']'|'|'|'+'|'?'|'{'|'}'|'('|')') output+="\\$char" ;;
      *) output+=$char ;;
    esac
  done
  printf '%s' "$output"
}

escape_replacement() {
  local value=$1 output='' char index
  for ((index = 0; index < ${#value}; index++)); do
    char=${value:index:1}
    case $char in
      "\\"|'&'|'|') output+="\\$char" ;;
      *) output+=$char ;;
    esac
  done
  printf '%s' "$output"
}

replace_literal() {
  local old=$1 new=$2 pattern replacement file
  shift 2
  [[ -n $old ]] || return 0
  pattern=$(escape_pattern "$old")
  replacement=$(escape_replacement "$new")
  for file in "$@"; do
    [[ -f $file && ! -L $file ]] || continue
    sed_inplace "s|$pattern|$replacement|g" "$file"
  done
}

replace_tracked_text() {
  local old=$1 new=$2 file
  [[ -n $old && $old != "$new" ]] || return 0
  while IFS= read -r -d '' file; do
    [[ -f $file && ! -L $file ]] || continue
    case $file in
      server/connect/* | web/connect/* | client/src/connect/* | docs/openapi.yaml) continue ;;
    esac
    grep -IqF -- "$old" "$file" || continue
    replace_literal "$old" "$new" "$file"
  done < <(git ls-files -z)
}

json_escape() {
  local value=$1
  value=${value//\\/\\\\}
  value=${value//\"/\\\"}
  printf '%s' "$value"
}

nix_escape() {
  local value=$1 interpolation="\${"
  value=${value//\\/\\\\}
  value=${value//\"/\\\"}
  value=${value//"$interpolation"/\\$interpolation}
  printf '%s' "$value"
}

parse_remote() {
  local input=$1 remote_path authority
  parsed_web_url=
  parsed_host=
  parsed_repo_path=
  case $input in
    http://* | https://*)
      remote_path=${input#*://}
      authority=${remote_path%%/*}
      parsed_host=${authority##*@}
      parsed_repo_path=${remote_path#*/}
      parsed_web_url="https://${parsed_host}/${parsed_repo_path}"
      ;;
    ssh://*)
      remote_path=${input#ssh://}
      authority=${remote_path%%/*}
      parsed_host=${authority##*@}
      parsed_host=${parsed_host%%:*}
      parsed_repo_path=${remote_path#*/}
      parsed_web_url="https://${parsed_host}/${parsed_repo_path}"
      ;;
    *:*)
      authority=${input%%:*}
      parsed_host=${authority##*@}
      parsed_repo_path=${input#*:}
      parsed_web_url="https://${parsed_host}/${parsed_repo_path}"
      ;;
    *)
      return 1
      ;;
  esac
  [[ $input != *'?'* && $input != *'#'* ]] || return 1
  parsed_repo_path=${parsed_repo_path#/}
  parsed_repo_path=${parsed_repo_path%/}
  parsed_repo_path=${parsed_repo_path%.git}
  parsed_host=${parsed_host%/}
  [[ $parsed_repo_path == */* ]] || return 1
  [[ $parsed_host =~ ^[A-Za-z0-9.-]+(:[0-9]+)?$ ]] || return 1
  [[ $parsed_repo_path =~ ^[A-Za-z0-9._~/-]+$ ]] || return 1
  [[ $parsed_repo_path != *'//'* && $parsed_repo_path != *'/./'* && $parsed_repo_path != *'/../'* ]] || return 1
  parsed_web_url="https://${parsed_host}/${parsed_repo_path}"
}

origin=
while IFS= read -r remote_url; do
  origin=$remote_url
  break
done < <(git config --get-all remote.origin.url || true)
[[ -n $origin ]] || fail 'Unable to read the origin remote.'
parse_remote "$origin" || fail 'Unsupported origin URL. Use HTTPS, SCP, or ssh:// syntax.'
origin_host=$parsed_host
origin_repo_path=$parsed_repo_path
origin_web_url=$parsed_web_url
origin_provider_host=${origin_host%%:*}
origin_provider_lower=$(printf '%s' "$origin_provider_host" | tr '[:upper:]' '[:lower:]')
origin_is_github=false
[[ $origin_provider_lower == github.com ]] && origin_is_github=true

remote_names=()
while IFS= read -r remote_name; do
  remote_names+=("$remote_name")
done < <(git remote)
source_module=$(sed -n 's/^module[[:space:]]\+//p' server/go.mod 2>/dev/null | head -n 1 || true)
source_base=${source_module%/server}
if [[ -z $source_base || $source_base == "$source_module" ]]; then
  source_base=${origin_host}/${origin_repo_path}
fi
source_display=$(sed -n 's/^#[[:space:]]\+//p' README.md 2>/dev/null | head -n 1 || true)
[[ -n $source_display ]] || source_display=$(sed -n 's/^        name: "\([^"]*\)".*/\1/p' web/vite.config.ts 2>/dev/null | head -n 1 || true)
source_slug=$(sed -n 's/^  "name":[[:space:]]*"\([^"]*\)-web".*/\1/p' web/package.json 2>/dev/null | head -n 1 || true)
[[ -n $source_slug ]] || source_slug=$(sed -n 's/^[[:space:]]*pname = "\([^"]*\)-web";.*/\1/p' web/default.nix 2>/dev/null | head -n 1 || true)
source_description=$(sed -n 's/^  description = "\(.*\)";/\1/p' flake.nix 2>/dev/null | head -n 1 || true)
source_license_year=$(sed -n 's/^Copyright (c) \([0-9][0-9][0-9][0-9]\) .*/\1/p' LICENSE 2>/dev/null | head -n 1 || true)
[[ -n $source_license_year ]] || source_license_year=$year
[[ -f README.md && -f flake.nix && -f LICENSE && -f client/Cargo.toml && -f client/Cargo.lock && -f client/default.nix && -f docs/default.nix && -f docs/package.json && -f docs/package-lock.json && -f server/default.nix && -f server/go.mod && -f web/default.nix && -f web/package.json ]] || fail 'Required project metadata files are missing.'
[[ -f CONTRIBUTING.md && -f docs/openapi.base.yaml ]] || fail 'Required project documentation files are missing.'
[[ $(grep -c '^## using$' README.md) == 1 ]] || fail 'README.md must contain exactly one using section.'
[[ -n $source_module && $source_module == */server ]] || fail 'Unable to find the server module metadata.'
[[ -n $source_base ]] || fail 'Unable to find the source repository identity.'
[[ -n $source_display ]] || fail 'Unable to find the project display name.'
[[ -n $source_slug ]] || fail 'Unable to find the project package name.'
[[ -n $source_description ]] || fail 'Unable to find the project description.'
grep -qE '^Copyright \(c\) [0-9]{4} ' LICENSE || fail 'Unable to find the license identity metadata.'
source_github_repository=$(sed -n 's/^[[:space:]]*"repositories":[[:space:]]*\["\([^"]*\)"\].*/\1/p' .github/renovate.json 2>/dev/null | head -n 1 || true)
source_github_owner=${source_github_repository%%/*}
source_forgejo_repository=$(sed -n 's/^[[:space:]]*"repositories":[[:space:]]*\["\([^"]*\)"\].*/\1/p' .forgejo/renovate.json 2>/dev/null | head -n 1 || true)
source_forgejo_endpoint=$(sed -n 's/^[[:space:]]*"endpoint":[[:space:]]*"\([^"]*\)".*/\1/p' .forgejo/renovate.json 2>/dev/null | head -n 1 || true)
source_registry=$(grep -hE '^[[:space:]]*REGISTRY:[[:space:]]*[^[:space:]]+' .forgejo/workflows/*.yaml 2>/dev/null | sed -n 's/.*REGISTRY:[[:space:]]*//p' | head -n 1 || true)

secondary_remote=
secondary_name=
secondary_host=
secondary_repo_path=
delete_provider=
read_secondary_repository() {
  local provider=$1 input secondary_host_lower
  read -r -p "$provider repository URL: " input || input=
  parse_remote "$input" || fail 'Unsupported repository URL. Use HTTPS, SCP, or ssh:// syntax.'
  secondary_remote=$input
  secondary_host=$parsed_host
  secondary_repo_path=$parsed_repo_path
  secondary_host_lower=$(printf '%s' "${secondary_host%%:*}" | tr '[:upper:]' '[:lower:]')
  if [[ $provider == GitHub && $secondary_host_lower != github.com ]]; then
    fail 'GitHub repository URL must use github.com.'
  fi
  if [[ $provider == Forgejo && $secondary_host_lower == github.com ]]; then
    fail 'Forgejo repository URL cannot use github.com.'
  fi
  if [[ $provider == GitHub ]]; then
    secondary_name=github
  else
    secondary_name=forgejo
  fi
}

if $origin_is_github && [[ -d .forgejo ]]; then
  read -r -p 'GitHub origin detected. Delete .forgejo? [y/N] ' reply || reply=
  if [[ $reply =~ ^[Yy]$ ]]; then
    delete_provider=forgejo
  else
    read_secondary_repository Forgejo
  fi
elif ! $origin_is_github && [[ -d .github ]]; then
  read -r -p 'Non-GitHub origin detected. Delete .github? [y/N] ' reply || reply=
  if [[ $reply =~ ^[Yy]$ ]]; then
    delete_provider=github
  else
    read_secondary_repository GitHub
  fi
fi

delete_directories=()
[[ -z $delete_provider ]] || delete_directories+=(".$delete_provider")
for editor in Zed 'VS Code'; do
  if [[ $editor == Zed ]]; then
    editor_directory=.zed
  else
    editor_directory=.vscode
  fi
  read -r -p "Are you using $editor? [Y/n] " reply || reply=
  if [[ $reply =~ ^[Nn]([Oo])?$ ]]; then
    delete_directories+=("$editor_directory")
  fi
done

for directory in "${delete_directories[@]}"; do
  ignored_files=$(git ls-files --others --ignored --exclude-standard -- "$directory") || fail 'Unable to inspect ignored configuration files.'
  if [[ -n $ignored_files ]]; then
    fail "Remove ignored files under $directory before deleting that configuration."
  fi
done

if $origin_is_github; then
  github_repo_path=$origin_repo_path
  forgejo_host=
  forgejo_repo_path=
  if [[ -n $secondary_remote ]]; then
    forgejo_host=$secondary_host
    forgejo_repo_path=$secondary_repo_path
  fi
else
  forgejo_host=$origin_host
  forgejo_repo_path=$origin_repo_path
  github_repo_path=
  if [[ -n $secondary_remote ]]; then
    github_repo_path=$secondary_repo_path
  fi
fi

target_base=${origin_host}/${origin_repo_path}
target_module=${target_base}/server
nix_description=$(nix_escape "$description")
json_description=$(json_escape "$description")

repo_name=${root##*/}
parent=${root%/*}
new_git_dir="$parent/.${repo_name}.git-new.$$"
old_git_dir="$parent/.${repo_name}.git-old.$$"
if [[ -e $new_git_dir || -e $old_git_dir ]]; then
  fail 'Temporary Git directory already exists.'
fi
script_backup=$(mktemp)
cp "$script_path" "$script_backup"

if ! git init --bare --quiet --initial-branch=main "$new_git_dir"; then
  rm -rf "$new_git_dir" "$script_backup"
  fail 'Unable to create the new Git repository.'
fi
new_git=(git --git-dir="$new_git_dir" --work-tree="$root")
changes_started=false
swap_started=false
cleanup() {
  local status=$?
  trap - EXIT
  trap '' INT TERM HUP
  if $swap_started && [[ -e $old_git_dir ]]; then
    rm -rf "$root/.git"
    mv "$old_git_dir" "$root/.git"
    git -C "$root" restore --source=HEAD --staged --worktree -- . 2>/dev/null || true
    git -C "$root" clean -fd 2>/dev/null || true
    cp "$script_backup" "$script_path" 2>/dev/null || true
    chmod +x "$script_path" 2>/dev/null || true
  elif $changes_started; then
    git -C "$root" restore --source=HEAD --staged --worktree -- . 2>/dev/null || true
    git -C "$root" clean -fd 2>/dev/null || true
    cp "$script_backup" "$script_path" 2>/dev/null || true
    chmod +x "$script_path" 2>/dev/null || true
  fi
  rm -rf "$new_git_dir" "$old_git_dir" "$script_backup"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP

changes_started=true
for directory in "${delete_directories[@]}"; do
  rm -rf "$directory"
done
# Replace template names before inserting URLs that may contain those same names.
replace_tracked_text "$source_slug" "$slug"
replace_tracked_text "$source_display" "$title"
renamed_source_base=${source_base//"$source_slug"/"$slug"}
renamed_source_base=${renamed_source_base//"$source_display"/"$title"}
replace_tracked_text "$renamed_source_base" "$target_base"
# Rust imports use underscores even when the Cargo package name contains hyphens.
replace_literal "${slug}_client" "${slug//-/_}_client" client/src/main.rs

# Rewrite project metadata structurally, deriving old values from the checked-out template.
if [[ -n $source_description && $source_description != "$description" ]]; then
  replace_literal "$source_description" "$nix_description" flake.nix client/default.nix server/default.nix
  replace_literal "$source_description" "$json_description" client/Cargo.toml
fi
for package_dir in docs web; do
  if [[ -f $package_dir/package.json ]]; then
    package_old_description=$(sed -n 's/^  "description":[[:space:]]*"\(.*\)",/\1/p' "$package_dir/package.json" | head -n 1 || true)
    [[ -z $package_old_description ]] || replace_literal "$package_old_description" "$json_description" "$package_dir/package.json"
  fi
  if [[ -f $package_dir/package-lock.json ]]; then
    package_lock_old_description=$(sed -n 's/^      "description":[[:space:]]*"\(.*\)",/\1/p' "$package_dir/package-lock.json" | head -n 1 || true)
    [[ -z $package_lock_old_description ]] || replace_literal "$package_lock_old_description" "$json_description" "$package_dir/package-lock.json"
  fi
done

# Project versions are reset by matching their structural fields, not by relying on a stale template version.
for package_dir in docs web; do
  if [[ -f $package_dir/package.json ]]; then
    package_version_line=$(grep -n -m 1 -E '^  "version":[[:space:]]*"[^"]+",' "$package_dir/package.json" | cut -d: -f1 || true)
    [[ -n $package_version_line ]] || fail "Unable to find the $package_dir package version metadata."
    sed_inplace -E "${package_version_line}s|^  \"version\":[[:space:]]*\"[^\"]+\",$|  \"version\": \"$version\",|" "$package_dir/package.json"
  fi
  if [[ -f $package_dir/package-lock.json ]]; then
    package_lock_version_line=$(grep -n -m 1 -E '^  "version":[[:space:]]*"[^"]+",' "$package_dir/package-lock.json" | cut -d: -f1 || true)
    package_lock_root_version_line=$(grep -n -m 1 -E '^      "version":[[:space:]]*"[^"]+",' "$package_dir/package-lock.json" | cut -d: -f1 || true)
    [[ -n $package_lock_version_line && -n $package_lock_root_version_line ]] || fail "Unable to find the $package_dir lockfile version metadata."
    sed_inplace -E "${package_lock_version_line}s|^  \"version\":[[:space:]]*\"[^\"]+\",$|  \"version\": \"$version\",|" "$package_dir/package-lock.json"
    sed_inplace -E "${package_lock_root_version_line}s|^      \"version\":[[:space:]]*\"[^\"]+\",$|      \"version\": \"$version\",|" "$package_dir/package-lock.json"
  fi
done
client_version_line=$(grep -n -m 1 -E '^version[[:space:]]*=[[:space:]]*"[^"]+"$' client/Cargo.toml | cut -d: -f1 || true)
[[ -n $client_version_line ]] || fail 'Unable to find the client package version metadata.'
sed_inplace -E "${client_version_line}s|^version[[:space:]]*=[[:space:]]*\"[^\"]+\"$|version = \"$version\"|" client/Cargo.toml
client_lock_name_line=$(grep -n -m 1 -F "name = \"${slug}-client\"" client/Cargo.lock | cut -d: -f1 || true)
[[ -n $client_lock_name_line ]] || fail 'Unable to find the client lockfile package metadata.'
client_lock_version_line=$((client_lock_name_line + 1))
sed_inplace -E "${client_lock_version_line}s|^version[[:space:]]*=[[:space:]]*\"[^\"]+\"$|version = \"$version\"|" client/Cargo.lock
for nix_package_file in client/default.nix docs/default.nix web/default.nix server/default.nix; do
  sed_inplace -E 's/(^[[:space:]]+version = ")[^"]+(";)/\1'"$version"'\2/' "$nix_package_file"
done
reset_openapi_metadata() {
  local file=$1 version_line description_line replacement
  version_line=$(grep -n -m 1 -E '^[[:space:]]*version:' "$file" | cut -d: -f1 || true)
  description_line=$(grep -n -m 1 -E '^[[:space:]]*description:' "$file" | cut -d: -f1 || true)
  [[ -n $version_line && -n $description_line ]] || fail "Unable to find OpenAPI metadata in $file."
  sed_inplace -E "${version_line}s|^[[:space:]]*version:.*$|  version: $version|" "$file"
  replacement=$(escape_replacement "  description: \"$json_description\"")
  sed_inplace -E "${description_line}s|^[[:space:]]*description:.*$|$replacement|" "$file"
}
reset_openapi_metadata docs/openapi.base.yaml

# Provider-specific metadata and workflow paths.
if [[ -f .github/renovate.json ]]; then
  [[ -n $github_repo_path && -n $source_github_repository ]] && replace_literal "$source_github_repository" "$github_repo_path" .github/renovate.json
  [[ -n $github_repo_path && -n $source_github_owner && -n $slug ]] && replace_literal "${source_github_owner}/${slug}" "$github_repo_path" .github/renovate.json
fi
if [[ -f .forgejo/renovate.json ]]; then
  [[ -n $forgejo_repo_path && -n $source_forgejo_repository ]] && replace_literal "$source_forgejo_repository" "$forgejo_repo_path" .forgejo/renovate.json
  [[ -n $forgejo_host && -n $source_forgejo_endpoint ]] && replace_literal "$source_forgejo_endpoint" "https://${forgejo_host}/api/v1" .forgejo/renovate.json
fi
if [[ -d .forgejo && -n $forgejo_host ]]; then
  [[ -n $source_registry ]] && replace_tracked_text "REGISTRY: $source_registry" "REGISTRY: $forgejo_host"
  replace_tracked_text "//$source_registry/api/packages" "//$forgejo_host/api/packages"
fi
remove_check_blocks() {
  local key
  for key in "$@"; do
    sed_inplace "/^[[:space:]]*${key} = {$/,/^[[:space:]]*};$/d" flake.nix
  done
}
if [[ -n $delete_provider ]]; then
  if [[ $delete_provider == github ]]; then
    remove_check_blocks actions-gh renovate-gh
  else
    remove_check_blocks actions-fj renovate-fj
  fi
fi
if [[ ! -d .github && -f flake.nix ]]; then
  remove_check_blocks actions-gh renovate-gh
  sed_inplace '\|^[[:space:]]*\./\.github/workflows$|d' flake.nix
  sed_inplace '/^          renovate = {$/,/^          };/d' flake.nix
fi
if [[ ! -d .forgejo && -f flake.nix ]]; then
  remove_check_blocks actions-fj renovate-fj
  sed_inplace '\|^[[:space:]]*\./\.forgejo/workflows$|d' flake.nix
fi

if [[ -n $secondary_remote ]]; then
  # The selected second provider is deliberately the only newly-created remote.
  :
fi

# Rewrite the license holder without assuming the template's current year or name.
if [[ -f LICENSE ]]; then
  license_replacement=$(escape_replacement "Copyright (c) $year $git_name")
  sed_inplace -E "s|^Copyright \\(c\\) [0-9]{4} .*$|$license_replacement|" LICENSE
fi

# Rebuild the README header, retaining only usage and contributing documentation.
raw_url=
if $origin_is_github; then
  raw_url="https://raw.githubusercontent.com/${github_repo_path}/refs/heads/main"
  check_badge="[![check](${origin_web_url}/actions/workflows/check.yaml/badge.svg?branch=main)](${origin_web_url}/actions/workflows/check.yaml)"
  vulnerable_badge="[![vulnerable](${origin_web_url}/actions/workflows/vulnerable.yaml/badge.svg?branch=main)](${origin_web_url}/actions/workflows/vulnerable.yaml)"
else
  raw_url="${origin_web_url}/raw/branch/main"
  check_badge="[![check](${origin_web_url}/actions/workflows/check.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=check&labelColor=%23313244)](${origin_web_url}/actions?workflow=check.yaml)"
  vulnerable_badge="[![vulnerable](${origin_web_url}/actions/workflows/vulnerable.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=vulnerable&labelColor=%23313244)](${origin_web_url}/actions?workflow=vulnerable.yaml)"
fi
# The raw lockfile URL is nested inside the Shields endpoint URL.
encoded_lock_url="${raw_url}/flake.lock"
encoded_lock_url=${encoded_lock_url//:/%253A}
encoded_lock_url=${encoded_lock_url//\//%252F}
nixpkgs_badge="[![nixpkgs](https://img.shields.io/endpoint?url=https%3A%2F%2Fnix-shield.trev.zip%2Fbadge%3Furl%3D${encoded_lock_url}%26input%3Dnixpkgs&logoColor=%23bac2de&labelColor=%23313244&color=%235277C3)](https://nixos.org/)"
go_badge="[![go](<https://img.shields.io/badge/dynamic/regex?url=${raw_url}/server/go.mod&search=toolchain%20go(.*)&replace=%241&style=flat&logo=go&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%2300ADD8>)](https://go.dev/doc/devel/release)"
node_badge="[![node](https://img.shields.io/badge/dynamic/json?url=${raw_url}/web/package.json&query=%24.engines.node&logo=nodedotjs&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23339933)](https://nodejs.org/en/about/previous-releases)"
solid_badge="[![solidjs](https://img.shields.io/badge/dynamic/json?url=${raw_url}/web/package.json&query=%24.dependencies.solid-js&logo=solid&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%232C4F7C)](https://www.solidjs.com/)"
rust_badge="[![rust](https://img.shields.io/badge/dynamic/toml?url=${raw_url}/client/Cargo.toml&query=%24.package.rust-version&logo=rust&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23D34516)](https://releases.rs/)"

readme_sections=$(sed -n '/^## using$/,$p' README.md)
readme_image=$(printf '%s\n' "$readme_sections" | sed -n 's/^docker run -P \([^[:space:]]*\)$/\1/p')
[[ -n $readme_image ]] || fail 'Unable to find the README server image.'
image_path=$(printf '%s' "$origin_repo_path" | tr '[:upper:]' '[:lower:]')
if $origin_is_github; then
  image="ghcr.io/${image_path}/server:latest"
else
  image="${origin_host}/${image_path}/server:latest"
fi
readme_sections=${readme_sections//"$readme_image"/"$image"}

{
  printf '# %s\n\n' "$title"
  printf '%s\n' "$check_badge" "$vulnerable_badge" "$nixpkgs_badge" "$go_badge" "$node_badge" "$solid_badge" "$rust_badge"
  printf '\n%s\n\n%s\n' "$description" "$readme_sections"
} >README.md
if ! $origin_is_github && [[ -f web/layout/layout.tsx ]]; then
  replace_literal 'on GitHub' 'on Forgejo' web/layout/layout.tsx
fi

# Preserve all existing remote URLs and push URLs, then add the explicitly selected provider remote.
"${new_git[@]}" config core.bare false
"${new_git[@]}" config user.name "$git_name"
"${new_git[@]}" config user.email "$git_email"
for remote_name in "${remote_names[@]}"; do
  fetch_urls=()
  while IFS= read -r remote_url; do
    fetch_urls+=("$remote_url")
  done < <(git config --get-all "remote.${remote_name}.url" || true)
  push_urls=()
  while IFS= read -r remote_url; do
    push_urls+=("$remote_url")
  done < <(git config --get-all "remote.${remote_name}.pushurl" || true)
  ((${#fetch_urls[@]})) || continue
  "${new_git[@]}" remote add "$remote_name" "${fetch_urls[0]}"
  for remote_url in "${fetch_urls[@]:1}"; do
    "${new_git[@]}" remote set-url --add "$remote_name" "$remote_url"
  done
  for remote_url in "${push_urls[@]}"; do
    "${new_git[@]}" remote set-url --add --push "$remote_name" "$remote_url"
  done
done
if [[ -n $secondary_remote ]]; then
  if "${new_git[@]}" remote get-url "$secondary_name" >/dev/null 2>&1; then
    "${new_git[@]}" remote set-url "$secondary_name" "$secondary_remote"
  else
    "${new_git[@]}" remote add "$secondary_name" "$secondary_remote"
  fi
fi

"${new_git[@]}" add -A
"${new_git[@]}" rm --cached --ignore-unmatch -- init.sh >/dev/null 2>&1 || true
env \
  GIT_AUTHOR_NAME="$git_name" \
  GIT_AUTHOR_EMAIL="$git_email" \
  GIT_COMMITTER_NAME="$git_name" \
  GIT_COMMITTER_EMAIL="$git_email" \
  "${new_git[@]}" -c commit.gpgsign=false -c core.hooksPath=/dev/null commit --quiet --message 'chore: initialize project'

swap_started=true
mv "$root/.git" "$old_git_dir"
rm "$script_path" || fail 'Unable to remove init.sh.'
mv "$new_git_dir" "$root/.git" || fail 'Unable to activate the new Git repository.'

# Configure only after the new root repository is active; rollback restores the original repository on failure.
(cd "$root" && nix run .#configure)

# Include generated and formatted files in the same fresh root commit.
git -C "$root" add -A
git -C "$root" rm --cached --ignore-unmatch -- init.sh >/dev/null 2>&1 || true
env \
  GIT_AUTHOR_NAME="$git_name" \
  GIT_AUTHOR_EMAIL="$git_email" \
  GIT_COMMITTER_NAME="$git_name" \
  GIT_COMMITTER_EMAIL="$git_email" \
  git -C "$root" -c commit.gpgsign=false -c core.hooksPath=/dev/null commit --quiet --amend --no-edit

assert_contains() {
  local file=$1 expected=$2
  [[ -f $file ]] || fail "Expected metadata file is missing: $file"
  grep -qF -- "$expected" "$file" || fail "Replacement validation failed in $file."
}
assert_contains README.md "# $title"
assert_contains README.md "$description"
assert_contains README.md '## using'
assert_contains README.md '[CONTRIBUTING.md](CONTRIBUTING.md)'
assert_contains README.md "docker run -P $image"
assert_contains README.md "$nixpkgs_badge"
assert_contains README.md "$rust_badge"
assert_contains docs/openapi.base.yaml "  version: $version"
assert_contains flake.nix "description = \"$nix_description\";"
assert_contains client/Cargo.toml "name = \"${slug}-client\""
assert_contains client/Cargo.toml "version = \"$version\""
assert_contains client/Cargo.toml "description = \"$json_description\""
assert_contains client/Cargo.lock "name = \"${slug}-client\""
assert_contains client/Cargo.lock "version = \"$version\""
assert_contains client/default.nix "version = \"$version\";"
assert_contains client/default.nix "description = \"$nix_description\";"
assert_contains docs/default.nix "version = \"$version\";"
assert_contains web/default.nix "version = \"$version\";"
assert_contains server/default.nix "version = \"$version\";"
assert_contains server/default.nix "description = \"$nix_description\";"
assert_contains LICENSE "Copyright (c) $year $git_name"
[[ ! -e init.sh ]] || fail 'Replacement validation failed: init.sh still exists.'
[[ -z $(git -C "$root" ls-files --error-unmatch init.sh 2>/dev/null || true) ]] || fail 'Replacement validation failed: init.sh is still tracked.'
if [[ -f server/go.mod ]]; then
  assert_contains server/go.mod "module $target_module"
fi
if [[ -f docs/package.json ]]; then
  assert_contains docs/package.json "\"name\": \"${slug}-docs\""
fi
if [[ -f web/package.json ]]; then
  assert_contains web/package.json "\"name\": \"${slug}-web\""
fi
if [[ -n $source_base && $source_base != "$target_base" ]]; then
  while IFS= read -r -d '' file; do
    [[ -f $file && ! -L $file ]] || continue
    grep -qF -- "$source_base" "$file" && fail "Replacement validation found stale repository path in $file."
  done < <(git ls-files -z)
fi

rm -rf "$old_git_dir"
trap - EXIT INT TERM HUP
printf 'Initialized %s (%s) from %s with a new root commit.\n' "$title" "$version" "$origin_web_url"
