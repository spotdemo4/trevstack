{
  buildNpmPackage,
  importNpmLock,
  nodejs_24,
}:
buildNpmPackage (final: {
  pname = "trevstack-docs";
  version = "1.0.15";

  src = ./.;
  nodejs = nodejs_24;
  npmConfigHook = importNpmLock.npmConfigHook;
  npmDeps = importNpmLock {
    npmRoot = final.src;
  };

  installPhase = ''
    runHook preInstall
    cp -r dist "$out"
    runHook postInstall
  '';
})
