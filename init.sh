#!/usr/bin/env bash
# The script often single-quotes text containing $ for printf, sed and Nix
# shellcheck disable=SC2016

# Initializes a project from this template. init.test.sh tests it.

set -euo pipefail

template_url=https://trev.zip/template/stack
version=0.0.1
year=$(date +%Y)

# Forges, by the name of their configuration directory
forges=(github gitea forgejo)

die() {
  printf '%s\n' "$1" >&2
  exit 1
}

# macOS has neither GNU realpath nor GNU sed.
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

# Sets remote_host, remote_repo_path and remote_web_url from a Git remote URL.
# The label names the URL in error messages.
parse_remote() {
  local label=$1 url=$2 remote_path authority
  if [[ $url == *'?'* || $url == *'#'* ]]; then
    die "$label cannot contain a query string or fragment."
  fi
  case $url in
    http://* | https://*)
      remote_path=${url#*://}
      authority=${remote_path%%/*}
      remote_host=${authority##*@}
      remote_repo_path=${remote_path#*/}
      ;;
    ssh://*)
      remote_path=${url#ssh://}
      authority=${remote_path%%/*}
      remote_host=${authority##*@}
      remote_host=${remote_host%%:*}
      remote_repo_path=${remote_path#*/}
      ;;
    *://*)
      die "$label must use HTTPS, SCP, or ssh:// syntax."
      ;;
    *:*)
      authority=${url%%:*}
      remote_host=${authority##*@}
      remote_repo_path=${url#*:}
      ;;
    *)
      die "$label must use HTTPS, SCP, or ssh:// syntax."
      ;;
  esac
  remote_repo_path=${remote_repo_path#/}
  remote_repo_path=${remote_repo_path%/}
  remote_repo_path=${remote_repo_path%.git}
  if [[ $remote_repo_path != */* ]]; then
    die "$label must identify an owner and repository."
  fi
  if [[ ! $remote_host =~ ^[A-Za-z0-9.-]+(:[0-9]+)?$ || ! $remote_repo_path =~ ^[A-Za-z0-9._~/-]+$ ||
    $remote_repo_path == *'//'* || $remote_repo_path == *'/./'* || $remote_repo_path == *'/../'* ]]; then
    die "$label contains unsupported host or path characters."
  fi
  remote_web_url="https://$remote_host/$remote_repo_path"
}

# Prints a forge's display name.
forge_name() {
  case $1 in
    github) printf 'GitHub\n' ;;
    gitea) printf 'Gitea\n' ;;
    forgejo) printf 'Forgejo\n' ;;
  esac
}

# Prints the suffix of a forge's flake checks.
forge_check() {
  case $1 in
    github) printf 'gh\n' ;;
    gitea) printf 'gt\n' ;;
    forgejo) printf 'fj\n' ;;
  esac
}

# Checks that a repository URL's host suits its forge: GitHub repositories are
# on github.com, and Gitea and Forgejo repositories are anywhere else.
check_forge_host() {
  local label=$1 forge=$2 host
  host=$(printf '%s' "${3%%:*}" | tr '[:upper:]' '[:lower:]')
  if [[ $forge == github && $host != github.com ]]; then
    die "$label must use github.com."
  fi
  if [[ $forge != github && $host == github.com ]]; then
    die "$label cannot use github.com."
  fi
}

# Prints the package name for a project title.
slugify() {
  local slug
  slug=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')
  if [[ ! $slug =~ ^[a-z][a-z0-9-]*$ ]]; then
    die 'Title must produce a package name beginning with a letter.'
  fi
  printf '%s\n' "$slug"
}

# Prints a value escaped for a double-quoted Nix string.
nix_string() {
  local value=$1
  # shellcheck disable=SC1003
  value=${value//'\'/'\\'}
  value=${value//'"'/'\"'}
  value=${value//'${'/'\${'}
  printf '%s\n' "$value"
}

# Prints a value escaped for a double-quoted JSON or TOML string.
json_string() {
  local value=$1
  # shellcheck disable=SC1003
  value=${value//'\'/'\\'}
  value=${value//'"'/'\"'}
  printf '%s\n' "$value"
}

# Prints a value escaped for the pattern of an s||| command.
escape_pattern() {
  local value=$1 output='' char index
  for ((index = 0; index < ${#value}; index++)); do
    char=${value:index:1}
    case $char in
      "\\" | '.' | '^' | '$' | '*' | '[' | ']' | '|' | '+' | '?' | '{' | '}' | '(' | ')') output+="\\$char" ;;
      *) output+=$char ;;
    esac
  done
  printf '%s' "$output"
}

# Prints a value escaped for the replacement of an s||| command.
escape_replacement() {
  local value=$1 output='' char index
  for ((index = 0; index < ${#value}; index++)); do
    char=${value:index:1}
    case $char in
      "\\" | '&' | '|') output+="\\$char" ;;
      *) output+=$char ;;
    esac
  done
  printf '%s' "$output"
}

# Replaces every occurrence of old with new in the given files, skipping
# missing files and symlinks.
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

# Replaces every occurrence of old with new in the tracked text files, except
# for the generated code, which the configure app regenerates.
replace_tracked_text() {
  local old=$1 new=$2 file
  [[ -n $old && $old != "$new" ]] || return 0
  while IFS= read -r -d '' file; do
    [[ -f $file && ! -L $file ]] || continue
    case $file in
      server/gen/* | web/gen/* | client/src/gen/*) continue ;;
    esac
    grep -IqF -- "$old" "$file" || continue
    replace_literal "$old" "$new" "$file"
  done < <(git ls-files -z)
}

# Deletes from the line equal to $2 through the line equal to $3, or through the
# next blank line when $3 is omitted, without leaving doubled or trailing blank
# lines behind.
remove_block() {
  local file=$1 temp
  temp=$(mktemp)
  if ! START=$2 END=${3-} awk '
    state == 1 {
      if (ENVIRON["END"] == "") { if ($0 == "") state = 0 }
      else if ($0 == ENVIRON["END"]) state = 2
      next
    }
    state == 2 { state = 0; if ($0 == "") next }
    !found && $0 == ENVIRON["START"] { found = 1; state = 1; next }
    $0 == "" { blanks++; next }
    { for (; blanks > 0; blanks--) print ""; print }
    END {
      if (state == 0) for (; blanks > 0; blanks--) print ""
      exit !found || (state == 1 && ENVIRON["END"] != "")
    }
  ' "$file" >"$temp"; then
    rm -f "$temp"
    die "Unable to find the \"$2\" block in $file."
  fi
  cat "$temp" >"$file"
  rm -f "$temp"
}

# Removes the named checks from flake.nix.
remove_checks() {
  local key
  for key in "$@"; do
    # Anchor both ends at the check indentation so nested attrsets (e.g. env) do not end the range early.
    grep -qFx "          $key = {" flake.nix || continue
    remove_block flake.nix "          $key = {" '          };'
  done
}

# Checks that the file contains the text.
require_text() {
  local file=$1 expected=$2
  [[ -f $file ]] || die "Expected metadata file is missing: $file"
  grep -qF -- "$expected" "$file" || die "Replacement validation failed in $file."
}

# Runs a Git command as the Git user, whatever the environment sets.
git_as_user() {
  env \
    GIT_AUTHOR_NAME="$git_name" \
    GIT_AUTHOR_EMAIL="$git_email" \
    GIT_COMMITTER_NAME="$git_name" \
    GIT_COMMITTER_EMAIL="$git_email" \
    "$@"
}

# Prints the template commit this project started from. Repositories created
# with "use this template" have a new commit, but its tree matches the
# template's.
find_template_commit() {
  local template_trees tree commit
  if ! GIT_TERMINAL_PROMPT=0 git fetch --quiet --no-tags "$template_url" main 2>/dev/null; then
    return
  fi
  # Keep the newest commit for each tree.
  template_trees=$(git log --format='%T %H' FETCH_HEAD | awk '!seen[$1]++')
  while IFS= read -r tree; do
    # nix flake init drops the executable bit from init.sh
    tree=$(git ls-tree "$tree" | awk -F '\t' '$2 == "init.sh" { sub(/^100644 /, "100755 ") } 1' | git mktree --missing)
    commit=$(printf '%s\n' "$template_trees" | awk -v tree="$tree" '$1 == tree { print $2 }')
    if [[ -n $commit ]]; then
      printf '%s\n' "$commit"
      return
    fi
  done < <(git log --format='%T' HEAD)
  return 0
}

# Checks that the initializer runs from the root of a regular Git repository
# with a Git user.
check_repository() {
  local sparse_config sparse_cone sparse_index worktree_count
  root=$(git rev-parse --show-toplevel 2>/dev/null) || die 'The initializer must be run from a Git repository.'
  root=$(realpath_portable "$root")
  script_path=$(realpath_portable "$0")
  if [[ $(pwd -P) != "$root" || $script_path != "$root/init.sh" ]]; then
    die 'Run ./init.sh from the repository root.'
  fi
  if [[ ! -d $root/.git || -L $root/.git ]]; then
    die 'Only repositories with a normal .git directory are supported.'
  fi
  if [[ $(git rev-parse --is-inside-work-tree 2>/dev/null) != true || $(git rev-parse --is-bare-repository 2>/dev/null) != false ]]; then
    die 'The initializer requires a non-bare Git repository.'
  fi

  sparse_config=$(git config --bool core.sparseCheckout || true)
  sparse_cone=$(git config --bool core.sparseCheckoutCone || true)
  sparse_index=$(git config --bool index.sparse || true)
  if [[ $sparse_config == true || $sparse_cone == true || $sparse_index == true ]]; then
    die 'Sparse checkouts and sparse indexes are not supported.'
  fi
  if git sparse-checkout list >/dev/null 2>&1; then
    die 'Sparse checkouts are not supported.'
  fi
  worktree_count=$(git worktree list --porcelain | grep -c '^worktree ' || true)
  if [[ $worktree_count -ne 1 ]]; then
    die 'Linked worktrees are not supported.'
  fi

  git_name=$(git config user.name || true)
  git_email=$(git config user.email || true)
  if [[ -z $git_name || -z $git_email ]]; then
    die 'Configure Git user.name and user.email before running the initializer.'
  fi
  if [[ $git_name =~ [[:cntrl:]] || $git_email =~ [[:cntrl:]] ]]; then
    die 'Git user.name and user.email cannot contain control characters.'
  fi
}

# Sets forge and origin_is_github from the forge the user says hosts origin.
# The default is GitHub for github.com, and Forgejo for any other host.
read_forge() {
  local default=forgejo reply
  if [[ $origin_provider_lower == github.com ]]; then
    default=github
  fi
  read -r -p "Which forge hosts origin? [github/gitea/forgejo] ($default) " reply || reply=
  forge=$(printf '%s' "$reply" | tr '[:upper:]' '[:lower:]')
  forge=${forge:-$default}
  case $forge in
    github | gitea | forgejo) ;;
    *) die 'Forge must be github, gitea or forgejo.' ;;
  esac
  check_forge_host "$(forge_name "$forge") origin URL" "$forge" "$origin_host"
  origin_is_github=false
  if [[ $forge == github ]]; then
    origin_is_github=true
  fi
}

# Checks the title and description, and reads the project's location from the
# origin remote.
check_project() {
  local remote_url remote_name
  if [[ $title =~ [[:cntrl:]] || $description =~ [[:cntrl:]] ]]; then
    die 'Title and description cannot contain control characters.'
  fi
  slug=$(slugify "$title")
  nix_description=$(nix_string "$description")
  json_description=$(json_string "$description")

  origin=
  while IFS= read -r remote_url; do
    origin=$remote_url
    break
  done < <(git config --get-all remote.origin.url || true)
  [[ -n $origin ]] || die 'Unable to read the origin remote.'
  parse_remote 'Origin URL' "$origin"
  origin_host=$remote_host
  origin_repo_path=$remote_repo_path
  origin_web_url=$remote_web_url
  origin_provider_host=${origin_host%%:*}
  origin_provider_lower=$(printf '%s' "$origin_provider_host" | tr '[:upper:]' '[:lower:]')
  read_forge

  remote_names=()
  while IFS= read -r remote_name; do
    remote_names+=("$remote_name")
  done < <(git remote)
}

# Checks that the worktree is clean.
check_worktree() {
  local worktree_status
  # Repositories set up with nix flake init have no commits yet. Commit the
  # template so a failed run can be rolled back; the new repository replaces it.
  if ! git rev-parse --verify --quiet HEAD >/dev/null; then
    git add -A
    git_as_user git -c commit.gpgsign=false -c core.hooksPath=/dev/null commit --quiet --message 'chore: add template'
  fi

  worktree_status=$(git status --porcelain=v1 --untracked-files=all) || die 'Unable to inspect the Git worktree.'
  if [[ -n $worktree_status ]]; then
    die 'Commit or discard all existing worktree changes before running the initializer.'
  fi
}

# Reads the template's own names and metadata from the checked-out files, so
# the script doesn't depend on what the template is currently called.
read_template_metadata() {
  local file
  for file in README.md flake.nix LICENSE client/Cargo.toml client/Cargo.lock client/default.nix server/default.nix server/go.mod web/default.nix web/package.json; do
    [[ -f $file ]] || die 'Required project metadata files are missing.'
  done
  [[ -f CONTRIBUTING.md ]] || die 'Required project documentation files are missing.'

  source_module=$(sed -n 's/^module[[:space:]]\+//p' server/go.mod | head -n 1 || true)
  # Go adds a /vN suffix to the module path from major version 2, which new
  # projects drop because their versions are reset.
  source_server_module=$source_module
  if [[ $source_module =~ ^(.+/server)/v[0-9]+$ ]]; then
    source_server_module=${BASH_REMATCH[1]}
  fi
  source_base=${source_server_module%/server}
  if [[ -z $source_base || $source_base == "$source_module" ]]; then
    source_base=${origin_host}/${origin_repo_path}
  fi
  source_display=$(sed -En 's/^#[[:space:]]+(<img[^>]*>[[:space:]]*)?//p' README.md | head -n 1 || true)
  [[ -n $source_display ]] || source_display=$(sed -n 's/^        name: "\([^"]*\)".*/\1/p' web/vite.config.ts 2>/dev/null | head -n 1 || true)
  source_slug=$(sed -n 's/^  "name":[[:space:]]*"\([^"]*\)-web".*/\1/p' web/package.json | head -n 1 || true)
  [[ -n $source_slug ]] || source_slug=$(sed -n 's/^[[:space:]]*pname = "\([^"]*\)-web";.*/\1/p' web/default.nix | head -n 1 || true)
  source_description=$(sed -n 's/^  description = "\(.*\)";/\1/p' flake.nix | head -n 1 || true)
  [[ $(grep -c '^## using$' README.md) == 1 ]] || die 'README.md must contain exactly one using section.'
  [[ -n $source_module && $source_server_module == */server ]] || die 'Unable to find the server module metadata.'
  [[ -n $source_base ]] || die 'Unable to find the source repository identity.'
  [[ -n $source_display ]] || die 'Unable to find the project display name.'
  [[ -n $source_slug ]] || die 'Unable to find the project package name.'
  [[ -n $source_description ]] || die 'Unable to find the project description.'
  grep -qE '^Copyright \(c\) [0-9]{4} ' LICENSE || die 'Unable to find the license identity metadata.'
  source_github_repository=$(sed -n 's/^[[:space:]]*"repositories":[[:space:]]*\["\([^"]*\)"\].*/\1/p' .github/renovate.jsonc 2>/dev/null | head -n 1 || true)
  source_github_owner=${source_github_repository%%/*}
  source_forgejo_repository=$(sed -n 's/^[[:space:]]*"repositories":[[:space:]]*\["\([^"]*\)"\].*/\1/p' .forgejo/renovate.jsonc 2>/dev/null | head -n 1 || true)
  source_forgejo_endpoint=$(sed -n 's/^[[:space:]]*"endpoint":[[:space:]]*"\([^"]*\)".*/\1/p' .forgejo/renovate.jsonc 2>/dev/null | head -n 1 || true)
  source_gitea_repository=$(sed -n 's/^[[:space:]]*"repositories":[[:space:]]*\["\([^"]*\)"\].*/\1/p' .gitea/renovate.jsonc 2>/dev/null | head -n 1 || true)
  source_gitea_endpoint=$(sed -n 's/^[[:space:]]*"endpoint":[[:space:]]*"\([^"]*\)".*/\1/p' .gitea/renovate.jsonc 2>/dev/null | head -n 1 || true)
  source_registry=$(grep -hE '^[[:space:]]*REGISTRY:[[:space:]]*[^[:space:]]+' .forgejo/workflows/*.yaml .gitea/workflows/*.yaml 2>/dev/null | sed -n 's/.*REGISTRY:[[:space:]]*//p' | head -n 1 || true)
}

# Sets secondary_remote, secondary_host and secondary_repo_path from the URL of
# the project's repository on another forge.
read_secondary_repository() {
  local forge=$1 label
  label="$(forge_name "$1") repository URL"
  read -r -p "$label: " secondary_remote || secondary_remote=
  parse_remote 'Repository URL' "$secondary_remote"
  check_forge_host "$label" "$forge" "$remote_host"
  secondary_host=$remote_host
  secondary_repo_path=$remote_repo_path
}

# Records the project's repository on a forge.
set_forge_repository() {
  case $1 in
    github) github_repo_path=$3 ;;
    gitea)
      gitea_host=$2
      gitea_repo_path=$3
      ;;
    forgejo)
      forgejo_host=$2
      forgejo_repo_path=$3
      ;;
  esac
}

# Asks whether the project also uses each other forge, and for its repository
# if so. The configuration of forges it doesn't use is deleted.
choose_providers() {
  local other reply
  mirror_names=()
  mirror_remotes=()
  github_repo_path=
  gitea_host=
  gitea_repo_path=
  forgejo_host=
  forgejo_repo_path=
  set_forge_repository "$forge" "$origin_host" "$origin_repo_path"
  for other in "${forges[@]}"; do
    [[ $other != "$forge" && -d .$other ]] || continue
    read -r -p "Also use $(forge_name "$other")? [y/N] " reply || reply=
    if [[ $reply =~ ^[Yy]([Ee][Ss])?$ ]]; then
      read_secondary_repository "$other"
      mirror_names+=("$other")
      mirror_remotes+=("$secondary_remote")
      set_forge_repository "$other" "$secondary_host" "$secondary_repo_path"
    else
      delete_directories+=(".$other")
    fi
  done
}

choose_editors() {
  local reply
  read -r -p 'Are you using Zed? [Y/n] ' reply || reply=
  if [[ $reply =~ ^[Nn]([Oo])?$ ]]; then
    delete_directories+=(.zed)
  fi

  read -r -p 'Are you using VS Code? [Y/n] ' reply || reply=
  if [[ $reply =~ ^[Nn]([Oo])?$ ]]; then
    delete_directories+=(.vscode)
  fi
}

# Asks which of the optional components to keep.
choose_components() {
  local reply
  keep_client=true
  read -r -p 'Keep the Rust client? [Y/n] ' reply || reply=
  if [[ $reply =~ ^[Nn]([Oo])?$ ]]; then
    keep_client=false
    # The Python integration tests exist only to drive the client binary.
    delete_directories+=(client test)
  fi
}

# Checks that the directories to delete hold nothing but build outputs.
check_ignored_files() {
  local directory ignored_files
  for directory in "${delete_directories[@]}"; do
    # Build outputs are safe to discard along with the directories that produced them.
    ignored_files=$(git ls-files --others --ignored --exclude-standard -- "$directory" ':(exclude)client/target' ':(exclude,glob)**/__pycache__/**' ':(exclude,glob)**/.ruff_cache/**') || die 'Unable to inspect ignored files.'
    if [[ -n $ignored_files ]]; then
      die "Remove ignored files under $directory before deleting it."
    fi
  done
}

# Works out the project's module path and package registry. The registry is
# the origin's unless that is GitHub, then a Forgejo or Gitea mirror's.
resolve_repositories() {
  if ! $origin_is_github; then
    registry_host=$origin_host
  elif [[ -n $forgejo_host ]]; then
    registry_host=$forgejo_host
  else
    registry_host=$gitea_host
  fi
  target_base=${origin_host}/${origin_repo_path}
  target_module=${target_base}/server
}

# Creates the repository that replaces .git once the project is initialized,
# and restores the template if initialization fails.
start_new_repository() {
  local repo_name=${root##*/}
  new_git_dir="${root%/*}/.${repo_name}.git-new.$$"
  old_git_dir="${root%/*}/.${repo_name}.git-old.$$"
  if [[ -e $new_git_dir || -e $old_git_dir ]]; then
    die 'Temporary Git directory already exists.'
  fi
  if ! git init --bare --quiet --initial-branch=main "$new_git_dir"; then
    rm -rf "$new_git_dir"
    die 'Unable to create the new Git repository.'
  fi
  new_git=(git --git-dir="$new_git_dir" --work-tree="$root")
  swap_started=false
  changes_started=false
  trap cleanup EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM
  trap 'exit 129' HUP
}

cleanup() {
  local status=$?
  trap - EXIT
  trap '' INT TERM HUP
  if $swap_started && [[ -e $old_git_dir ]]; then
    rm -rf "$root/.git"
    mv "$old_git_dir" "$root/.git"
  fi
  if $changes_started; then
    # This also restores init.sh
    git -C "$root" restore --source=HEAD --staged --worktree -- . 2>/dev/null || true
    git -C "$root" clean -fdq 2>/dev/null || true
  fi
  rm -rf "$new_git_dir" "$old_git_dir"
  exit "$status"
}

# Removes the template's own files and records the template commit.
remove_template_files() {
  local directory
  rm template.schema.json init.test.sh
  sed_inplace '/^        # nix flake init /,/^        };$/d' flake.nix
  remove_checks init
  remove_block .gitattributes '# Linguist: exclude template initializer from language stats'
  if [[ -n $template_commit ]]; then
    printf '{ "$schema": "%s", "url": "%s", "commit": "%s" }\n' \
      "$template_url/raw/branch/main/template.schema.json" "$template_url" "$template_commit" >.template.json
  else
    printf 'Warning: unable to find the template commit; .template.json was not created.\n' >&2
  fi
  for directory in "${delete_directories[@]}"; do
    rm -rf "$directory"
  done
}

# Replaces the template's name and repository with the project's.
rename_project() {
  local renamed_source_base
  replace_tracked_text "$source_module" "$source_server_module"
  # Replace template names before inserting URLs that may contain those same names.
  replace_tracked_text "$source_slug" "$slug"
  replace_tracked_text "$source_display" "$title"
  renamed_source_base=${source_base//"$source_slug"/"$slug"}
  renamed_source_base=${renamed_source_base//"$source_display"/"$title"}
  replace_tracked_text "$renamed_source_base" "$target_base"
  # Rust imports use underscores even when the Cargo package name contains hyphens.
  replace_literal "${slug}_client" "${slug//-/_}_client" client/src/main.rs
}

# Replaces the template's description with the project's.
set_descriptions() {
  local old_description
  if [[ -n $source_description && $source_description != "$description" ]]; then
    replace_literal "$source_description" "$nix_description" flake.nix client/default.nix server/default.nix
    replace_literal "$source_description" "$json_description" client/Cargo.toml
  fi
  old_description=$(sed -n 's/^  "description":[[:space:]]*"\(.*\)",/\1/p' web/package.json | head -n 1 || true)
  [[ -z $old_description ]] || replace_literal "$old_description" "$json_description" web/package.json
  if [[ -f web/package-lock.json ]]; then
    old_description=$(sed -n 's/^      "description":[[:space:]]*"\(.*\)",/\1/p' web/package-lock.json | head -n 1 || true)
    [[ -z $old_description ]] || replace_literal "$old_description" "$json_description" web/package-lock.json
  fi
}

# Project versions are reset by matching their structural fields, not by relying
# on a stale template version.
reset_versions() {
  local version_line root_version_line client_lock_name_line nix_package_file
  local nix_package_files=(web/default.nix server/default.nix)
  version_line=$(grep -n -m 1 -E '^  "version":[[:space:]]*"[^"]+",' web/package.json | cut -d: -f1 || true)
  [[ -n $version_line ]] || die 'Unable to find the web package version metadata.'
  sed_inplace -E "${version_line}s|^  \"version\":[[:space:]]*\"[^\"]+\",$|  \"version\": \"$version\",|" web/package.json
  if [[ -f web/package-lock.json ]]; then
    version_line=$(grep -n -m 1 -E '^  "version":[[:space:]]*"[^"]+",' web/package-lock.json | cut -d: -f1 || true)
    root_version_line=$(grep -n -m 1 -E '^      "version":[[:space:]]*"[^"]+",' web/package-lock.json | cut -d: -f1 || true)
    [[ -n $version_line && -n $root_version_line ]] || die 'Unable to find the web lockfile version metadata.'
    sed_inplace -E "${version_line}s|^  \"version\":[[:space:]]*\"[^\"]+\",$|  \"version\": \"$version\",|" web/package-lock.json
    sed_inplace -E "${root_version_line}s|^      \"version\":[[:space:]]*\"[^\"]+\",$|      \"version\": \"$version\",|" web/package-lock.json
  fi
  if $keep_client; then
    version_line=$(grep -n -m 1 -E '^version[[:space:]]*=[[:space:]]*"[^"]+"$' client/Cargo.toml | cut -d: -f1 || true)
    [[ -n $version_line ]] || die 'Unable to find the client package version metadata.'
    sed_inplace -E "${version_line}s|^version[[:space:]]*=[[:space:]]*\"[^\"]+\"$|version = \"$version\"|" client/Cargo.toml
    client_lock_name_line=$(grep -n -m 1 -F "name = \"${slug}-client\"" client/Cargo.lock | cut -d: -f1 || true)
    [[ -n $client_lock_name_line ]] || die 'Unable to find the client lockfile package metadata.'
    sed_inplace -E "$((client_lock_name_line + 1))s|^version[[:space:]]*=[[:space:]]*\"[^\"]+\"$|version = \"$version\"|" client/Cargo.lock
    nix_package_files+=(client/default.nix)
  fi
  for nix_package_file in "${nix_package_files[@]}"; do
    sed_inplace -E 's/(^[[:space:]]+version = ")[^"]+(";)/\1'"$version"'\2/' "$nix_package_file"
  done
}

# Points a Forgejo or Gitea Renovate configuration at the project's repository.
configure_forge_renovate() {
  local file=$1 host=$2 repo_path=$3 source_repository=$4 source_endpoint=$5
  [[ -f $file ]] || return 0
  [[ -n $repo_path && -n $source_repository ]] && replace_literal "$source_repository" "$repo_path" "$file"
  [[ -n $host && -n $source_endpoint ]] && replace_literal "$source_endpoint" "https://${host}/api/v1" "$file"
  return 0
}

# Points each forge's configuration at the project's repositories, and removes
# the checks of forges whose configuration was deleted.
configure_providers() {
  local other check
  if [[ -f .github/renovate.jsonc ]]; then
    [[ -n $github_repo_path && -n $source_github_repository ]] && replace_literal "$source_github_repository" "$github_repo_path" .github/renovate.jsonc
    [[ -n $github_repo_path && -n $source_github_owner && -n $slug ]] && replace_literal "${source_github_owner}/${slug}" "$github_repo_path" .github/renovate.jsonc
  fi
  configure_forge_renovate .forgejo/renovate.jsonc "$forgejo_host" "$forgejo_repo_path" "$source_forgejo_repository" "$source_forgejo_endpoint"
  configure_forge_renovate .gitea/renovate.jsonc "$gitea_host" "$gitea_repo_path" "$source_gitea_repository" "$source_gitea_endpoint"
  if [[ -n $registry_host ]] && [[ -d .forgejo || -d .gitea ]]; then
    [[ -n $source_registry ]] && replace_tracked_text "REGISTRY: $source_registry" "REGISTRY: $registry_host"
    replace_tracked_text "//$source_registry/api/packages" "//$registry_host/api/packages"
  fi

  [[ -f flake.nix ]] || return 0
  for other in "${forges[@]}"; do
    [[ ! -d .$other ]] || continue
    check=$(forge_check "$other")
    remove_checks "actions-$check" "renovate-$check"
    sed_inplace "\\|^[[:space:]]*\\./\\.$other/workflows\$|d" flake.nix
    if [[ $other == github ]]; then
      sed_inplace '/^          renovate = {$/,/^          };/d' flake.nix
    fi
  done
}

# Removes the Rust client and the Python tests that drive it.
remove_client() {
  local provider_directory cargo_rule_line
  rm -f pyproject.toml
  sed_inplace \
    -e '/^            RUST_SRC_PATH = /d' \
    -e '/^              protoc-gen-prost$/d' \
    -e '/^              trevrpc\.packages\.\${system}\.trevrpc-rust$/d' \
    -e '/^              cargo # cargo update$/d' \
    -e '/^              cargo-audit # cargo-audit audit$/d' \
    -e '/^            rustfmt$/d' \
    -e '/^            ruff$/d' \
    -e '/^          client = pkgs\.callPackage \.\/client { };$/d' \
    -e '/^          client = pkgs\.mkImage {$/,/^          };$/d' \
    -e '/^            client$/d' \
    flake.nix
  remove_block flake.nix '              # rust'
  remove_block flake.nix '              # python'
  remove_checks client-lint integration python-lint
  remove_block buf.gen.yaml '  - local: protoc-gen-prost'
  remove_block buf.gen.yaml '  - local: protoc-gen-trevrpc-rust'
  remove_block treefmt.toml '[formatter.rustfmt]'
  remove_block treefmt.toml '[formatter.ruff]'
  sed_inplace '/^  "client\/src\/gen\/\*\*",$/d' treefmt.toml
  remove_block .gitignore '# client'
  remove_block .gitignore '# python'
  remove_block .gitattributes '# Rust source files use Rust diff driver'
  sed_inplace -e '/^Cargo\.lock linguist-generated=true$/d' -e '/^client\/src\/gen\/\*\* linguist-generated=true$/d' .gitattributes
  replace_literal '`server/gen/`, `web/gen/`, or `client/src/gen/`' '`server/gen/` or `web/gen/`' AGENTS.md
  remove_block CONTRIBUTING.md 'run the client with optional arguments:' '```'
  sed_inplace '/^nix build \.#client$/d' CONTRIBUTING.md
  for provider_directory in .github .gitea .forgejo; do
    [[ -d $provider_directory ]] || continue
    replace_literal ', "client/v*"' '' "$provider_directory/workflows/release.yaml"
    remove_block "$provider_directory/workflows/release.yaml" '  client:' '            images.x86_64-linux.client.armv7l-unknown-linux-gnueabihf'
    remove_block "$provider_directory/workflows/vulnerable.yaml" '  rust:' '        run: cargo-audit audit'
    remove_block "$provider_directory/labeler.yaml" 'python:'
    remove_block "$provider_directory/labeler.yaml" 'rust:'
    cargo_rule_line=$(grep -n -m 1 -F '"matchManagers": ["cargo"],' "$provider_directory/renovate.jsonc" | cut -d: -f1 || true)
    if [[ -z $cargo_rule_line || $(sed -n "$((cargo_rule_line - 1))p;$((cargo_rule_line + 2))p" "$provider_directory/renovate.jsonc") != $'    {\n    },' ]]; then
      die "Unable to find the cargo rule in $provider_directory/renovate.jsonc."
    fi
    sed_inplace "$((cargo_rule_line - 1)),$((cargo_rule_line + 2))d" "$provider_directory/renovate.jsonc"
  done
  if [[ -d .zed ]]; then
    sed_inplace '/^    "Python": {$/,/^    },$/d' .zed/settings.json
  fi
  if [[ -d .vscode ]]; then
    remove_block .vscode/settings.json '  // charliermarsh.ruff'
    sed_inplace -e '/^    "ms-pyright\.pyright",$/d' -e '/^    "charliermarsh\.ruff",$/d' .vscode/extensions.json
  fi
}

# Rewrites the license holder without assuming the template's current year or name.
write_license() {
  local replacement
  if [[ -f LICENSE ]]; then
    replacement=$(escape_replacement "Copyright (c) $year $git_name")
    sed_inplace -E "s|^Copyright \\(c\\) [0-9]{4} .*$|$replacement|" LICENSE
  fi
}

# Replaces the README's title, badges and description, retaining only usage and
# contributing documentation, and points the components' READMEs at the
# project's container images.
write_readme() {
  local raw_url check_badge vulnerable_badge encoded_lock_url go_badge solid_badge
  local readme_badges readme_sections readme_image_base image_path
  if $origin_is_github; then
    raw_url="https://raw.githubusercontent.com/${github_repo_path}/refs/heads/main"
    check_badge="[![check](${origin_web_url}/actions/workflows/check.yaml/badge.svg?branch=main)](${origin_web_url}/actions/workflows/check.yaml)"
    vulnerable_badge="[![vulnerable](${origin_web_url}/actions/workflows/vulnerable.yaml/badge.svg?branch=main)](${origin_web_url}/actions/workflows/vulnerable.yaml)"
  else
    raw_url="${origin_web_url}/raw/branch/main"
    check_badge="[![check](${origin_web_url}/actions/workflows/check.yaml/badge.svg?branch=main&logo=${forge}&logoColor=%23bac2de&label=check&labelColor=%23313244)](${origin_web_url}/actions?workflow=check.yaml)"
    vulnerable_badge="[![vulnerable](${origin_web_url}/actions/workflows/vulnerable.yaml/badge.svg?branch=main&logo=${forge}&logoColor=%23bac2de&label=vulnerable&labelColor=%23313244)](${origin_web_url}/actions?workflow=vulnerable.yaml)"
  fi
  # The raw lockfile URL is nested inside the Shields endpoint URL.
  encoded_lock_url="${raw_url}/flake.lock"
  encoded_lock_url=${encoded_lock_url//:/%253A}
  encoded_lock_url=${encoded_lock_url//\//%252F}
  nixpkgs_badge="[![nixpkgs](https://shields.trev.zip/endpoint?url=https%3A%2F%2Fnix-shield.trev.zip%2Fbadge%3Furl%3D${encoded_lock_url}%26input%3Dnixpkgs&logoColor=%23bac2de&labelColor=%23313244&color=%235277C3)](https://nixos.org/)"
  go_badge="[![go](<https://shields.trev.zip/badge/dynamic/regex?url=${raw_url}/server/go.mod&search=toolchain%20go(.*)&replace=%241&style=flat&logo=go&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%2300ADD8>)](https://go.dev/doc/devel/release)"
  solid_badge="[![solidjs](https://shields.trev.zip/badge/dynamic/json?url=${raw_url}/web/package.json&query=%24.dependencies.solid-js&logo=solid&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%232C4F7C)](https://www.solidjs.com/)"
  rust_badge="[![rust](https://shields.trev.zip/badge/dynamic/toml?url=${raw_url}/client/Cargo.toml&query=%24.package.rust-version&logo=rust&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23D34516)](https://releases.rs/)"

  readme_sections=$(sed -n '/^## using$/,$p' README.md)
  readme_badges=("$check_badge" "$vulnerable_badge" "$nixpkgs_badge" "$go_badge" "$solid_badge")
  if $keep_client; then
    readme_badges+=("$rust_badge")
  else
    readme_sections=$(printf '%s\n' "$readme_sections" | sed '/^- \[client\](client\/README\.md)/d')
  fi
  # Container images share a registry path, so rewrite the prefix in every component README.
  readme_image_base=$(sed -n 's|^[[:space:]]*\([^[:space:]]*\)/server:latest$|\1|p' server/README.md)
  [[ -n $readme_image_base ]] || die 'Unable to find the server README image.'
  image_path=$(printf '%s' "$origin_repo_path" | tr '[:upper:]' '[:lower:]')
  if $origin_is_github; then
    image_base="ghcr.io/${image_path}"
  else
    image_base="${origin_host}/${image_path}"
  fi
  image="${image_base}/server:latest"
  replace_literal "$readme_image_base/" "$image_base/" server/README.md client/README.md

  readme_heading="# <img src=\"web/public/icons/icon.svg\" alt=\"\" height=\"32\"> $title"
  {
    printf '%s\n\n' "$readme_heading"
    printf '%s\n' "${readme_badges[@]}"
    printf '\n%s\n\n%s\n' "$description" "$readme_sections"
  } >README.md
  if ! $origin_is_github && [[ -f web/layout/layout.tsx ]]; then
    replace_literal 'on GitHub' "on $(forge_name "$forge")" web/layout/layout.tsx
  fi
}

# Commits the project as the new repository's root commit, preserving every
# remote and adding those of the other forges the project uses.
commit_new_repository() {
  local remote_name remote_url fetch_urls push_urls index
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
  for ((index = 0; index < ${#mirror_names[@]}; index++)); do
    remote_name=${mirror_names[index]}
    if "${new_git[@]}" remote get-url "$remote_name" >/dev/null 2>&1; then
      "${new_git[@]}" remote set-url "$remote_name" "${mirror_remotes[index]}"
    else
      "${new_git[@]}" remote add "$remote_name" "${mirror_remotes[index]}"
    fi
  done

  "${new_git[@]}" add -A
  "${new_git[@]}" rm --cached --ignore-unmatch --quiet -- init.sh
  git_as_user "${new_git[@]}" -c commit.gpgsign=false -c core.hooksPath=/dev/null commit --quiet --message 'chore: initialize project'
}

# Replaces .git with the new repository and removes the initializer.
activate_new_repository() {
  swap_started=true
  mv "$root/.git" "$old_git_dir"
  rm "$script_path" || die 'Unable to remove init.sh.'
  mv "$new_git_dir" "$root/.git" || die 'Unable to activate the new Git repository.'
}

# Generates and formats code, and adds it to the root commit. This waits until
# the new repository is active because Nix only sees tracked files.
generate_code() {
  nix run .#configure
  nix fmt
  git add -A
  git rm --cached --ignore-unmatch --quiet -- init.sh
  git_as_user git -c commit.gpgsign=false -c core.hooksPath=/dev/null commit --quiet --amend --no-edit
}

# Checks that the replacements took effect.
validate_project() {
  local file
  require_text README.md "$readme_heading"
  require_text README.md "$description"
  require_text README.md '## using'
  require_text README.md '[CONTRIBUTING.md](CONTRIBUTING.md)'
  require_text server/README.md "  $image"
  require_text README.md "$nixpkgs_badge"
  require_text flake.nix "description = \"$nix_description\";"
  if $keep_client; then
    require_text README.md "$rust_badge"
    require_text client/README.md "docker run ${image_base}/client:latest"
    require_text client/Cargo.toml "name = \"${slug}-client\""
    require_text client/Cargo.toml "version = \"$version\""
    require_text client/Cargo.toml "description = \"$json_description\""
    require_text client/Cargo.lock "name = \"${slug}-client\""
    require_text client/Cargo.lock "version = \"$version\""
    require_text client/default.nix "version = \"$version\";"
    require_text client/default.nix "description = \"$nix_description\";"
  else
    [[ ! -e client && ! -e test && ! -e pyproject.toml ]] || die 'Replacement validation failed: client files still exist.'
    ! grep -qE '\./client|client/|#client|\<cargo\>|rust(fmt|c|-analyzer)|\<ruff\>|pyright|python3|protoc-gen-prost|trevrpc-rust|^### client$' flake.nix buf.gen.yaml treefmt.toml README.md CONTRIBUTING.md AGENTS.md || die 'Replacement validation failed: client references remain.'
  fi
  require_text web/default.nix "version = \"$version\";"
  require_text server/default.nix "version = \"$version\";"
  require_text server/default.nix "description = \"$nix_description\";"
  require_text LICENSE "Copyright (c) $year $git_name"
  [[ ! -e init.sh ]] || die 'Replacement validation failed: init.sh still exists.'
  [[ -z $(git ls-files --error-unmatch init.sh 2>/dev/null || true) ]] || die 'Replacement validation failed: init.sh is still tracked.'
  require_text server/go.mod "module $target_module"
  require_text web/package.json "\"name\": \"${slug}-web\""
  if [[ $source_base != "$target_base" ]]; then
    while IFS= read -r -d '' file; do
      [[ -f $file && ! -L $file ]] || continue
      # .template.json records the template URL on purpose.
      [[ $file != .template.json ]] || continue
      if grep -qF -- "$source_base" "$file"; then
        die "Replacement validation found stale repository path in $file."
      fi
    done < <(git ls-files -z)
  fi
}

main() {
  if [[ $# -ne 2 || -z $1 || -z $2 ]]; then
    die "Usage: ${0##*/} \"Title\" \"Description\""
  fi
  title=$1
  description=$2

  check_repository
  check_project
  check_worktree
  read_template_metadata

  delete_directories=()
  choose_providers
  choose_editors
  choose_components
  check_ignored_files
  resolve_repositories

  start_new_repository
  template_commit=$(find_template_commit)

  changes_started=true
  remove_template_files
  rename_project
  set_descriptions
  reset_versions
  configure_providers
  if ! $keep_client; then
    remove_client
  fi
  write_license
  write_readme
  commit_new_repository
  activate_new_repository
  generate_code
  validate_project

  swap_started=false
  rm -rf "$old_git_dir"
  trap - EXIT INT TERM HUP
  printf 'Initialized %s (%s) from %s with a new root commit.\n' "$title" "$version" "$origin_web_url"
}

# init.test.sh sources the script to test its functions.
if [[ ${BASH_SOURCE[0]} == "$0" ]]; then
  main "$@"
fi
