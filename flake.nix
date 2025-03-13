{
  description = "A trevstack development environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { 
          inherit system;
          config.allowUnfree = true;
        };
        protoc-gen-connect-openapi = pkgs.buildGoModule rec {
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
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ 
            # Go backend
            go
            gotools
            gopls
            air
            
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
              name = "run";

              text = ''
                gitroot=$(git rev-parse --show-toplevel)

                (cd "''${gitroot}/server" && air) &
                P1=$!

                (cd "''${gitroot}/client" && npm run dev) &
                P2=$!

                protobufwatch &
                P3=$!

                trap 'kill $P1 $P2 $P3' SIGINT SIGTERM
                wait $P1
                wait $P2
                wait $P3
              '';
            })

            (writeShellApplication {
              name = "build";

              text = ''
                gitroot=$(git rev-parse --show-toplevel)

                cd "''${gitroot}"
                buf lint
                buf generate

                cd "''${gitroot}/client"
                npm run build
                cp -r build ../server/client

                cd "''${gitroot}/server"
                go build -o ../build/trevstack .
              '';
            })

            (writeShellApplication {
              name = "protobufwatch";

              text = ''
                gitroot=$(git rev-parse --show-toplevel)

                cd "''${gitroot}"
                inotifywait -mre close_write,moved_to,create proto | while read -r _ _ basename;
                do
                  echo "file changed: $basename"
                  buf lint
                  buf generate
                  echo "regenerated proto services"
                done
              '';
            })
          ];
        };

        packages.default = pkgs.stdenv.mkDerivation {
          pname = "trevstack";
          version = "1.0";

          buildInputs = with pkgs; [
            # Go backend
            go
            gotools
            gopls
            
            # Protobuf middleware
            buf
            protoc-gen-go
            protoc-gen-connect-go
            protoc-gen-es
            protoc-gen-connect-openapi

            # Svelte frontend
            nodejs_22
          ];

          buildPhase = ''
            gitroot=$(git rev-parse --show-toplevel)

            cd "''${gitroot}"
            buf lint
            buf generate

            cd "''${gitroot}/client"
            npm run build
            cp -r build ../server/client

            cd "''${gitroot}/server"
            go build -o ../build/trevstack .
          '';

          installPhase = ''
            mkdir -p $out/bin
            cp build/trevstack $out/bin
          '';
        };
      }
    );
}
