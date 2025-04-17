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
    version = "0.0.18";

    supportedSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
    forSystem = f:
      nixpkgs.lib.genAttrs supportedSystems (
        system:
          f {
            inherit system;
            pkgs = import nixpkgs {
              inherit system;
            };
          }
      );
  in {
    devShells = forSystem ({pkgs, ...}: {
      default = pkgs.mkShell {
        packages = with pkgs;
          [
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
            (buildGoModule {
              name = "protoc-gen-connect-openapi";
              src = pkgs.fetchFromGitHub {
                owner = "sudorandom";
                repo = "protoc-gen-connect-openapi";
                rev = "v0.16.1";
                sha256 = "sha256-3XBQCc9H9N/AZm/8J5bJRgBhVtoZKFvbdTB+glHxYdA=";
              };
              vendorHash = "sha256-CIiG/XhV8xxjYY0sZcSvIFcJ1Wh8LyDDwqem2cSSwBA=";
              nativeCheckInputs = with pkgs; [less];
            })

            # Client
            nodejs_22
          ]
          # Use .scripts
          ++ map (
            x: (
              pkgs.writeShellApplication {
                name = "${pname}-${(lib.nameFromURL (baseNameOf x) ".")}";
                text = builtins.readFile x;
              }
            )
          ) (pkgs.lib.filesystem.listFilesRecursive ./.scripts);
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
          npmDepsHash = "sha256-u7zkBgaxDEB2XFrNl0f7/HtW0Oy2B7FVPot9MLPzXGc=";
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

    formatter = forSystem ({pkgs, ...}: pkgs.alejandra);

    packages = forSystem ({pkgs, ...}: rec {
      trevstack-client = pkgs.buildNpmPackage {
        pname = "${pname}-client";
        inherit version;
        src = ./client;
        npmDepsHash = "sha256-u7zkBgaxDEB2XFrNl0f7/HtW0Oy2B7FVPot9MLPzXGc=";
        nodejs = pkgs.nodejs_22;

        installPhase = ''
          cp -r build "$out"
          chmod -R u+w "$out"
        '';
      };
      trevstack = pkgs.buildGoModule {
        inherit trevstack-client pname version;
        src = ./server;
        vendorHash = "sha256-uXyCYODrBWNm7nbibm66oO90SYXRvrNtjF0K4ZI7IkM=";

        preBuild = ''
          cp -r ${trevstack-client} client
        '';
      };
      default = trevstack;
    });
  };
}
