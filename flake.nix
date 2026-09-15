{
  description = "full-stack template";

  nixConfig = {
    extra-substituters = [
      "https://nix.trev.zip"
    ];
    extra-trusted-public-keys = [
      "trev:I39N/EsnHkvfmsbx8RUW+ia5dOzojTQNCTzKYij1chU="
    ];
  };

  inputs = {
    systems.url = "github:spotdemo4/systems";
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    trevpkgs = {
      url = "github:spotdemo4/trevpkgs";
      inputs.systems.follows = "systems";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      trevpkgs,
      ...
    }:
    trevpkgs.libs.mkFlake (
      system: pkgs: {
        devShells = {
          default = pkgs.mkShell {
            shellHook = pkgs.shellhook.ref;
            RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;
            packages = with pkgs; [
              # go
              go
              gopls
              gotools
              go-tools
              govulncheck

              # solid
              nodejs_24
              oxlint

              # rust
              rustc
              cargo
              rust-analyzer
              rustfmt
              clippy

              # proto
              buf
              protoc-gen-go
              protoc-gen-es
              protoc-gen-connect-go
              protoc-gen-connect-openapi
              protoc-gen-prost
              protoc-gen-tonic

              # sql
              sqlfluff
              sqls

              # nix
              nixd
              nixfmt

              zizmor # actions
              vscode-json-languageserver # json
              yaml-language-server # yaml
              tombi # toml
              oxfmt # format

              # util
              treefmt
              bumper
              fix-hash
            ];
          };

          bump = pkgs.mkShell {
            packages = with pkgs; [
              bumper
            ];
          };

          release = pkgs.mkShell {
            packages = with pkgs; [
              flake-release
            ];
          };

          update = pkgs.mkShell {
            packages = with pkgs; [
              renovate
              go # go get
              cargo # cargo update
              buf # buf dep update
              nodejs_24 # npm audit fix
              fix-hash # vendorHash & bufDeps
            ];
          };

          vulnerable = pkgs.mkShell {
            packages = with pkgs; [
              # go
              go
              govulncheck

              nodejs_24 # npm audit
              flake-checker # nix flake
              zizmor # actions
            ];
          };
        };

        apps = pkgs.mkApps {
          configure = {
            inputsFrom = [ self.devShells.${system}.default ];
            script = ''
              buf generate
              cd server && go mod tidy && cd ..
              cd docs && npm install && cd ..
              cd web && npm install && cd ..
              treefmt
            '';
          };
        };

        formatter = pkgs.treefmt.withConfig {
          configFile = ./treefmt.toml;
          runtimeInputs = with pkgs; [
            go
            oxfmt
            buf
            sqlfluff
            nixfmt
            rustfmt
          ];
        };

        packages = rec {
          default = server;
          client = pkgs.callPackage ./client { };
          docs = pkgs.buildPackages.callPackage ./docs { };
          web = pkgs.buildPackages.callPackage ./web { };
          server = pkgs.callPackage ./server { inherit docs web; };
        };

        images.default = pkgs.mkImage {
          src = self.packages.${system}.default;
          contents = with pkgs; [ dockerTools.caCertificates ];
          config.ExposedPorts = {
            "8080/tcp" = { };
          };
        };

        checks = pkgs.mkChecks {
          inherit (self.packages.${system})
            client
            docs
            web
            server
            ;

          sql = {
            root = ./.;
            filter = file: file.hasExt "sql";
            include = [
              ./.sqlfluff
            ];
            packages = with pkgs; [
              sqlfluff
            ];
            script = ''
              sqlfluff lint
            '';
          };

          nix = {
            root = ./.;
            filter = file: file.hasExt "nix";
            packages = with pkgs; [
              nixfmt
            ];
            script = ''
              nixfmt --check "$file"
            '';
          };

          rust = {
            root = ./client;
            filter = file: file.hasExt "rs";
            include = [
              ./client/Cargo.lock
              ./client/Cargo.toml
            ];
            packages = with pkgs; [
              cargo
              rustfmt
            ];
            script = ''
              cargo fmt --check
            '';
          };

          protobuf = {
            root = ./.;
            filter = file: file.hasExt "proto";
            include = [
              ./buf.lock
              ./buf.yaml
              ./buf.gen.yaml
            ];
            bufDeps = pkgs.bufFetchDeps {
              src = ./.;
              pname = "trevstack-proto-deps";
              hash = "sha256-FHWO4jScAnsb5BjWZbbEXGlfWMh3MKR4f9xPwzfJc0I=";
            };
            packages = with pkgs; [
              bufHook
              buf
            ];
            script = ''
              buf lint
              buf format -d --exit-code
            '';
          };

          actions-gh = {
            root = ./.github/workflows;
            filter = file: file.hasExt "yaml";
            packages = with pkgs; [
              action-validator
              zizmor
            ];
            script = ''
              action-validator "$file"
              zizmor --offline "$file"
            '';
          };

          actions-fj = {
            root = ./.forgejo/workflows;
            filter = file: file.hasExt "yaml";
            packages = with pkgs; [
              forgejo-runner
              zizmor
            ];
            script = ''
              forgejo-runner validate --workflow --path "$file"
              zizmor --offline "$file"
            '';
          };

          renovate-gh = {
            root = ./.github;
            files = ./.github/renovate.json;
            packages = with pkgs; [
              renovate
            ];
            script = ''
              renovate-config-validator renovate.json
            '';
          };

          renovate-fj = {
            root = ./.forgejo;
            files = ./.forgejo/renovate.json;
            packages = with pkgs; [
              renovate
            ];
            script = ''
              renovate-config-validator renovate.json
            '';
          };

          config = {
            root = ./.;
            filter = file: file.hasExt "json" || file.hasExt "yaml" || file.hasExt "toml" || file.hasExt "md";
            packages = with pkgs; [
              oxfmt
            ];
            script = ''
              oxfmt --check
            '';
          };
        };
      }
    );
}
