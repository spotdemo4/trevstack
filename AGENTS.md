# AGENTS.md

## Task Completion Requirements

- `nix fmt` must be used for formatting.
- `nix flake check` must pass before considering tasks completed. Intent-to-add (`git add -N`) new files so `nix flake check` can see them.

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
