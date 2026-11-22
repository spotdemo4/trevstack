{
  callPackage,
  gradle_9,
  jdk25,
  lib,
  runtimeShell,
  stdenvNoCC,
}:
let
  androidSdk = callPackage ./sdk.nix { };
  androidHome = "${androidSdk}/libexec/android-sdk";
in
stdenvNoCC.mkDerivation (final: {
  pname = "trevstack-android";
  version = "0.1.0";

  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./app
      ./build.gradle.kts
      ./gradle.properties
      ./gradle/libs.versions.toml
      ./settings-gradle.lockfile
      ./settings.gradle.kts
    ];
  };

  nativeBuildInputs = [
    gradle_9
    jdk25
  ];

  # the Maven artifacts the build downloads; nix run .#update-android-deps rewrites deps.json
  mitmCache = gradle_9.fetchDeps {
    pkg = final.finalPackage;
    data = ./deps.json;
    bwrapFlags = ''--ro-bind "$PWD" "$PWD" --dir /bin --symlink ${runtimeShell} /bin/sh'';
  };
  __darwinAllowLocalNetworking = true;

  gradleFlags = [
    "-Dorg.gradle.java.home=${jdk25.home}"
    # the aapt2 that the Android Gradle plugin downloads doesn't run on NixOS
    "-Dorg.gradle.project.android.aapt2FromMavenOverride=${androidHome}/build-tools/37.0.0/aapt2"
    # both caches start empty in every build
    "--no-build-cache"
    "--no-configuration-cache"
  ];
  # one Gradle run builds, lints and tests, since each run configures the build from nothing
  gradleBuildTask = [
    ":app:assembleRelease"
    ":app:lintRelease"
    ":app:testDebugUnitTest"
  ];
  # runs the build itself, since nixDownloadDeps can't resolve every Android configuration
  gradleUpdateScript = ''
    runHook preBuild
    gradle ${lib.escapeShellArgs final.gradleBuildTask}
  '';

  preBuild = ''
    export ANDROID_HOME=${androidHome}
    # the Android Gradle plugin writes its analytics settings here
    export ANDROID_USER_HOME="$TMPDIR/android"
  '';

  # the tests ran with the build
  doCheck = true;
  dontUseGradleCheck = true;

  installPhase = ''
    runHook preInstall
    # unsigned, since signing keys don't belong in the store; sign it with apksigner to install it
    install -Dm644 app/build/outputs/apk/release/app-release-unsigned.apk \
      "$out/share/trevstack/trevstack-${final.version}-unsigned.apk"
    runHook postInstall
  '';

  passthru = {
    inherit androidSdk;
  };

  meta = {
    description = "full-stack template";
    license = lib.licenses.mit;
    # the only platform it's tested on
    platforms = [ "x86_64-linux" ];
    homepage = "https://trev.zip/template/stack";
    sourceProvenance = with lib.sourceTypes; [
      fromSource
      # Maven dependencies, and Cronet's native library
      binaryBytecode
      binaryNativeCode
    ];
    maintainers = [ lib.maintainers.spotdemo4 ];
  };
})
