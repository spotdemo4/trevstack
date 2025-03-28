{
  description = "A template for trevstack";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, gitignore }:
    flake-utils.lib.eachDefaultSystem (system:

      let
        pname = "trevstack";
        version = "0.0.10";

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

        client = pkgs.buildNpmPackage {
          pname = "${pname}-client";
          inherit version;
          src = gitignore.lib.gitignoreSource ./client;
          npmDepsHash = "sha256-GQNChIR5kEInghn/W3t8BVnOv5X4pvn/SNCXyFnvLNo=";
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
          vendorHash = "sha256-sANPwYLGwMcWyMR7Veho81aAMfIQpVzZS5Q9eveR8o8=";
          env.CGO_ENABLED = 0;

          preBuild = ''
            cp -r ${client} internal/handlers/client/client
          '';
        };

      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            git
            nix-update

            # Go backend
            go
            gotools
            gopls
            air
            revive
            
            # Protobuf middleware
            buf
            protoc-gen-go
            protoc-gen-connect-go
            protoc-gen-es
            protoc-gen-connect-openapi
            inotify-tools

            # Svelte frontend
            nodejs_22

            # Helper scripts
            (writeShellApplication {
              name = "ts-run";

              text = ''
                git_root=$(git rev-parse --show-toplevel)

                (cd "''${git_root}/server" && air) &
                P1=$!

                (cd "''${git_root}/client" && npm run dev) &
                P2=$!

                (cd "''${git_root}" && ts-pbwatch) &
                P3=$!

                trap 'kill $P1 $P2 $P3' SIGINT SIGTERM
                wait $P1
                wait $P2
                wait $P3
                
                kill $P1 $P2 $P3
              '';
            })

            (writeShellApplication {
              name = "ts-pbwatch";

              text = ''
                inotifywait -mre close_write,moved_to,create proto | while read -r _ _ basename;
                do
                  echo "file changed: $basename"
                  if buf lint ; then
                    buf generate
                  fi
                  echo "regenerated proto services"
                done
              '';
            })

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
                revive -config revive.toml -formatter friendly ./...
              '';
            })

            (writeShellApplication {
              name = "ts-build";

              text = ''
                git_root=$(git rev-parse --show-toplevel)

                cd "''${git_root}"
                echo "Building client"
                nix build .#trevstack-client
                cp -a result/. server/internal/handlers/client/client
                chmod -R u+w server/internal/handlers/client/client

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
