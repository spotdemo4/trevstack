{
  description = "go template";

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

              # lint
              go-tools

              # format
              nixfmt
              tombi
              prettier

              # util
              mprocs
              air
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
            ];
          };

          vulnerable = pkgs.mkShell {
            packages = with pkgs; [
              govulncheck # go
              flake-checker # nix
              zizmor # actions
            ];
          };
        };

        apps = pkgs.mkApps {
          default = "go run .";
          dev = "air";
          vendor = "go mod tidy && go mod vendor";
        };

        checks = pkgs.mkChecks {
          go = {
            src = self.packages.${system}.default;
            packages = with pkgs; [
              go-tools
            ];
            script = ''
              go test ./...
              go vet ./...
              staticcheck ./...
            '';
          };

          biome = {
            root = ./.;
            filter = file: file.hasExt "ts" || file.hasExt "tsx" || file.hasExt "html" || file.hasExt "css";
            include = [
              ./.gitignore
              ./biome.json
            ];
            packages = with pkgs; [
              biome
            ];
            script = ''
              biome ci
            '';
          };

          actions = {
            root = ./.;
            files = ./.github/workflows;
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

          nix = {
            root = ./.;
            filter = file: file.hasExt "nix";
            ignore = pkgs.lib.fileset.maybeMissing ./server/vendor;
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
            ignore = pkgs.lib.fileset.maybeMissing ./server/vendor;
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
            ignore = pkgs.lib.fileset.maybeMissing ./server/vendor;
            packages = with pkgs; [
              tombi
            ];
            forEach = ''
              tombi format --offline --check "$file"
              tombi lint --offline --error-on-warnings "$file"
            '';
          };
        };

        formatter = pkgs.treefmt.withConfig {
          configFile = ./treefmt.toml;
          runtimeInputs = with pkgs; [
            go
            biome
            nixfmt
            tombi
            prettier
          ];
        };

        packages = rec {
          default = server;

          server = pkgs.buildGo125Module (
            final: with pkgs.lib; {
              pname = "trevstack-server";
              version = "0.0.1";

              src = fileset.toSource {
                root = ./server;
                fileset = fileset.unions [
                  ./server/go.mod
                  ./server/go.sum
                  (fileset.fileFilter (file: file.hasExt "go") ./server)
                  (fileset.maybeMissing ./server/vendor)
                ];
              };
              goSum = ./server/go.sum;
              vendorHash = null;

              meta = {
                mainProgram = "server";
                description = "trevstack-server";
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
              version = "0.0.1";

              src = fileset.toSource {
                root = ./web;
                fileset = fileset.unions [
                  ./web/.npmrc
                  ./web/index.html
                  ./web/package.json
                  ./web/package-lock.json
                  ./web/tsconfig.json
                  ./web/vite.config.ts
                  ./web/src
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
                description = "trevstack-web";
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
        };

        schemas = trev.schemas;
      }
    );
}
