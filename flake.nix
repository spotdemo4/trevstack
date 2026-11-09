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
    trevrpc = {
      url = "git+https://trev.zip/llc/TrevRPC.git?ref=refs/tags/v1.7.0";
      inputs.systems.follows = "systems";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.trevpkgs.follows = "trevpkgs";
    };
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.inputs.systems.follows = "systems";
    };
  };

  outputs =
    {
      self,
      trevpkgs,
      trevrpc,
      gomod2nix,
      ...
    }:
    trevpkgs.libs.mkFlake (
      system: pkgs: {

        # nix develop [#...]
        devShells = {
          default = pkgs.mkShell {
            shellHook = pkgs.shellhook.ref;
            env = {
              GOTOOLCHAIN = "local";
              RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;
              PLAYWRIGHT_BROWSERS_PATH = pkgs.playwright-driver.browsers;
              PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS = "true";
            };
            packages = with pkgs; [
              # go
              go
              gopls
              gotools
              golangci-lint
              govulncheck

              # solid
              nodejs_26
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
              protoc-gen-prost
              trevrpc.packages.${system}.trevrpc-go
              trevrpc.packages.${system}.trevrpc-rust

              # python
              (python3.withPackages (ps: [ ps.playwright ]))
              ruff
              pyright

              # nix
              nixd
              nixfmt

              oxfmt # format
              sqlfluff # sql
              tombi # toml
              vscode-json-languageserver # json
              yaml-language-server # yaml
              zizmor # actions

              # util
              treefmt
              bumper
              fix-hash
              gomod2nix.packages.${system}.default
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
              nodejs_26 # npm install
            ];
          };

          vulnerable = pkgs.mkShell {
            packages = with pkgs; [
              go # for govluncheck
              govulncheck # govulncheck ./...
              cargo-audit # cargo-audit audit
              nodejs_26 # npm audit
              flake-checker # flake-checker -f
              zizmor # actions
            ];
          };
        };

        # nix run [#...]
        apps = pkgs.mkApps {
          configure = {
            inputsFrom = [ self.devShells.${system}.default ];
            script = ''
              buf generate
              go -C server mod tidy
              npm install --prefix web
            '';
          };

          update-buf-deps = {
            inputsFrom = [ self.devShells.${system}.default ];
            script = ''
              buf dep update
              buf generate
              fix-hash .#checks.${system}.protobuf-lint
            '';
          };

          update-go-deps = {
            packages = with pkgs; [
              go
              gomod2nix.packages.${system}.default
            ];
            script = ''
              go -C server mod tidy
              # the default (development) build's packages are a superset of the prod build's
              gomod2nix generate --with-deps --dir server
            '';
          };
        };

        # nix flake init -t git+https://trev.zip/template/stack
        templates = {
          default = {
            path = ./.;
            description = "full-stack template";
            welcomeText = ''
              # full-stack template

              `init.sh` requires a git repository with an `origin` remote:

              ```sh
              git init
              git remote add origin <url>
              bash init.sh "Title" "Description"
              ```
            '';
          };
        };

        # nix fmt
        formatter = pkgs.treefmt.withConfig {
          configFile = ./treefmt.toml;
          runtimeInputs = with pkgs; [
            golangci-lint
            oxfmt
            buf
            sqlfluff
            nixfmt
            rustfmt
            ruff
          ];
        };

        # nix build [#...]
        packages = rec {
          default = server;
          client = pkgs.callPackage ./client { };
          web = pkgs.buildPackages.callPackage ./web { };
          server = pkgs.callPackage ./server { inherit web; };
        };

        # nix build #images.[...]
        images = rec {
          default = server;
          client = pkgs.mkImage {
            src = self.packages.${system}.client;
            contents = with pkgs; [ dockerTools.caCertificates ];
          };
          server = pkgs.mkImage {
            src = self.packages.${system}.default;
            contents = with pkgs; [ dockerTools.caCertificates ];
            # QUIC, HTTP/3 and WebTransport use UDP; their fallbacks use TCP
            config.ExposedPorts = {
              "8080/tcp" = { };
              "8080/udp" = { };
            };
          };
        };

        # nix flake check
        checks = pkgs.mkChecks {
          inherit (self.packages.${system})
            client
            web
            server
            ;

          server-lint = {
            src = self.packages.${system}.server;
            packages = with pkgs; [ golangci-lint ];
            script = ''
              golangci-lint run ./...
              golangci-lint run --build-tags=prod ./...
            '';
          };

          client-lint = {
            src = self.packages.${system}.client;
            packages = with pkgs; [
              rustfmt
              clippy
            ];
            script = ''
              cargo fmt --check
              cargo clippy --offline --all-targets -- -D warnings
            '';
          };

          web-lint = {
            src = self.packages.${system}.web;
            packages = with pkgs; [
              oxfmt
              oxlint
            ];
            script = ''
              oxfmt --check
              oxlint --deny-warnings
            '';
          };

          integration = {
            root = ./test;
            packages = with pkgs; [ (python3.withPackages (ps: [ ps.playwright ])) ];
            env = {
              PLAYWRIGHT_BROWSERS_PATH = pkgs.playwright-driver.browsers;
              PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS = "true";
              # the build sandbox has no fonts, and without them text has no size and is never visible
              FONTCONFIG_FILE = pkgs.makeFontsConf { fontDirectories = [ pkgs.dejavu_fonts ]; };
            };
            script = ''
              for test in test_*.py; do
                python3 "$test" \
                  ${pkgs.lib.getExe self.packages.${system}.client} \
                  ${pkgs.lib.getExe self.packages.${system}.server}
              done
            '';
            __darwinAllowLocalNetworking = true;
          };

          init = {
            root = ./.;
            packages = with pkgs; [
              git
              jq
              shellcheck
              self.formatter.${system}
            ];
            script = ''
              shellcheck init.sh init.test.sh
              bash init.test.sh
            '';
          };

          sql-lint = {
            root = ./.;
            filter = file: file.hasExt "sql";
            include = [ ./.sqlfluff ];
            packages = with pkgs; [ sqlfluff ];
            script = ''
              sqlfluff lint
            '';
          };

          nix = {
            root = ./.;
            filter = file: file.hasExt "nix";
            packages = with pkgs; [ nixfmt ];
            script = ''
              nixfmt --check "$file"
            '';
          };

          python-lint = {
            root = ./.;
            filter = file: file.hasExt "py";
            include = [ ./pyproject.toml ];
            packages = with pkgs; [
              (python3.withPackages (ps: [ ps.playwright ]))
              ruff
              pyright
            ];
            script = ''
              ruff check
              ruff format --check
              pyright
            '';
          };

          protobuf-lint = {
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
              hash = "sha256-wkF8EYPfVivwqyfQwizSTbviVBuylABBem6k9wheCjY=";
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
            files = ./.github/renovate.jsonc;
            packages = with pkgs; [ renovate ];
            script = ''
              renovate-config-validator renovate.jsonc
            '';
          };

          renovate-fj = {
            root = ./.forgejo;
            files = ./.forgejo/renovate.jsonc;
            packages = with pkgs; [ renovate ];
            script = ''
              renovate-config-validator renovate.jsonc
            '';
          };

          config = {
            root = ./.;
            filter =
              file:
              file.hasExt "json"
              || file.hasExt "jsonc"
              || file.hasExt "yaml"
              || file.hasExt "toml"
              || file.hasExt "md";
            packages = with pkgs; [ oxfmt ];
            script = ''
              oxfmt --check
            '';
          };
        };
      }
    );
}
