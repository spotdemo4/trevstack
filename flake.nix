{
  description = "A template for trevstack";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    treli.url = "github:spotdemo4/treli";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, gitignore, treli }:
    flake-utils.lib.eachDefaultSystem (system:

      let
        pname = "trevstack";
        version = "0.0.11";

        pkgs = import nixpkgs { 
          inherit system;
          config.allowUnfree = true;
        };

        protoc-gen-connect-openapi = pkgs.buildGoModule {
          name = "protoc-gen-connect-openapi";
          src = pkgs.fetchFromGitHub {
            owner = "sudorandom";
            repo = "protoc-gen-connect-openapi";
            rev = "v0.16.1";
            sha256 = "sha256-3XBQCc9H9N/AZm/8J5bJRgBhVtoZKFvbdTB+glHxYdA=";
          };
          vendorHash = "sha256-CIiG/XhV8xxjYY0sZcSvIFcJ1Wh8LyDDwqem2cSSwBA=";
          nativeCheckInputs = with pkgs; [ less ];
        };

        bobgen = pkgs.buildGoModule {
          name = "bobgen";
          src = pkgs.fetchFromGitHub {
            owner = "stephenafamo";
            repo = "bob";
            rev = "v0.31.0";
            sha256 = "sha256-APAckQ+EDAu459NTPXUISLIrcAcX3aQ5B/jrMUEW0EY=";
          };
          vendorHash = "sha256-3blGiSxlKpWH8k0acAXXks8nCdnoWmXLmzPStJmmGcM=";
          subPackages = [
            "gen/bobgen-sql"
          ];
        };

        client = pkgs.buildNpmPackage {
          pname = "${pname}-client";
          inherit version;
          src = gitignore.lib.gitignoreSource ./client;
          npmDepsHash = "sha256-Mu04whysDA1U5wvECJJ+KopGfSzTPR/OhWz9cjTRIfU=";
          nodejs = pkgs.nodejs_22;
          npmFlags = [ "--legacy-peer-deps" ];

          installPhase = ''
            cp -r build "$out"
            chmod -R u+w "$out"
          '';
        };

        server = pkgs.buildGoModule {
          inherit client pname version;
          src = gitignore.lib.gitignoreSource ./server;
          vendorHash = "sha256-YmMKl9X1kVz6dk/JOSi2jghCUKObUKdm2O+JpO9PDCA=";
          env.CGO_ENABLED = 0;

          preBuild = ''
            cp -r ${client} client
          '';
        };

      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            git
            nix-update
            treli.packages."${system}".default
            sqlite

            # Go backend
            go
            gotools
            gopls
            revive
            bobgen
            dbmate
            
            # Protobuf middleware
            buf
            protoc-gen-go
            protoc-gen-connect-go
            protoc-gen-es
            protoc-gen-connect-openapi

            # Svelte frontend
            nodejs_22

            # Update
            (writeShellApplication {
              name = "ts-update";

              text = ''
                git_root=$(git rev-parse --show-toplevel)

                cd "''${git_root}/client"
                npm update --save
                if ! git diff --exit-code package.json package-lock.json; then
                  git add package-lock.json
                  git add package.json
                  git commit -m "build(client): updated npm dependencies"
                fi

                cd "''${git_root}/server"
                go get -u
                go mod tidy
                if ! git diff --exit-code go.mod go.sum; then
                  git add go.mod
                  git add go.sum
                  git commit -m "build(server): updated go dependencies"
                fi

                cd "''${git_root}"
                nix-update --flake --version=skip --subpackage client default
                if ! git diff --exit-code flake.nix; then
                  git add flake.nix
                  git commit -m "build(nix): updated nix hashes"
                fi
              '';
            })

            # Bump version
            (writeShellApplication {
              name = "ts-bump";

              text = ''
                git_root=$(git rev-parse --show-toplevel)
                next_version=$(echo "${version}" | awk -F. -v OFS=. '{$NF += 1 ; print}')

                cd "''${git_root}/client"
                npm version "''${next_version}"
                git add package-lock.json
                git add package.json

                cd "''${git_root}"
                nix-update --flake --version "''${next_version}" --subpackage client default
                git add flake.nix
                git commit -m "bump: v${version} -> v''${next_version}"
                git push origin main

                git tag -a "v''${next_version}" -m "bump: v${version} -> v''${next_version}"
                git push origin "v''${next_version}"
              '';
            })

            # Lint
            (writeShellApplication {
              name = "ts-lint";

              text = ''
                git_root=$(git rev-parse --show-toplevel)

                if [ -n "''${1:-}" ]; then
                  cd "''${git_root}/client"
                  npm run format
                fi

                cd "''${git_root}"
                echo "Linting protobuf"
                buf lint

                cd "''${git_root}/client"
                echo "Linting client"
                npm run check
                npm run lint

                cd "''${git_root}/server"
                echo "Linting server"
                revive -config revive.toml -set_exit_status ./...
              '';
            })

            # Build
            (writeShellApplication {
              name = "ts-build";

              text = ''
                git_root=$(git rev-parse --show-toplevel)

                cd "''${git_root}"
                echo "Building client"
                nix build .#trevstack-client
                cp -a result/. server/client
                chmod -R u+w server/client

                cd "''${git_root}/server"
                echo "Building ${pname}-windows-amd64-${version}.exe"
                GOOS=windows GOARCH=amd64 go build -o "../build/${pname}-windows-amd64-${version}.exe" .

                echo "Building ${pname}-linux-amd64-${version}"
                GOOS=linux GOARCH=amd64 go build -o "../build/${pname}-linux-amd64-${version}" .

                echo "Building ${pname}-linux-amd64-${version}"
                GOOS=linux GOARCH=arm64 go build -o "../build/${pname}-linux-arm64-${version}" .

                echo "Building ${pname}-linux-arm-${version}"
                GOOS=linux GOARCH=arm go build -o "../build/${pname}-linux-arm-${version}" .
              '';
            })
          ];
        };

        packages = rec {
          default = trevstack;

          trevstack = server;
          trevstack-client = client;
        };
      }
    );
}
