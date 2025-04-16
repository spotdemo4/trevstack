{
  description = "A template for trevstack";

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
    version = "0.0.11";

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

    formatter = forSystem ({pkgs, ...}: pkgs.alejandra);

    packages = forSystem ({pkgs, ...}: rec {
      default = trevstack;

      trevstack-client = pkgs.buildNpmPackage {
        pname = "${pname}-client";
        inherit version;
        src = ./client;
        npmDepsHash = "sha256-yKabRCijP2oSe6AhMVvxbY/gCA+7/6xSw0YoIzaGezA=";
        nodejs = pkgs.nodejs_22;
        npmFlags = ["--legacy-peer-deps"];

        installPhase = ''
          cp -r build "$out"
          chmod -R u+w "$out"
        '';
      };

      trevstack = pkgs.buildGoModule {
        inherit trevstack-client pname version;
        src = ./server;
        vendorHash = "sha256-ocOqypV4OjlepoMgYFpk/+zpRzBlHg/dljBVMZzS9Yg=";

        preBuild = ''
          cp -r ${trevstack-client} client
        '';
      };
    });
  };
}
