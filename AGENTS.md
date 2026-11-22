# AGENTS.md

## Task Completion Requirements

- `nix fmt` must be used for formatting.
- `nix flake check` must pass before considering tasks completed. Intent-to-add (`git add -N`) new files so `nix flake check` can see them.

## Running and Debugging the Server

- Start the server from the repository root with `nix run .#server`, using the Bash tool's `run_in_background` option, since it runs until stopped. It is ready when it logs `"msg":"starting"`, and serves the web app and API at `https://localhost:8080`. Pass flags after `--`, such as `nix run .#server -- --port 8081` when 8080 is taken.
- It is the production build with the web app embedded, so stop it and run `nix run .#server` again to pick up server or web changes. Like `nix flake check`, it only sees files tracked by Git.
- Without `TLS_CERT` and `TLS_KEY`, the server uses a self-signed development certificate, so in Chrome DevTools, navigating to `https://localhost:8080` fails with `net::ERR_CERT_AUTHORITY_INVALID` and leaves the tab on a "Your connection is not private" page. Ignore the warning: take a snapshot, click "Advanced", then "Proceed to localhost (unsafe)". The app then loads, and its API calls work because WebTransport trusts the certificate by its hashes. Chrome remembers this choice until the browser restarts.
- The server also writes the development certificate to `~/.config/trevstack/development.pem`, which the client trusts by default, so `nix run .#client -- watch` connects to it without `--ca-cert`.
- The database is the user's own, at `~/.config/trevstack/trevstack.db`, and `Signup` is limited to 3 requests per hour for each IP, so reuse a test account rather than signing up again.

## Generated API Code

- Regenerate TrevRPC libraries with `buf generate` from the repository root (or `nix develop -c buf generate` outside the dev shell); do not manually edit generated files in `server/gen/`, `web/gen/`, `client/src/gen/`, or `android/app/src/gen/`.
- Make API changes in the proto source files.

## Version Control Requirements

- Commit messages created by agents must follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).
- Commits created by agents must include an `Assisted-by: [tool name] ([primary model name and version])` Git commit trailer.
- Branch names created by agents must follow [Conventional Branch 1.0.0](https://conventionalbranch.org/v1.0.0/).
- In pull request descriptions and commit messages, put a closing keyword (`close`, `closes`, `closed`, `fix`, `fixes`, `fixed`, `resolve`, `resolves`, `resolved`) right before an issue reference only when the change should close that issue, and a reopening keyword (`reopen`, `reopens`, `reopened`) only when it should reopen it. Forgejo, Gitea and GitHub act on a closing keyword on merge wherever it appears, even in a negated sentence such as "this does not close #123", and Forgejo and Gitea act on a reopening keyword the same way. Refer to an issue that stays as it is as "Refs #123" or "#123 stays open".
