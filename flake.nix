{
  description = "A trevstack development environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }@inputs:
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
            goreleaser
            air
            
            # Protobuf middleware
            buf
            protoc-gen-go
            protoc-gen-connect-go
            protoc-gen-connect-openapi

            # Svelte frontend
            nodejs_22

            # Helper scripts
            (writeShellApplication {
              name = "run";

              text = ''
                gitroot=$(git rev-parse --show-toplevel)
                (cd "''${gitroot}" && air) & (cd "''${gitroot}" && npm run dev) && fg 
              '';
            })

            (writeShellApplication {
              name = "build";

              text = ''
                gitroot=$(git rev-parse --show-toplevel)
                (cd "''${gitroot}" && npm run build) && (cd "''${gitroot}" && go build -o ./build_server/golte .)
              '';
            })

            (writeShellApplication {
              name = "gen";

              text = ''
                gitroot=$(git rev-parse --show-toplevel)
                (cd "''${gitroot}" && buf generate)
              '';
            })
          ];
        };
      }
    );
}
