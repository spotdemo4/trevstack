# <img src="web/public/icons/icon.svg" alt="" height="32"> TrevStack

[![check](https://trev.zip/template/stack/actions/workflows/check.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=check&labelColor=%23313244)](https://trev.zip/template/stack/actions?workflow=check.yaml)
[![vulnerable](https://trev.zip/template/stack/actions/workflows/vulnerable.yaml/badge.svg?branch=main&logo=forgejo&logoColor=%23bac2de&label=vulnerable&labelColor=%23313244)](https://trev.zip/template/stack/actions?workflow=vulnerable.yaml)
[![nixpkgs](https://shields.trev.zip/endpoint?url=https%3A%2F%2Fnix-shield.trev.zip%2Fbadge%3Furl%3Dhttps%253A%252F%252Ftrev.zip%252Ftemplate%252Fstack%252Fraw%252Fbranch%252Fmain%252Fflake.lock%26input%3Dnixpkgs&logoColor=%23bac2de&labelColor=%23313244&color=%235277C3)](https://nixos.org/)
[![go](<https://shields.trev.zip/badge/dynamic/regex?url=https%3A%2F%2Ftrev.zip%2Ftemplate%2Fstack%2Fraw%2Fbranch%2Fmain%2Fserver%2Fgo.mod&search=toolchain%20go(.*)&replace=%241&style=flat&logo=go&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%2300ADD8>)](https://go.dev/doc/devel/release)
[![solidjs](https://shields.trev.zip/badge/dynamic/json?url=https://trev.zip/template/stack/raw/branch/main/web/package.json&query=%24.dependencies.solid-js&logo=solid&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%232C4F7C)](https://www.solidjs.com/)
[![rust](https://shields.trev.zip/badge/dynamic/toml?url=https://trev.zip/template/stack/raw/branch/main/client/Cargo.toml&query=%24.package.rust-version&logo=rust&logoColor=%23bac2de&label=version&labelColor=%23313244&color=%23D34516)](https://releases.rs/)

TrevStack is a template for full-stack applications built with Go, Rust, SolidJS, and [TrevRPC](https://trev.zip/llc/TrevRPC).
A new project starts out as a working app: a server, a web app, and a command-line client that share
one typed API. The build, test, CI, and release setup around them comes included, all driven by
[Nix](https://nixos.org/).

The example app shares the cursors of every signed-in user. Each page of the web app and each
subcommand of the client shows a different kind of RPC: unary, server streaming, client streaming,
and bidirectional streaming, so there is a working example of each to build from.

part of [trev.zip/template](https://trev.zip/template)

## features

- **api**: Protobuf definitions in [`proto/`](proto) are the source of truth. `buf generate` turns
  them into typed Go, TypeScript, and Rust code, and the server and web app both check requests
  against their [`buf.validate`](https://buf.build/bufbuild/protovalidate) rules.
- **[server](server/README.md)**: a single Go binary that serves the API and the embedded web app
  over QUIC, and falls back to TCP when UDP is blocked. It comes with an embedded SQLite database
  and its migrations, session authentication, rate limiting, OpenTelemetry traces, metrics, and logs,
  and a self-signed development certificate, so it runs with no setup.
- **[web](web/README.md)**: a SolidJS single-page app built with Vite and styled with Tailwind CSS,
  installable as a progressive web app. It calls the API over WebTransport, wraps calls in
  [Effect](https://effect.website/) for typed errors and cancellation, and validates forms against the
  same rules as the server. Light and dark themes use [Catppuccin](https://catppuccin.com/) colors.
- **[client](client/README.md)**: a Rust command-line client that turns the terminal into a live
  cursor pad, and reads and writes plain lines when scripted.
- **nix**: a flake with the development shell, reproducible builds, cross-compiled binaries,
  container images, and `nix fmt`. `nix flake check` runs every build, linter, and test, including
  Playwright integration tests that drive the real server, client, and web app.
- **automation**: workflows for GitHub, Gitea, and Forgejo. They check every push and pull request,
  label pull requests, update dependencies with [Renovate](https://docs.renovatebot.com/), scan for
  vulnerabilities each week, bump versions with [bumper](https://trev.zip/llc/bumper), and publish
  releases and container images.
- **editors**: settings and recommended extensions for [Zed](https://zed.dev/) and
  [VS Code](https://code.visualstudio.com/).

## getting started

You need [nix](https://nixos.org/), and Git with `user.name` and `user.email` set. To start a new
project, create a repository from this template, clone it, and from its root run:

```sh
./init.sh "Title" "Description"
```

or, in a new git repository with an `origin` remote, run:

```sh
nix flake init -t git+https://trev.zip/template/stack
bash init.sh "Title" "Description"
```

`init.sh` reads the project's location from the `origin` remote, then asks:

- which forge hosts `origin`, GitHub, Gitea, or Forgejo, and whether the project also has a
  repository on each of the others
- whether you use Zed and VS Code
- whether to keep the Rust client

It then:

- renames the project after its title everywhere, from the Go module to the package names
- sets the description and resets every version to `0.0.1`
- removes the configuration of forges and editors you don't use, and the client if you don't keep it
- regenerates the API code
- replaces the Git history with one `chore: initialize project` commit

It also records the template commit the project started from in `.template.json`, so later template
changes can be ported to the project.

## using

- [server](server/README.md): serves the API and the web app; running, configuration, TLS,
  authentication, and observability
  - [web](web/README.md): web app served by the server; pages, API calls, and theme
- [client](client/README.md): command-line client for the API; subcommands, options, and scripting

The API is defined in [`proto/`](proto); change it there and run `buf generate` to regenerate the code
that uses it.

## contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development requirements and instructions.
