# AGENTS.md

## Task Completion Requirements

- `nix fmt` must be used for formatting.
- `nix flake check` must pass before considering tasks completed. Intent-to-add (`git add -N`) new files so `nix flake check` can see them.

## Generated API Code and Documentation

- Regenerate Connect libraries and OpenAPI documentation with `buf generate` from the repository root (or `nix develop -c buf generate` outside the dev shell); do not manually edit generated files in `server/connect/`, `web/connect/`, `client/src/connect/`, or `docs/openapi.yaml`.
- Make API changes in the proto source files. Keep `docs/openapi.base.yaml` limited to shared OpenAPI metadata and configuration; do not duplicate generated paths, operations, or schemas there.

## Version Control Requirements

- Commit messages created by agents must follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).
- Commits created by agents must include an `Assisted-by: [tool name] ([primary model name and version])` Git commit trailer.
- Branch names created by agents must follow [Conventional Branch 1.1.0](https://conventionalbranch.org/).
