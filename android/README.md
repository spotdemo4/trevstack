# android

The TrevStack Android app is a client for the server's cursor service, like the
[command-line client](../client/README.md). It signs in, then shows a pad with one mode for each
kind of RPC: tap or drag on it to send your position, and see everyone else's cursors move. It talks
to the server with [TrevRPC](https://trev.zip/llc/TrevRPC) over HTTP/3 through
[Cronet](https://developer.android.com/develop/connectivity/cronet), which falls back to HTTP/2 when
UDP is blocked.

## usage

Sign in with an existing account, or create one with **Sign up**. The server URL must use `https`
and name only a host and an optional port (`443` by default).

| mode    | RPC                              | touching the pad                                                 |
| ------- | -------------------------------- | ---------------------------------------------------------------- |
| `Send`  | `Send`, unary                    | records the tapped position and shows its stream ID              |
| `Watch` | `Watch`, server streaming        | nothing; shows every user's cursor as it moves                   |
| `Track` | `Track`, client streaming        | streams each drag as a stream of its own, then shows its summary |
| `Share` | `Share`, bidirectional streaming | streams your position while it shows every other stream's cursor |

Positions are fractions of the pad's width and height, from `0` to `1`, measured from its top-left
corner, so a position shows in the same place on every client whatever its size. Other streams'
cursors are coloured by username, as in the command-line client, and disappear when their stream
ends or after 10 seconds without moving. Switching modes ends the last mode's calls.

The app sends the session token from `AuthService.Login` as `authorization: Bearer <token>`
connection metadata on every call, and calls `AuthService.Logout` when you sign out or leave the app
with back. It doesn't store the token, so you sign in again whenever it starts. Signing in, signing out
and `Send` time out after 10 seconds; the streams last for as long as their mode is selected.

### install

`nix build .#android` builds an unsigned release APK in `result/share/trevstack/`. Sign it with your
own key, using `apksigner` from the Android SDK's build tools, before installing it:

```sh
apksigner sign --ks release.jks --out trevstack.apk result/share/trevstack/trevstack-*-unsigned.apk
adb install trevstack.apk
```

## development

Inside the development shell, which has Gradle, a JDK and the Android SDK, from this directory:

```sh
gradle installDebug                 # build and install the debug app on a device or emulator
gradle test                         # unit tests
gradle lint                         # Android lint
```

Android Studio can open this directory too; the wrapper (`./gradlew`) downloads the same Gradle.
`ktlint` checks the Kotlin style, and `nix fmt` fixes it.

The app's code depends on code that `buf generate` writes to `app/src/gen/` from the protos in
[`../proto`](../proto): the protobuf-java messages, and a TrevRPC client class for each service.

### local server

To use a server on your computer from an emulator or a USB-connected device, forward its port, run
the server, and sign in to `https://localhost:8080`, the URL debug builds start with:

```sh
adb reverse tcp:8080 tcp:8080
nix run .#server
```

`adb reverse` forwards TCP only, so the app talks to the server over HTTP/2.

The server's [development certificate](../server/README.md#tls) is self-signed, so debug builds
trust the copy the server writes to `trevstack/development.pem` in the user configuration directory,
as the command-line client does. The build reads it when it runs, so start the server once before
building, and build again whenever the server restarts or replaces it, every 6 days. Pass
`-PstackCaCert=FILE` to trust another certificate instead. Release builds trust only the system's
certificates.

Pass `-PstackUrl=https://example.com` to start the sign-in screen with another server.

### dependencies

The versions of the app's dependencies are in [`gradle/libs.versions.toml`](gradle/libs.versions.toml).
After changing them, update the lock files (`app/gradle.lockfile` and `settings-gradle.lockfile`), which
pin every resolved version, and [`deps.json`](deps.json), which pins every artifact the Nix build
downloads:

```sh
nix run .#update-android-deps
```

Keep `trevrpc` at the version of the flake's `trevrpc` input, since TrevRPC packages are compatible
only with those of the same version.

### layout

| path                                              | description                                                    |
| ------------------------------------------------- | -------------------------------------------------------------- |
| `app/src/main/kotlin/zip/trev/template/stack/`    | the app                                                        |
| &emsp;`MainActivity.kt`                           | entry point, which shows the sign-in screen or the pad         |
| &emsp;`StackViewModel.kt`                         | signing in and out, and what each mode does                    |
| &emsp;`StackClient.kt`                            | the Cronet engine, the TrevRPC channel, and the calls          |
| &emsp;`Cursors.kt`                                | positions, cursor events, and the cursors on the pad           |
| &emsp;`ServiceUrl.kt`                             | server URL validation                                          |
| &emsp;`ui/`                                       | the sign-in screen, the pad, and the theme                     |
| `app/src/test/`                                   | unit tests                                                     |
| `app/src/gen/`                                    | code generated by `buf generate`; do not edit                  |
| `app/build.gradle.kts`                            | the app's build, and the debug build's development certificate |
| `default.nix`, `sdk.nix`                          | the Nix build and its Android SDK                              |
| `deps.json`                                       | the Maven artifacts of the Nix build                           |
| `app/gradle.lockfile`, `settings-gradle.lockfile` | the resolved version of every dependency                       |
