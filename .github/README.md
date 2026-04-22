# TrevStack

[![check](https://img.shields.io/github/actions/workflow/status/spotdemo4/trevstack/check.yaml?branch=main&logo=github&logoColor=%23bac2de&label=check&labelColor=%23313244)](https://github.com/spotdemo4/trevstack/actions/workflows/check.yaml/)
[![vulnerable](https://img.shields.io/github/actions/workflow/status/spotdemo4/trevstack/vulnerable.yaml?branch=main&logo=github&logoColor=%23bac2de&label=vulnerable&labelColor=%23313244)](https://github.com/spotdemo4/trevstack/actions/workflows/vulnerable.yaml)
[![nix](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fraw.githubusercontent.com%2Fspotdemo4%2Ftrevstack%2Frefs%2Fheads%2Fmain%2Fflake.lock&query=%24.nodes.nixpkgs.original.ref&logo=nixos&logoColor=%23bac2de&label=channel&labelColor=%23313244&color=%234d6fb7)](https://nixos.org/)
[![go](https://img.shields.io/github/go-mod/go-version/spotdemo4/trevstack?filename=server%2Fgo.mod&logo=go&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%2300ADD8)](https://go.dev/doc/devel/release)
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
nix run
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
bumper "web" "openapi.yaml" ".github/README.md"
```

releases are automatically created for [significant](https://www.conventionalcommits.org/en/v1.0.0/#summary) changes

## use

### download

| OS      | Architecture | Download                                                                                                                               |
| ------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| Linux   | amd64        | [trevstack_0.0.3_linux_amd64](https://github.com/spotdemo4/trevstack/releases/download/v0.0.3/trevstack_0.0.3_linux_amd64)             |
| Linux   | arm64        | [trevstack_0.0.3_linux_arm64](https://github.com/spotdemo4/trevstack/releases/download/v0.0.3/trevstack_0.0.3_linux_arm64)             |
| Linux   | arm          | [trevstack_0.0.3_linux_arm](https://github.com/spotdemo4/trevstack/releases/download/v0.0.3/trevstack_0.0.3_linux_arm)                 |
| MacOS   | arm64        | [trevstack_0.0.3_darwin_arm64](https://github.com/spotdemo4/trevstack/releases/download/v0.0.3/trevstack_0.0.3_darwin_arm64)           |
| Windows | amd64        | [trevstack_0.0.3_windows_amd64.exe](https://github.com/spotdemo4/trevstack/releases/download/v0.0.3/trevstack_0.0.3_windows_amd64.exe) |

### docker

```elm
docker run -P ghcr.io/spotdemo4/trevstack:0.0.3
```

### nix

```elm
nix run github:spotdemo4/trevstack
```
