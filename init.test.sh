#!/usr/bin/env bash
# The tests often single-quote text containing $
# shellcheck disable=SC2016

# Tests init.sh. The init flake check runs it from a copy of the template.

set -euo pipefail

# shellcheck source=init.sh
source ./init.sh

template=$PWD
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
export HOME=$tmp/home XDG_CONFIG_HOME=$tmp/home/.config GIT_CONFIG_NOSYSTEM=1
mkdir "$HOME"

fail() {
  printf 'not ok: %s\n' "$1" >&2
  exit 1
}

assert_eq() {
  local actual=$1 expected=$2 name=$3
  if [[ $actual != "$expected" ]]; then
    fail "$name: expected '$expected', got '$actual'"
  fi
}

# Runs a command that must fail with the expected output.
assert_fails() {
  local expected=$1 output
  shift
  if output=$("$@" 2>&1); then
    fail "$* succeeded"
  fi
  assert_eq "$output" "$expected" "$*"
}

assert_contains() {
  local file=$1 text
  shift
  for text in "$@"; do
    grep -qF -- "$text" "$file" || fail "$file does not contain '$text'"
  done
}

assert_lacks() {
  local file=$1 text
  shift
  for text in "$@"; do
    if grep -qF -- "$text" "$file"; then
      fail "$file contains '$text'"
    fi
  done
}

assert_missing() {
  local path
  for path in "$@"; do
    if [[ -e $path ]]; then
      fail "$path exists"
    fi
  done
}

test_parse_remote() {
  local url expected
  while read -r url expected; do
    parse_remote 'Origin URL' "$url"
    assert_eq "$remote_host $remote_repo_path $remote_web_url" "$expected" "parse_remote $url"
  done <<'EOF'
https://github.com/octo/hello.git github.com octo/hello https://github.com/octo/hello
http://user@git.example.com:3000/octo/hello/ git.example.com:3000 octo/hello https://git.example.com:3000/octo/hello
git@github.com:octo/hello.git github.com octo/hello https://github.com/octo/hello
forgejo@git.example.com:group/sub/hello git.example.com group/sub/hello https://git.example.com/group/sub/hello
ssh://git@git.example.com:2222/group/sub/hello.git git.example.com group/sub/hello https://git.example.com/group/sub/hello
EOF

  assert_fails 'Origin URL cannot contain a query string or fragment.' parse_remote 'Origin URL' 'https://github.com/octo/hello?tab=readme'
  assert_fails 'Repository URL must use HTTPS, SCP, or ssh:// syntax.' parse_remote 'Repository URL' 'ftp://github.com/octo/hello'
  assert_fails 'Repository URL must use HTTPS, SCP, or ssh:// syntax.' parse_remote 'Repository URL' 'octo/hello'
  assert_fails 'Origin URL must identify an owner and repository.' parse_remote 'Origin URL' 'git@github.com:hello.git'
  assert_fails 'Origin URL contains unsupported host or path characters.' parse_remote 'Origin URL' 'https://github.com/octo/hello world'
  assert_fails 'Origin URL contains unsupported host or path characters.' parse_remote 'Origin URL' 'https://github.com/octo/../hello'
}

test_slugify() {
  assert_eq "$(slugify 'Hello World')" hello-world 'slugify Hello World'
  assert_eq "$(slugify ' My_App 2! ')" my-app-2 'slugify My_App 2!'
  assert_fails 'Title must produce a package name beginning with a letter.' slugify '2048'
  assert_fails 'Title must produce a package name beginning with a letter.' slugify '!!!'
}

test_nix_string() {
  assert_eq "$(nix_string 'Say "hi" to ${USER} \o/')" 'Say \"hi\" to \${USER} \\o/' nix_string
}

test_json_string() {
  assert_eq "$(json_string 'Say "hi" to ${USER} \o/')" 'Say \"hi\" to ${USER} \\o/' json_string
}

test_replace_literal() (
  cd "$tmp"
  printf 'a.b*c [1] | $x & \\ a.b*c' >file
  ln -s file link
  replace_literal 'a.b*c' 'A|B&\C' file link missing
  assert_eq "$(<file)" 'A|B&\C [1] | $x & \ A|B&\C' replace_literal
  # The symlink must stay a symlink to the same file
  assert_eq "$(readlink link)" file 'replace_literal symlink'
  rm file link
)

test_remove_block() (
  cd "$tmp"
  printf 'a\n\nb\nstart\nc\n\nd\n' >file
  remove_block file start
  assert_eq "$(<file)" $'a\n\nb\nd' 'remove_block to blank line'
  printf 'a\nstart\nb\nend\n\nc\n' >file
  remove_block file start end
  assert_eq "$(<file)" $'a\nc' 'remove_block to end'
  assert_fails 'Unable to find the "start" block in file.' remove_block file start
  rm file
)

