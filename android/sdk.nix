# The Android SDK that builds the app, matching compileSdk and buildToolsVersion in app/build.gradle.kts
{ androidenv }:
(androidenv.composeAndroidPackages {
  platformVersions = [ "37.2" ];
  buildToolsVersions = [ "37.0.0" ];
  includeEmulator = false;
  includeSystemImages = false;
}).androidsdk
