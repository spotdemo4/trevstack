{
  description = "TrevStack is a template for full-stack applications built with Go, SolidJS, and ConnectRPC.";

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

              # solid
              nodejs_24

              # proto
              buf
              protoc-gen-go
              protoc-gen-connect-go
              protoc-gen-connect-openapi
              protoc-gen-es

              # lint
              go-tools
              oxlint
              sqlfluff
              nixd
              nil

              # format
              oxfmt
              nixfmt
              treefmt

              # util
              mprocs
              bumper
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
              nodejs_24 # npm audit fix
            ];
          };

          vulnerable = pkgs.mkShell {
            packages = with pkgs; [
              govulncheck # go
              nodejs_24 # npm audit
              flake-checker # nix
              zizmor # actions
            ];
          };
        };

        apps = pkgs.mkApps {
          default = "mprocs";
          server = "cd server && go run -tags dev .";
          web = "cd web && npm run dev";
          gen = ''
            buf generate
            cd server && go mod tidy && go mod vendor && cd ..
            cd web && npm install && cd ..
            treefmt
          '';
        };

        formatter = pkgs.treefmt.withConfig {
          configFile = ./treefmt.toml;
          runtimeInputs = with pkgs; [
            oxfmt
            go
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
              version = "0.6.0";

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

          server = pkgs.buildGo125Module (
            final: with pkgs.lib; {
              pname = "trevstack-server";
              version = "0.6.0";

              src = ./server;
              goSum = ./server/go.sum;
              proxyVendor = true;
              vendorHash = "sha256-fSpN4gsxHBPt/ojWos8UR7quSt5cU9fO70k1+pmS7FQ=";

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
                description = "Template for TrevStack applications";
                license = licenses.mit;
                platforms = platforms.all;
                homepage = "https://github.com/spotdemo4/trevstack";
                changelog = "https://github.com/spotdemo4/trevstack/releases/tag/v${final.version}";
                downloadPage = "https://github.com/spotdemo4/trevstack/releases/tag/v${final.version}";
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
          };

          oxfmt = {
            root = ./.;
            filter = file: file.hasExt "json" || file.hasExt "yaml" || file.hasExt "toml" || file.hasExt "md";

            packages = with pkgs; [
              oxfmt
            ];

            script = ''
              oxfmt --check
            '';
          };

          sqlfluff = {
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

          nixfmt = {
            root = ./.;
            filter = file: file.hasExt "nix";

            packages = with pkgs; [
              nixfmt
            ];

            forEach = ''
              nixfmt --check "$file"
            '';
          };

          buf = {
            root = ./.;
            filter = file: file.hasExt "proto";
            include = [
              ./buf.lock
              ./buf.yaml
              ./buf.gen.yaml
            ];

            bufDeps = pkgs.bufFetchDeps {
              pname = "trevstack-proto-deps";
              src = ./.;
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
            root = ./.github/workflows;
            filter = file: file.hasExt "yaml";

            packages = with pkgs; [
              action-validator
              zizmor
            ];

            forEach = ''
              action-validator "$file"
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
        };
      }
    );
}
