# AGENTS.md

## Task Completion Requirements

- `nix fmt` must be used for formatting.
- `nix flake check` must pass before considering tasks completed. Intent-to-add (`git add -N`) new files so `nix flake check` can see them.

## Running and Debugging the Server

- Start the server from the repository root with `nix run .#server`, using the Bash tool's `run_in_background` option, since it runs until stopped. It is ready when it logs `"msg":"starting"`, and serves the web app and API at `https://localhost:8080`. Pass flags after `--`, such as `nix run .#server -- --port 8081` when 8080 is taken.
- It is the production build with the web app embedded, so stop it and run `nix run .#server` again to pick up server or web changes. Like `nix flake check`, it only sees files tracked by Git.
- Without `TLS_CERT` and `TLS_KEY`, the server uses a self-signed development certificate, so in Chrome DevTools, navigating to `https://localhost:8080` fails with `net::ERR_CERT_AUTHORITY_INVALID` and leaves the tab on a "Your connection is not private" page. Ignore the warning: take a snapshot, click "Advanced", then "Proceed to localhost (unsafe)". The app then loads, and its API calls work because WebTransport trusts the certificate by its hashes. Chrome remembers this choice until the browser restarts.
- The database is the user's own, at `~/.config/trevstack/trevstack.db`, and `Signup` is limited to 3 requests per hour for each IP, so reuse a test account rather than signing up again.

## Generated API Code

- Regenerate TrevRPC libraries with `buf generate` from the repository root (or `nix develop -c buf generate` outside the dev shell); do not manually edit generated files in `server/gen/`, `web/gen/`, or `client/src/gen/`.
- Make API changes in the proto source files.

## Version Control Requirements

- Commit messages created by agents must follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).
- Commits created by agents must include an `Assisted-by: [tool name] ([primary model name and version])` Git commit trailer.
- Branch names created by agents must follow [Conventional Branch 1.0.0](https://conventionalbranch.org/v1.0.0/).

## Template Update Requirements

- If `.template.json` exists, this project was created from the template at `url`, and `commit` is the latest template commit whose changes have been applied.
- To update from the template, port each commit in `<commit>..main` of the template to this project, adapting names, URLs and paths. Skip dependency updates (`chore(deps)`) that this project's Renovate already handles, and changes to `init.sh`, `init.test.sh`, `template.schema.json`, the `templates` flake output and the `init` check, which are not part of this project.
- Set `commit` to the template commit that was synced to in the same commit that applies its changes.
