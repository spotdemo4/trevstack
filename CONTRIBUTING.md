# contributing

## requirements

- [nix](https://nixos.org/)

## getting started

```sh
nix develop
```

with [direnv](https://direnv.net/):

```sh
ln -s .envrc.project .envrc
direnv allow
```

run project configuration with:

```sh
nix run .#configure
```

### run

run the server:

```sh
nix run .#server
```

run the client with optional arguments:

```sh
nix run .#client -- alice 10
```

### format

```sh
nix fmt
```

### check

```sh
nix flake check
```

### build

```sh
nix build .#server
nix build .#client
```

### release

with [bumper](https://trev.zip/llc/bumper):

```sh
bumper
```

releases are automatically created for [significant](https://www.conventionalcommits.org/en/v1.0.0/#summary) changes.