test_remove_checks() (
  cd "$tmp"
  cat >flake.nix <<'EOF'
        checks = {
          one = {
            env = {
              A = "a";
            };
          };

          two = {
            script = "";
          };

          three = {
            script = "";
          };
        };
EOF
  remove_checks two missing one
  assert_eq "$(<flake.nix)" $'        checks = {\n          three = {\n            script = "";\n          };\n        };' remove_checks
  rm flake.nix
)

# Creates the template's upstream repository, which init.sh fetches to find the
# template commit.
create_upstream() {
  git init --quiet "$tmp/upstream"
  cp -a "$template/." "$tmp/upstream"
  git -C "$tmp/upstream" add -A
  git -C "$tmp/upstream" commit --quiet --message 'feat: template'
  git config --global url."$tmp/upstream".insteadOf "$template_url"
  upstream_commit=$(git -C "$tmp/upstream" rev-parse HEAD)
}

# Copies the template into a new repository with an origin remote.
new_project() {
  local dir=$1 origin=$2
  git init --quiet "$dir"
  cp -a "$template/." "$dir"
  git -C "$dir" remote add origin "$origin"
}

# Checks what init.sh does for every project in the current directory.
assert_initialized() {
  local module=$1 slug=$2 description=$3
  assert_missing init.sh init.test.sh template.schema.json
  assert_eq "$(git log --format='%an <%ae> %s')" 'Test User <test@example.com> chore: initialize project' 'commits'
  assert_eq "$(git status --porcelain --untracked-files=all)" '' 'worktree status'
  assert_eq "$(jq -r .commit .template.json)" "$upstream_commit" '.template.json commit'
  # The generated code is replaced by buf generate, which the build sandbox can't run
  assert_eq "$(git grep -n -i -e trevstack -e template/stack -e init.sh -e init.test.sh -- ':!.template.json' ':!AGENTS.md' || true)" '' 'template references'
  assert_contains server/go.mod "module $module/server"
  assert_contains server/default.nix "pname = \"$slug-server\";" 'version = "0.0.1";'
  assert_contains web/default.nix "pname = \"$slug-web\";" 'version = "0.0.1";'
  assert_contains web/package.json "\"name\": \"$slug-web\"," '"version": "0.0.1",' "\"description\": \"$description\","
  assert_contains LICENSE "Copyright (c) $(date +%Y) Test User"
  assert_lacks flake.nix 'nix flake init' 'init = {'
  assert_lacks .gitattributes init
}

test_application() (
  local project=$tmp/hello-world description='Says "hi" to ${USER} \o/'
  new_project "$project" git@github.com:octo/hello-world.git
  cd "$project"
  # "use this template" creates a new commit with the template's tree
  git add -A
  git commit --quiet --message 'Initial commit'
  # The build sandbox has no /usr/bin/env for the shebang
  printf '%s\n' n https://git.example.com/octo/hello-world '' '' '' | bash ./init.sh 'Hello World' "$description"

  assert_initialized github.com/octo/hello-world hello-world 'Says \"hi\" to ${USER} \\o/'
  assert_eq "$(git remote)" $'forgejo\norigin' remotes
  assert_eq "$(git remote get-url forgejo)" https://git.example.com/octo/hello-world 'forgejo remote'
  assert_eq "$(head -n 1 README.md)" '# <img src="web/public/icons/icon.svg" alt="" height="32"> Hello World' 'README title'
  assert_contains README.md "$description" \
    'https://github.com/octo/hello-world/actions/workflows/check.yaml' \
    'refs/heads/main/client/Cargo.toml'
  assert_contains server/README.md 'ghcr.io/octo/hello-world/server:latest' 'nix run git+https://github.com/octo/hello-world#server'
  assert_contains client/README.md 'docker run ghcr.io/octo/hello-world/client:latest'
  assert_contains flake.nix \
    'description = "Says \"hi\" to \${USER} \\o/";' \
    'client-lint = {' 'integration = {' 'python-lint = {' \
    'actions-gh = {' 'actions-fj = {' 'renovate-gh = {' 'renovate-fj = {'
  assert_contains client/Cargo.toml 'name = "hello-world-client"' 'description = "Says \"hi\" to ${USER} \\o/"'
  assert_contains client/Cargo.lock 'name = "hello-world-client"'
  assert_contains client/src/main.rs 'use hello_world_client::'
  assert_contains .github/renovate.jsonc '"repositories": ["octo/hello-world"]'
  assert_contains .forgejo/renovate.jsonc \
    '"endpoint": "https://git.example.com/api/v1"' \
    '"repositories": ["octo/hello-world"]'
  [[ -d .zed && -d .vscode && -d .github && -d .forgejo && -d client && -d test ]] || fail 'application files are missing'
)

