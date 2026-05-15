# TrevStack

[![check](https://img.shields.io/github/actions/workflow/status/spotdemo4/trevstack/check.yaml?branch=main&logo=github&logoColor=%23bac2de&label=check&labelColor=%23313244)](https://github.com/spotdemo4/trevstack/actions/workflows/check.yaml)
[![vulnerable](https://img.shields.io/github/actions/workflow/status/spotdemo4/trevstack/vulnerable.yaml?branch=main&logo=github&logoColor=%23bac2de&label=vulnerable&labelColor=%23313244)](https://github.com/spotdemo4/trevstack/actions/workflows/vulnerable.yaml)
[![go](https://img.shields.io/github/go-mod/go-version/spotdemo4/trevstack?filename=server%2Fgo.mod&logo=go&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%2300ADD8)](https://go.dev/doc/devel/release)
[![node](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fraw.githubusercontent.com%2Fspotdemo4%2Ftrevstack%2Frefs%2Fheads%2Fmain%2Fweb%2Fpackage.json&query=%24.engines.node&logo=nodedotjs&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23339933)](https://nodejs.org/en/about/previous-releases)
[![solidjs](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fraw.githubusercontent.com%2Fspotdemo4%2Ftrevstack%2Frefs%2Fheads%2Fmain%2Fweb%2Fpackage.json&query=%24.dependencies.solid-js&logo=solid&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%232C4F7C)](https://www.solidjs.com/)
[![flakehub](https://img.shields.io/endpoint?url=https://flakehub.com/f/spotdemo4/trevstack/badge&labelColor=%23313244)](https://flakehub.com/flake/spotdemo4/trevstack)

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
docker run -P ghcr.io/spotdemo4/trevstack:0.11.0
```

### nix

```elm
nix run github:spotdemo4/trevstack
```

### [download](https://github.com/spotdemo4/trevstack/releases/latest)
