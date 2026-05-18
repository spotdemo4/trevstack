> [!NOTE]
> This repository is mirrored to GitHub for better visibility, but development happens on [trev.zip](https://trev.zip/trev/stack)

# TrevStack

[![check](https://trev.zip/trev/stack/actions/workflows/check.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=check&labelColor=%23313244)](https://trev.zip/trev/stack/actions?workflow=check.yaml)
[![vulnerable](https://trev.zip/trev/stack/actions/workflows/vulnerable.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=vulnerable&labelColor=%23313244)](https://trev.zip/trev/stack/actions?workflow=vulnerable.yaml)
[![go](<https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Ftrev.zip%2Ftrev%2Fstack%2Fraw%2Fbranch%2Fmain%2Fserver%2Fgo.mod&search=toolchain%20go(.*)&replace=%241&style=flat&logo=go&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%2300ADD8>)](https://go.dev/doc/devel/release)
[![node](https://img.shields.io/badge/dynamic/json?url=https://trev.zip/trev/stack/raw/branch/main/web/package.json&query=%24.engines.node&logo=nodedotjs&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23339933)](https://nodejs.org/en/about/previous-releases)
[![solidjs](https://img.shields.io/badge/dynamic/json?url=https://trev.zip/trev/stack/raw/branch/main/web/package.json&query=%24.dependencies.solid-js&logo=solid&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%232C4F7C)](https://www.solidjs.com/)

TrevStack is a template for full-stack applications built with Go, SolidJS, and ConnectRPC

part of [spotdemo4/templates](https://github.com/spotdemo4/templates)

## requirements

- [nix](https://nixos.org/)

## getting started

```elm
nix develop
```

### run

```elm
nix run #dev
```

### format

```elm
nix fmt
```

### check

```elm
nix flake check
```

### build

```elm
nix build
```

### release

```elm
bumper "openapi.yaml" "README.md" "web"
```

releases are automatically created for [significant](https://www.conventionalcommits.org/en/v1.0.0/#summary) changes

## use

### docker

```elm
docker run -P ghcr.io/spotdemo4/trevstack:0.10.1
```

### nix

```elm
nix run github:spotdemo4/trevstack
```

### download

https://trev.zip/trev/stack/releases
