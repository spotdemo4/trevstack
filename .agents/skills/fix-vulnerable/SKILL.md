---
name: fix-vulnerable
description: Fixes failures reported by the vulnerable workflow, which runs govulncheck, cargo audit, npm audit, osv-scanner, flake-checker and zizmor. Use when the vulnerable workflow fails, or when asked about a govulncheck finding, a Go vulnerability (GO-YYYY-NNNN), a RUSTSEC advisory, an npm advisory, an Android or Gradle dependency advisory, an outdated nixpkgs input or a zizmor finding.
---

# Fix Vulnerable

The `vulnerable` workflow runs weekly in each forge directory (`.github`, `.gitea`, `.forgejo`, `.tangled`) that exists, with a job for each tool. Tangled runs each tool as its own workflow (`vulnerable-nix`, `vulnerable-go`, `vulnerable-rust`, `vulnerable-npm`, `vulnerable-android`) and skips zizmor, which only audits Actions workflows. Read the failed job's log to find which tool failed and why, then reproduce it locally in the `vulnerable` dev shell, from the job's working directory:

```sh
(cd server && nix develop ..#vulnerable -c govulncheck ./...)
(cd client && nix develop ..#vulnerable -c cargo-audit audit)
(cd web && nix develop ..#vulnerable -c npm audit --omit=dev --audit-level moderate)
(cd android && lockfile=$(mktemp -d)/gradle.lockfile && grep -E '[=,]releaseRuntimeClasspath(,|$)' app/gradle.lockfile > "$lockfile" && nix develop ..#vulnerable -c osv-scanner scan source --lockfile "$lockfile")
nix develop .#vulnerable -c flake-checker -f
nix develop .#vulnerable -c zizmor .github/workflows
```

## govulncheck

The job runs in `server`. govulncheck only fails for vulnerabilities whose vulnerable symbols the code can reach (its "Symbol Results"). Vulnerabilities in packages that are imported or modules that are required, but not called, are only mentioned in its summary and don't fail the job. List them with `-show verbose`, and see the call stacks with `-show traces`.

Each finding names the vulnerable module and the version that fixes it (`Fixed in`). If the module is `stdlib`, follow the standard library steps instead.

Renovate opens automerging fix PRs for vulnerable direct dependencies, so check for an open Renovate PR first. A failure without one is usually an indirect dependency.

1. Update the vulnerable module to the patched version with `go -C server get <module>@<fixed version>`, then run `nix run .#update-go-deps` to tidy `go.mod` and regenerate `gomod2nix.toml`.
2. To see what pulls an indirect module in, run `go -C server mod why -m <module>` for the shortest import path from this project, or `go -C server mod graph` for every module requiring it. If that direct dependency has a release that requires the patched version, prefer updating it with `go -C server get <module>@latest` over keeping an `// indirect` requirement.
3. A finding with no `Fixed in` version can't be fixed by updating. govulncheck has no ignore file or inline comment, so don't try to silence it or mask the job's failure. Tell the user, and remove or replace the vulnerable call or dependency if they agree.

### Standard library

A `stdlib` finding is fixed by a newer Go release, which this project gets from the `go` in nixpkgs, not from `server/go.mod`. The `toolchain` line in `server/go.mod` is kept equal to nixpkgs-unstable's `go` by a Renovate custom manager, and the `go` line is only the minimum language version, so leave it.

1. Run `nix flake update nixpkgs` to get the patched `go`, then check that `nix develop .#vulnerable -c go version` is at or above the fixed version. If nixpkgs-unstable doesn't have it yet, there is nothing to fix until it does, so tell the user.
2. Set the `toolchain` line to that version with `go -C server mod edit -toolchain=go<version>`, never higher than the `go` in nixpkgs, then run `nix run .#update-go-deps`.

## cargo audit

The job runs in `client`, so run the cargo commands there, or pass `--manifest-path client/Cargo.toml`.

Renovate opens automerging fix PRs for vulnerable direct dependencies, so check for an open Renovate PR first. A failure without one is usually a transitive dependency.

1. Update the vulnerable crate within its current requirements with `cargo update -p <crate>`, adding `--precise <version>` to pick the patched version.
2. If a dependent's version requirement rules out the patched version, update the direct dependency that pulls it in, in `Cargo.toml`. Find it with `cargo tree -i <crate>`.
3. Only when no patched version exists, or the vulnerable code is unreachable, ignore the advisory in `client/.cargo/audit.toml`, which cargo-audit reads from the directory it runs in, with a comment explaining why, and tell the user:

   ```toml
   [advisories]
   ignore = [
     "RUSTSEC-0000-0000", # reason
   ]
   ```

