# AGENTS.md

## Task Completion Requirements

- `nix flake check` must pass before considering tasks completed.
- `nix fmt` must be used for formatting.

`nix flake check` reads from git's tree. Intent-to-add new files so `nix flake check` can see them.
