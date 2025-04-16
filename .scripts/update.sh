#!/usr/bin/env bash

git_root=$(git rev-parse --show-toplevel)

echo "updating client"
cd "${git_root}/client"
npm update --save
if ! git diff --exit-code package.json package-lock.json; then
    git add package-lock.json
    git add package.json
    git commit -m "build(client): updated npm dependencies"
fi

echo "updating server"
cd "${git_root}/server"
go get -u
go mod tidy
if ! git diff --exit-code go.mod go.sum; then
    git add go.mod
    git add go.sum
    git commit -m "build(go): updated go dependencies"
fi

echo "updating nix"
cd "${git_root}"
nix-update --flake --version=skip default
if ! git diff --exit-code flake.nix; then
    git add flake.nix
    git commit -m "build(nix): updated nix hashes"
fi