Warnings for unmaintained, unsound or yanked crates don't fail the job.

## npm audit

The job runs in `web`, and passes `--omit=dev`, so it only audits the dependencies the web app ships, not dev dependencies such as its build tools.

Renovate opens automerging fix PRs for vulnerable direct dependencies, so check for an open Renovate PR first. A failure without one is usually a transitive dependency.

1. Update the vulnerable package within its current requirements with `npm update <package>`, or fix every advisory that has such an update with `npm audit fix`. Don't use `--force`, which installs semver-major updates of direct dependencies, or downgrades them. Run npm in `web`, or pass `--prefix web`.
2. If a dependent's version requirement rules out the patched version, update the direct dependency that pulls it in, by raising its range in `package.json`, for example with `npm install <package>@<version>`. Find it with `npm explain <package>` or `npm ls <package>`.
3. If no release of the direct dependency allows the patched version, force it with an entry in `package.json` `overrides`, scoped to the dependent that pulls it in, then run `npm install`, and tell the user:

   ```json
   {
     "overrides": {
       "<dependent>": {
         "<package>": "<patched version>"
       }
     }
   }
   ```

4. Only when no patched version exists, or the vulnerable code is unreachable, tell the user. npm audit has no ignore file, so the advisory can't be ignored. Leave `--audit-level` and `--omit=dev` in the workflows as they are unless the user asks to change them.

The flake reads `package-lock.json` directly with `importNpmLock`, so changing the lock file needs no hash update.

## osv-scanner

The job runs in `android`. It scans only the dependencies the release APK ships, the lines of `app/gradle.lockfile` whose configurations include `releaseRuntimeClasspath`, so it skips test dependencies and build tools such as Android lint, Kotlin's compiler and the Android Gradle plugin's own dependencies. Each finding names the vulnerable Maven package and its `FIXED VERSION`.

Renovate opens automerging fix PRs for vulnerable direct dependencies, so check for an open Renovate PR first. A failure without one is usually a transitive dependency.

1. If the package is a direct dependency, raise its version in `android/gradle/libs.versions.toml`.
2. Otherwise find the direct dependency that pulls it in with `gradle -p android :app:dependencyInsight --configuration releaseRuntimeClasspath --dependency <group>:<name>` in the default dev shell, and prefer updating that dependency to a release that requires the patched version.
3. If no release of the direct dependency does, force the patched version with a constraint in `android/app/build.gradle.kts`, and tell the user:

   ```kotlin
   dependencies {
       constraints {
           implementation("<group>:<name>:<patched version>") { because("<advisory ID>") }
       }
   }
   ```

4. After any of these, run `nix run .#update-android-deps` to rewrite the lock files and `android/deps.json`, which the Nix build needs.
5. Only when no patched version exists, or the vulnerable code is unreachable, ignore the advisory in `android/osv-scanner.toml`, with the reason, and tell the user. Pass `--config osv-scanner.toml` to the job's osv-scanner in every forge's workflow, since it looks for the config next to the scanned lock file, which the job writes to a temporary directory:

   ```toml
   [[IgnoredVulns]]
   id = "GHSA-xxxx-xxxx-xxxx"
   reason = "..."
   ```

## flake-checker

flake-checker fails when a nixpkgs input is outdated (over 30 days old), not owned by `NixOS`, or tracks an unsupported branch. Renovate only updates Nix inputs during its weekly lock file maintenance, which automerges.

1. Check for an open Renovate lock file maintenance PR. If one exists, its checks are probably failing, so fix the failure on that PR's branch instead of opening a competing one.
2. Otherwise run `nix flake update`, then fix anything the update breaks.
3. For an owner or branch failure, change the input's URL in `flake.nix` to a supported `NixOS/nixpkgs` branch.

## zizmor

zizmor audits the workflows for security issues. CI runs its online audits with a GitHub token. Locally, set `GH_TOKEN` to enable them, for example with `GH_TOKEN=$(gh auth token)`, or they're skipped.

1. Fix the finding in every forge directory that has the workflow, not only the one that failed. The forges' workflows differ: GitHub and Gitea set `permissions`, while Forgejo doesn't support them and uses `enable-openid-connect`. Adapt the fix to each forge rather than copying it.
2. Pinned action updates (`unpinned-uses`, known vulnerable versions) are Renovate's job, so check for an open Renovate PR first.
3. Only when a finding doesn't apply to a forge, ignore it with an inline `# zizmor: ignore[<audit>]` comment on the flagged line, as the Forgejo workflows do for `excessive-permissions`.

## Finish

Run `nix flake check`, and confirm each tool that failed now passes locally.
