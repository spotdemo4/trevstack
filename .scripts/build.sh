#!/usr/bin/env bash

git_root=$(git rev-parse --show-toplevel)
url=$(git config --get remote.origin.url)
name=$(basename -s .git "${url}")
git_version=$(git describe --tags --abbrev=0)
version=${git_version#v}

echo "building client"
cd "${git_root}"
nix build .#trevstack-client
cp -a result/. server/client
chmod -R u+w server/client

echo "building server"
cd "${git_root}/server"
echo "Building ${name}-windows-amd64-${version}.exe"
GOOS=windows GOARCH=amd64 go build -o "./build/${name}-windows-amd64-${version}.exe" .
echo "Building ${name}-linux-amd64-${version}"
GOOS=linux GOARCH=amd64 go build -o "./build/${name}-linux-amd64-${version}" .
echo "Building ${name}-linux-amd64-${version}"
GOOS=linux GOARCH=arm64 go build -o "./build/${name}-linux-arm64-${version}" .
echo "Building ${name}-linux-arm-${version}"
GOOS=linux GOARCH=arm go build -o "./build/${name}-linux-arm-${version}" .