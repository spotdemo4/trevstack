#!/usr/bin/env bash

git_root=$(git rev-parse --show-toplevel)
updated=false

echo "updating nix flake"
cd "${git_root}"
nix flake update --accept-flake-config
if ! git diff --exit-code flake.lock; then
    git add flake.lock
    git commit -m "build(nix): updated nix dependencies"
fi

echo "updating protobuf deps"
cd "${git_root}/proto"
buf dep update
if ! git diff --exit-code buf.lock; then
    git add buf.lock
    git commit -m "build(buf): updated buf dependencies"
fi

echo "updating client"
cd "${git_root}/client"
npm update --save && npm i
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
    git commit -m "build(server): updated go dependencies"
    updated=true
fi

if [ "${updated}" = true ]; then
    echo "updating nix hashes"
    cd "${git_root}"
    nix-update --flake --version=skip --subpackage client default
    git add flake.nix
    git commit -m "build(nix): updated nix hashes"
fi