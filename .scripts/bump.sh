#!/usr/bin/env bash

git_root=$(git rev-parse --show-toplevel)
git_version=$(git describe --tags --abbrev=0)
version=${git_version#v}
next_version=$(echo "${version}" | awk -F. -v OFS=. '{$NF += 1 ; print}')

echo "bumping client"
cd "${git_root}/client"
npm version "${next_version}"
git add package-lock.json
git add package.json

echo "bumping nix"
cd "${git_root}"
nix-update --flake --version "${next_version}" default
git add flake.nix

git commit -m "bump: v${version} -> v${next_version}"
git push origin main
git tag -a "v${next_version}" -m "bump: v${version} -> v${next_version}"
git push origin "v${next_version}"