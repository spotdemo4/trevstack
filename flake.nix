{
  description = "A template for trevstack";

  nixConfig = {
    extra-substituters = [
      "https://trevstack.cachix.org"
    ];
    extra-trusted-public-keys = [
      "trevstack.cachix.org-1:wlY2/NBLC4U4u8fD4WgW1kMstfiGbGmgDwE3dBho4tE="
    ];
  };

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    treli.url = "github:spotdemo4/treli";
  };

  outputs = {
    nixpkgs,
    treli,
    ...
  }: let
    pname = "trevstack";
    version = "0.0.20";

    build-systems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
    host-systems = [
      {
        GOOS = "linux";
        GOARCH = "amd64";
      }
      {
        GOOS = "linux";
        GOARCH = "arm64";
      }
      {
        GOOS = "linux";
        GOARCH = "arm";
      }
      {
        GOOS = "windows";
        GOARCH = "amd64";
      }
      {
        GOOS = "darwin";
        GOARCH = "amd64";
      }
      {
        GOOS = "darwin";
        GOARCH = "arm64";
      }
    ];
    forSystem = f:
      nixpkgs.lib.genAttrs build-systems (
        system:
          f {
            inherit system;
            pkgs = import nixpkgs {
              inherit system;
            };
          }
      );
  in {
    devShells = forSystem ({pkgs, ...}: let
      protoc-gen-connect-openapi = pkgs.buildGoModule {
        name = "protoc-gen-connect-openapi";
        src = pkgs.fetchFromGitHub {
          owner = "sudorandom";
          repo = "protoc-gen-connect-openapi";
          rev = "v0.16.1";
          sha256 = "sha256-3XBQCc9H9N/AZm/8J5bJRgBhVtoZKFvbdTB+glHxYdA=";
        };
        vendorHash = "sha256-CIiG/XhV8xxjYY0sZcSvIFcJ1Wh8LyDDwqem2cSSwBA=";
      };
    in {
      default = pkgs.mkShell {
        packages = with pkgs; [
          git
          nix-update
          treli.packages."${system}".default

          # Server
          go
          gotools
          gopls
          revive
          sqlc

          # database
          sqlite
          dbmate
          sqlfluff

          # Protobuf
          buf
          protoc-gen-go
          protoc-gen-connect-go
          protoc-gen-es
          protoc-gen-connect-openapi

          # Client
          nodejs_22
        ];
      };
    });

    checks = forSystem ({pkgs, ...}: {
      buf = with pkgs;
        runCommandLocal "check-buf" {
          nativeBuildInputs = with pkgs; [
            buf
          ];
        } ''
          export HOME=$(pwd)
          cd ${./.}
          buf lint
          touch $out
        '';

      nix = with pkgs;
        runCommandLocal "check-nix" {
          nativeBuildInputs = with pkgs; [
            alejandra
          ];
        } ''
          cd ${./.}
          alejandra -c .
          touch $out
        '';

      client = with pkgs;
        buildNpmPackage {
          pname = "check-client";
          inherit version;
          src = ./client;
          npmDepsHash = "sha256-tOyEhHa/N0FSSVqxW9PrAY2VIXu24fXbAVylPgYjeYs=";
          dontNpmInstall = true;

          buildPhase = ''
            npx prettier --check .
            npx eslint .
            npx svelte-kit sync && npx svelte-check
            touch $out
          '';
        };

      server = with pkgs;
        runCommandLocal "check-server" {
          nativeBuildInputs = with pkgs; [
            revive
            sqlfluff
          ];
        } ''
          cd ${./server}
          revive -config revive.toml -set_exit_status ./...
          sqlfluff lint
          touch $out
        '';
    });

    apps = forSystem ({pkgs, ...}: {
      update = {
        type = "app";
        program = pkgs.lib.getExe (pkgs.writeShellApplication {
          name = "update";
          runtimeInputs = with pkgs; [
            git
            nix
            nodejs_22
            go
            nix-update
          ];
          text = builtins.readFile ./.scripts/update.sh;
        });
      };

      bump = {
        type = "app";
        program = pkgs.lib.getExe (pkgs.writeShellApplication {
          name = "bump";
          runtimeInputs = with pkgs; [
            git
            nodejs_22
            nix-update
          ];
          text = builtins.readFile ./.scripts/bump.sh;
        });
      };
    });

    formatter = forSystem ({pkgs, ...}: pkgs.alejandra);

    packages = forSystem (
      {pkgs, ...}: let
        client = pkgs.buildNpmPackage {
          inherit pname version;
          src = ./client;
          npmDepsHash = "sha256-tOyEhHa/N0FSSVqxW9PrAY2VIXu24fXbAVylPgYjeYs=";

          installPhase = ''
            cp -r build "$out"
          '';
        };

        server = pkgs.buildGoModule {
          inherit client pname version;
          src = ./server;
          vendorHash = "sha256-uXyCYODrBWNm7nbibm66oO90SYXRvrNtjF0K4ZI7IkM=";
          env.CGO_ENABLED = 0;

          preBuild = ''
            cp -r ${client} client
          '';
        };
      in
        {
          default = server;
        }
        // builtins.listToAttrs (builtins.map (x: {
            name = "${pname}-${x.GOOS}-${x.GOARCH}";
            value = server.overrideAttrs {
              nativeBuildInputs =
                server.nativeBuildInputs
                ++ [
                  pkgs.rename
                ];
              env.CGO_ENABLED = 0;
              env.GOOS = x.GOOS;
              env.GOARCH = x.GOARCH;

              installPhase = ''
                runHook preInstall

                mkdir -p $out/bin
                find $GOPATH/bin -type f -exec mv -t $out/bin {} +
                rename 's/(.+\/)(.+?)(\.[^.]*$|$)/$1${pname}-${x.GOOS}-${x.GOARCH}-${version}$3/' $out/bin/*

                runHook postInstall
              '';
            };
          })
          host-systems)
    );
  };
}
