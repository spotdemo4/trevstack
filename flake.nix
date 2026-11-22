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
      url = "git+https://trev.zip/llc/TrevRPC.git?ref=refs/tags/v1.12.0";
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
      system: pkgs:
      let
        androidSdk = pkgs.callPackage ./android/sdk.nix { };
      in
      {

        # nix develop [#...]
        devShells = {
          default = pkgs.mkShell {
            shellHook = pkgs.trev.shellhook.ref;
            env = {
              GOTOOLCHAIN = "local";
              RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;
              PLAYWRIGHT_BROWSERS_PATH = pkgs.playwright-driver.browsers;
              PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS = "true";
              ANDROID_HOME = "${androidSdk}/libexec/android-sdk";
              # the aapt2 that the Android Gradle plugin downloads doesn't run on NixOS
              GRADLE_OPTS = "-Dorg.gradle.project.android.aapt2FromMavenOverride=${androidSdk}/libexec/android-sdk/build-tools/37.0.0/aapt2";
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

              # android
              jdk25
              gradle_9
              ktlint
              androidSdk

              # proto
              buf
              protoc-gen-go
              protoc-gen-es
              protoc-gen-prost
              protobuf # protoc, which runs buf's java plugin
              trevrpc.packages.${system}.trevrpc-go
              trevrpc.packages.${system}.trevrpc-js
              trevrpc.packages.${system}.trevrpc-rust
              trevrpc.packages.${system}.trevrpc-kotlin

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
              trev.bumper
              trev.fix-hash
              gomod2nix.packages.${system}.default
            ];
          };

          bump = pkgs.mkShell {
            packages = with pkgs; [
              trev.bumper
            ];
          };

          release = pkgs.mkShell {
            packages = with pkgs; [
              trev.flake-release
            ];
          };

          update = pkgs.mkShell {
            # Renovate updates android/app/gradle.lockfile with Gradle, which configures the app
            env.ANDROID_HOME = "${androidSdk}/libexec/android-sdk";
            packages = with pkgs; [
              trev.renovate
              gradle_9 # gradle --write-locks
              jdk25
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
        apps = pkgs.trev.libs.mkApps {
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

          update-android-deps = {
            packages = with pkgs; [
              gradle_9
              jdk25
            ];
            script = ''
              export ANDROID_HOME=${androidSdk}/libexec/android-sdk
              # the dependencies tasks resolve every declared configuration, and the builds resolve
              # the ones the Android Gradle plugin adds while building, which writes all their locks
              gradle -p android --write-locks \
                -Dorg.gradle.project.android.aapt2FromMavenOverride=${androidSdk}/libexec/android-sdk/build-tools/37.0.0/aapt2 \
                dependencies :app:dependencies \
                :app:assembleDebug :app:assembleRelease :app:lintDebug :app:lintRelease :app:testDebugUnitTest
              # then pins every artifact the Nix build downloads in android/deps.json. bwrap only
              # keeps the build out of ~, and needs user namespaces, which CI runners may not allow
              update_script=$(nix build --no-link --print-out-paths .#android.mitmCache.updateScript)
              USE_BWRAP=0 "$update_script"
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
            ktlint
          ];
        };

        # nix build [#...]
        packages = rec {
          default = server;
          server = pkgs.callPackage ./server { inherit web; };
          client = pkgs.callPackage ./client { };
          web = pkgs.buildPackages.callPackage ./web { };
          android = pkgs.buildPackages.callPackage ./android { };
        };

        # nix build #images.[...]
        images = rec {
          default = server;
          server = pkgs.trev.libs.mkImage {
            src = self.packages.${system}.default;
            contents = with pkgs; [ dockerTools.caCertificates ];
            # QUIC, HTTP/3 and WebTransport use UDP; their fallbacks use TCP
            config.ExposedPorts = {
              "8080/tcp" = { };
              "8080/udp" = { };
            };
          };
          client = pkgs.trev.libs.mkImage {
            src = self.packages.${system}.client;
            contents = with pkgs; [ dockerTools.caCertificates ];
          };
        };

        # nix flake check
        checks = pkgs.trev.libs.mkChecks {
          # each package's checks.nix, named <package>-<check>
          server = pkgs.trev.libs.callChecks ./server { inherit self; };
          client = pkgs.trev.libs.callChecks ./client { inherit self; };
          web = pkgs.trev.libs.callChecks ./web { inherit self; };
          android = ./android;

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
            bufDeps = pkgs.trev.libs.bufFetchDeps {
              src = ./.;
              pname = "trevstack-proto-deps";
              hash = "sha256-TvDzKn9ZUOylwfCmxND3bwhmHwPCg8NfbNL/nkvQnS8=";
            };
            packages = with pkgs; [
              trev.libs.bufHook
              buf
            ];
            script = ''
              buf lint
              buf format -d --exit-code
            '';
          };

          protobuf-generate = {
            root = ./.;
            files = [
              ./proto
              ./buf.lock
              ./buf.yaml
              ./buf.gen.yaml
              ./server/gen
              ./web/gen
              ./client/src/gen
              ./android/app/src/gen
            ];
            inherit (self.checks.${system}.protobuf-lint) bufDeps;
            packages = with pkgs; [
              trev.libs.bufHook
              buf
              protoc-gen-go
              protoc-gen-es
              protoc-gen-prost
              protobuf
              trevrpc.packages.${system}.trevrpc-go
              trevrpc.packages.${system}.trevrpc-js
              trevrpc.packages.${system}.trevrpc-rust
              trevrpc.packages.${system}.trevrpc-kotlin
            ];
            # regenerates every plugin output from scratch, so stale and leftover files both fail
            script = ''
              committed=$(mktemp -d)
              cp -r . "$committed"
              sed -n 's/^ *out: *//p' buf.gen.yaml | sort -u | xargs rm -rf
              buf generate
              if ! diff -ru "$committed" .; then
                echo "the generated code does not match buf generate, run buf generate"
                exit 1
              fi
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

          actions-gt = {
            root = ./.gitea/workflows;
            filter = file: file.hasExt "yaml";
            packages = with pkgs; [
              action-validator
              gitea-actions-runner
              zizmor
            ];
            script = ''
              gitea-runner exec --list --workflows "$file"
              action-validator "$file"
              zizmor --offline "$file"
            '';
          };

          renovate-gh = {
            root = ./.github;
            files = ./.github/renovate.jsonc;
            packages = with pkgs; [ trev.renovate ];
            script = ''
              renovate-config-validator renovate.jsonc
            '';
          };

          renovate-fj = {
            root = ./.forgejo;
            files = ./.forgejo/renovate.jsonc;
            packages = with pkgs; [ trev.renovate ];
            script = ''
              renovate-config-validator renovate.jsonc
            '';
          };

          renovate-gt = {
            root = ./.gitea;
            files = ./.gitea/renovate.jsonc;
            packages = with pkgs; [ trev.renovate ];
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