test_minimal() (
  local project=$tmp/hello-min
  new_project "$project" https://git.example.com/octo/hello-min.git
  cd "$project"
  # nix flake init leaves the template uncommitted and init.sh not executable
  chmod -x init.sh
  printf '%s\n' y n n n | bash init.sh 'Hello Min' 'Says hello'

  assert_initialized git.example.com/octo/hello-min hello-min 'Says hello'
  assert_eq "$(git log --format=%s)" 'chore: initialize project' commits
  assert_eq "$(git remote)" origin remotes
  assert_eq "$(git remote get-url origin)" https://git.example.com/octo/hello-min.git 'origin remote'
  assert_missing .github .zed .vscode client test pyproject.toml
  assert_contains README.md \
    'https://git.example.com/octo/hello-min/actions?workflow=check.yaml' \
    '- [server](server/README.md)'
  assert_lacks README.md 'client' 'rust'
  assert_contains server/README.md 'git.example.com/octo/hello-min/server:latest'
  assert_contains .forgejo/renovate.jsonc \
    '"endpoint": "https://git.example.com/api/v1"' \
    '"repositories": ["octo/hello-min"]'
  assert_lacks flake.nix \
    client cargo python ruff trevrpc-rust \
    'client-lint = {' 'integration = {' 'python-lint = {' 'actions-gh = {' 'renovate-gh = {'
  assert_contains flake.nix 'actions-fj = {' 'renovate-fj = {'
  assert_lacks .gitattributes Cargo client
  assert_lacks buf.gen.yaml prost trevrpc-rust
  assert_lacks AGENTS.md client
)

test_rollback() (
  local project=$tmp/rollback head mode
  new_project "$project" git@github.com:octo/rollback.git
  cd "$project"
  git add -A
  git commit --quiet --message 'Initial commit'
  head=$(git rev-parse HEAD)
  mode=$(stat -c %a init.sh)

  touch untracked
  assert_fails 'Commit or discard all existing worktree changes before running the initializer.' \
    bash ./init.sh Rollback Rollback </dev/null
  rm untracked

  # The Forgejo URL is read before the template is changed
  assert_fails 'Repository URL must use HTTPS, SCP, or ssh:// syntax.' \
    bash ./init.sh Rollback Rollback < <(printf '%s\n' n ftp://git.example.com/octo/rollback)

  # Generating code fails after the new repository has replaced the old one
  mkdir "$tmp/failing"
  printf '#!%s\nexit 1\n' "$BASH" >"$tmp/failing/nix"
  chmod +x "$tmp/failing/nix"
  assert_fails '' env PATH="$tmp/failing:$PATH" \
    bash ./init.sh Rollback Rollback < <(printf '%s\n' n https://git.example.com/octo/rollback '' '' '')

  assert_eq "$(git rev-parse HEAD)" "$head" HEAD
  assert_eq "$(git status --porcelain --untracked-files=all)" '' 'worktree status'
  assert_eq "$(git remote)" origin remotes
  assert_eq "$(stat -c %a init.sh)" "$mode" 'init.sh mode'
  assert_missing "$tmp"/.rollback.git-* .template.json
)

git config --global user.name 'Test User'
git config --global user.email test@example.com
git config --global init.defaultBranch main

# The build sandbox has no network or flake inputs, so stand in for the nix
# commands that init.sh runs. nix fmt runs its formatter. nix run .#configure
# runs buf generate, npm install and go mod tidy, which download dependencies,
# so it only updates the module path in the generated Go code.
template_base=${template_url#https://}
mkdir "$tmp/bin"
cat >"$tmp/bin/nix" <<EOF
#!$BASH
case \$* in
  fmt)
    TREEFMT_TREE_ROOT=\$PWD exec treefmt
    ;;
  'run .#configure')
    base=\$(sed -n 's|^module \(.*\)/server\$|\1|p' server/go.mod)
    grep -rlF "$template_base" server/gen | xargs sed -i "s|$template_base|\$base|g"
    ;;
  *)
    exit 1
    ;;
esac
EOF
chmod +x "$tmp/bin/nix"
PATH=$tmp/bin:$PATH

create_upstream
for test in test_parse_remote test_slugify test_nix_string test_json_string test_replace_literal test_remove_block test_remove_checks test_application test_minimal test_rollback; do
  "$test"
  printf 'ok: %s\n' "$test"
done
