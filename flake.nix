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
              oxfmt

              # proto
              buf
              protoc-gen-go
              protoc-gen-connect-go
              protoc-gen-connect-openapi
              protoc-gen-es

              # sql
              sqlfluff

              # actions
              zizmor

              # nix
              nixd
              nil
              nixfmt

              # util
              treefmt
              mprocs
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
          dev = "mprocs";
          server = "cd server && go run -tags dev .";
          web = "cd web && npm run dev";
          gen = ''
            buf generate
            cd server && go mod tidy && cd ..
            cd web && npm install && cd ..
            treefmt
          '';
        };

        formatter = pkgs.treefmt.withConfig {
          configFile = ./treefmt.toml;
          runtimeInputs = with pkgs; [
            go
            oxfmt
            buf
            sqlfluff
            nixfmt
          ];
        };

        packages = rec {
          default = server;

          web = pkgs.buildNpmPackage (
            final: with pkgs.lib; {
              pname = "trevstack-web";
              version = "0.13.0";

              src = ./web;
              nodejs = pkgs.nodejs_24;
              npmConfigHook = pkgs.importNpmLock.npmConfigHook;
              npmDeps = pkgs.importNpmLock {
                npmRoot = final.src;
              };

              nativeCheckInputs = with pkgs; [
                oxfmt
                oxlint
              ];
              checkPhase = ''
                oxfmt --check
                oxlint --deny-warnings
              '';

              installPhase = ''
                cp -r dist "$out"
              '';
            }
          );

          server = pkgs.buildGoModule (
            final: with pkgs.lib; {
              pname = "trevstack-server";
              version = "0.13.0";

              src = ./server;
              goSum = ./server/go.sum;
              proxyVendor = true;
              vendorHash = "sha256-nNTfI8WPpjPw25DVxJCdARUQFA07B5OnHEtpPvi1Eq8=";

              postConfigure = ''
                cp -r ${self.packages.${system}.web} web
              '';

              nativeCheckInputs = with pkgs; [
                go-tools
              ];
              checkPhase = ''
                export HOME=$(mktemp -d)
                go test ./...
                go vet ./...
                staticcheck ./...
              '';

              meta = {
                mainProgram = "server";
                description = "full-stack template";
                license = licenses.mit;
                platforms = platforms.all;
                homepage = "https://trev.zip/llc/stack";
                changelog = "https://trev.zip/llc/stack/releases";
                downloadPage = "https://trev.zip/llc/stack/releases/tag/v${final.version}";
              };
            }
          );
        };

        images.default = pkgs.mkImage {
          src = self.packages.${system}.default;
          contents = with pkgs; [ dockerTools.caCertificates ];
          config.ExposedPorts = {
            "8080/tcp" = { };
          };
        };

        checks = pkgs.mkChecks {
          web = self.packages.${system}.web.overrideAttrs {
            dontBuild = true;
            installPhase = ''
              touch $out
            '';
          };

          server = self.packages.${system}.server.overrideAttrs {
            dontBuild = true;
            installPhase = ''
              touch $out
            '';

            # skip bundling web
            postConfigure = "";
            env.GOFLAGS = "-tags=dev";
          };

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
              hash = "sha256-GTNJ2FSF9ljf7zgp0B7mFDxebbanl3HqBVm07TTkCRo=";
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

          actions = {
            root = ./.;
            files = [
              ./.forgejo/workflows
              ./.github/workflows
            ];
            filter = file: file.hasExt "yaml";
            packages = with pkgs; [
              zizmor
            ];
            script = ''
              zizmor --offline "$file"
            '';
          };

          renovate = {
            root = ./.github;
            files = ./.github/renovate.json;
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
