#!/usr/bin/env bash

git_root=$(git rev-parse --show-toplevel)

echo "linting client"
cd "${git_root}/client"
npx prettier --check .
npx eslint .
npx svelte-check

echo "linting server"
cd "${git_root}/server"
revive -config revive.toml -set_exit_status ./...
sqlfluff lint

echo "linting protobuf"
cd "${git_root}"
buf lint

echo "linting nix"
cd "${git_root}"
nix fmt -- flake.nix --check
nix flake check --all-systems