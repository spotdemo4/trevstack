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
    trev = {
      url = "github:spotdemo4/nur";
      inputs.systems.follows = "systems";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      trev,
      ...
    }:
    trev.libs.mkFlake (
      system: pkgs: {
        devShells = {
          default = pkgs.mkShell {
            shellHook = pkgs.shellhook.ref;
            packages = with pkgs; [
              # go
              go
              gotools
              gopls

              # solid
              nodejs_24

              # proto
              buf
              protoc-gen-go
              protoc-gen-connect-go
              protoc-gen-es
              protoc-gen-connect-openapi

              # lint
              go-tools
              biome
              sqlfluff

              # format
              nixfmt
              tombi
              prettier
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
              go # go mod vendor
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
          vendor = "cd server && go mod tidy && go mod vendor";
        };

        formatter = pkgs.treefmt.withConfig {
          configFile = ./treefmt.toml;
          runtimeInputs = with pkgs; [
            go
            biome
            buf
            sqlfluff
            nixfmt
            tombi
            prettier
          ];
        };

        packages = rec {
          default = server;

          server = pkgs.buildGo125Module (
            final: with pkgs.lib; {
              pname = "trevstack";
              version = "0.2.2";

              src = fileset.toSource {
                root = ./server;
                fileset = fileset.unions [
                  ./server/embed.go
                  ./server/go.mod
                  ./server/go.sum
                  ./server/main.go
                  ./server/connect
                  ./server/database
                  ./server/handlers
                  ./server/interceptors
                  ./server/logger
                  ./server/vendor
                ];
              };
              goSum = ./server/go.sum;
              vendorHash = null;

              preBuild = ''
                cp -r ${self.packages.${system}.web} web
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

          web = pkgs.buildNpmPackage (
            final: with pkgs.lib; {
              pname = "trevstack-web";
              version = "0.2.2";

              src = fileset.toSource {
                root = ./web;
                fileset = fileset.unions [
                  ./web/.npmrc
                  ./web/index.css
                  ./web/index.html
                  ./web/index.tsx
                  ./web/package.json
                  ./web/package-lock.json
                  ./web/tsconfig.json
                  ./web/vite.config.ts
                  ./web/assets
                  ./web/connect
                  ./web/layout
                  ./web/lib
                  ./web/routes
                ];
              };

              nodejs = pkgs.nodejs_24;
              npmConfigHook = pkgs.importNpmLock.npmConfigHook;
              npmDeps = pkgs.importNpmLock {
                npmRoot = final.src;
              };

              installPhase = ''
                cp -r dist "$out"
              '';

              meta = {
                description = "Template for TrevStack applications - web client";
                license = licenses.mit;
                platforms = platforms.all;
                homepage = "https://github.com/spotdemo4/trevstack";
                changelog = "https://github.com/spotdemo4/trevstack/releases/tag/v${final.version}";
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
          go = {
            root = ./server;
            packages = with pkgs; [
              gcc
              go
              go-tools
            ];
            script = ''
              go test -tags dev ./...
              go vet -tags dev ./...
              staticcheck -tags dev ./...
            '';
          };

          biome = {
            root = ./.;
            filter = file: file.hasExt "ts" || file.hasExt "tsx" || file.hasExt "html" || file.hasExt "css";
            include = [
              ./.gitignore
              ./biome.json
            ];
            ignore = [
              ./server/vendor
              ./web/connect
            ];
            packages = with pkgs; [
              biome
            ];
            script = ''
              biome ci
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

          sql = {
            root = ./.;
            filter = file: file.hasExt "sql";
            ignore = ./server/vendor;
            packages = with pkgs; [
              sqlfluff
            ];
            forEach = ''
              sqlfluff lint --dialect sqlite "$file"
            '';
          };

          nix = {
            root = ./.;
            filter = file: file.hasExt "nix";
            ignore = ./server/vendor;
            packages = with pkgs; [
              nixfmt
            ];
            forEach = ''
              nixfmt --check "$file"
            '';
          };

          prettier = {
            root = ./.;
            filter = file: file.hasExt "yaml" || file.hasExt "json" || file.hasExt "md";
            ignore = ./server/vendor;
            packages = with pkgs; [
              prettier
            ];
            forEach = ''
              prettier --check "$file"
            '';
          };

          tombi = {
            root = ./.;
            filter = file: file.hasExt "toml";
            ignore = ./server/vendor;
            packages = with pkgs; [
              tombi
            ];
            forEach = ''
              tombi format --offline --check "$file"
              tombi lint --offline --error-on-warnings "$file"
            '';
          };
        };
      }
    );
}
