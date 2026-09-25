# TrevStack

[![check](https://trev.zip/template/stack/actions/workflows/check.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=check&labelColor=%23313244)](https://trev.zip/template/stack/actions?workflow=check.yaml)
[![vulnerable](https://trev.zip/template/stack/actions/workflows/vulnerable.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=vulnerable&labelColor=%23313244)](https://trev.zip/template/stack/actions?workflow=vulnerable.yaml)
[![nixpkgs](https://img.shields.io/endpoint?url=https%3A%2F%2Fnix-shield.trev.zip%2Fbadge%3Furl%3Dhttps%253A%252F%252Ftrev.zip%252Ftemplate%252Fstack%252Fraw%252Fbranch%252Fmain%252Fflake.lock%26input%3Dnixpkgs&logoColor=%23bac2de&labelColor=%23313244&color=%235277C3)](https://nixos.org/)
[![go](<https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Ftrev.zip%2Ftemplate%2Fstack%2Fraw%2Fbranch%2Fmain%2Fserver%2Fgo.mod&search=toolchain%20go(.*)&replace=%241&style=flat&logo=go&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%2300ADD8>)](https://go.dev/doc/devel/release)
[![node](https://img.shields.io/badge/dynamic/json?url=https://trev.zip/template/stack/raw/branch/main/web/package.json&query=%24.engines.node&logo=nodedotjs&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23339933)](https://nodejs.org/en/about/previous-releases)
[![solidjs](https://img.shields.io/badge/dynamic/json?url=https://trev.zip/template/stack/raw/branch/main/web/package.json&query=%24.dependencies.solid-js&logo=solid&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%232C4F7C)](https://www.solidjs.com/)
[![rust](https://img.shields.io/badge/dynamic/toml?url=https://trev.zip/template/stack/raw/branch/main/client/Cargo.toml&query=%24.package.rust-version&logo=rust&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23D34516)](https://releases.rs/)

TrevStack is a template for full-stack applications built with Go, Rust, SolidJS, and ConnectRPC

to initialize a new project, run:

```sh
./init.sh "Title" "Description"
```

part of [spotdemo4/templates](https://github.com/spotdemo4/templates)

## using

### server

#### docker

```sh
docker run -P trev.zip/llc/stack/server:latest
```

#### nix

```sh
nix run git+https://trev.zip/template/stack#server
```

#### download

https://trev.zip/template/stack/releases

### client

#### nix

```sh
nix run git+https://trev.zip/template/stack#client -- alice 10
```

#### download

https://trev.zip/template/stack/releases

## contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development requirements and instructions.
