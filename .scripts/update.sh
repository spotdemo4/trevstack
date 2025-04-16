#!/usr/bin/env bash

git_root=$(git rev-parse --show-toplevel)
updated=false

echo "updating client"
cd "${git_root}/client"
npm update --save
if ! git diff --exit-code package.json package-lock.json; then
    git add package-lock.json
    git add package.json
    git commit -m "build(client): updated npm dependencies"
    updated=true
fi

echo "updating server"
cd "${git_root}/server"
go get -u
go mod tidy
if ! git diff --exit-code go.mod go.sum; then
    git add go.mod
    git add go.sum
    git commit -m "build(go): updated go dependencies"
    updated=true
fi

if [ "${updated}" = true ]; then
    echo "updating nix"
    cd "${git_root}"
    nix-update --flake --version=skip --subpackage trevstack-client trevstack
    git add flake.nix
    git commit -m "build(nix): updated nix hashes"
else
    echo "nothing to update"
fi