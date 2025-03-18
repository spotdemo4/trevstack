{
  description = "A trevstack development environment";

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
        version = "0.1.1";

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
          npmDepsHash = "sha256-TO7eumSEJdmG4wZ5B/k8gChmmSDvfYKEtExHLLYHjj4=";
          nodejs = pkgs.nodejs_22;
          npmFlags = [ "--legacy-peer-deps" ];

          installPhase = ''
            cp -r build "$out"
            chmod -R u+w "$out"
          '';
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

            # Nix
            nix-update

            # Helper scripts
            (writeShellApplication {
              name = "run";

              text = ''
                gitroot=$(git rev-parse --show-toplevel)

                (cd "''${gitroot}/server" && air) &
                P1=$!

                (cd "''${gitroot}/client" && npm run dev) &
                P2=$!

                (cd "''${gitroot}" && protobufwatch) &
                P3=$!

                trap 'kill $P1 $P2 $P3' SIGINT SIGTERM
                wait $P1
                wait $P2
                wait $P3
                
                kill $P1 $P2 $P3
              '';
            })

            (writeShellApplication {
              name = "protobufwatch";

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
          ];
        };

        packages.default = pkgs.buildGoModule {
          inherit pname version;
          src = gitignore.lib.gitignoreSource ./server;
          vendorHash = "sha256-FyqcKhJy58uL3UiGA9tg7pSt0OQ1NIZw+khTictyzcw=";

          preBuild = ''
            cp -r ${client} internal/handlers/client/client
          '';
        };
      }
    );
}
