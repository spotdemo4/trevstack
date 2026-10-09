{
  buildNpmPackage,
  importNpmLock,
  nodejs_24,
}:
buildNpmPackage (final: {
  pname = "trevstack-web";
  version = "1.5.0";

  src = ./.;
  nodejs = nodejs_24;
  npmConfigHook = importNpmLock.npmConfigHook;
  npmDeps = importNpmLock {
    npmRoot = final.src;
  };

  checkPhase = ''
    runHook preCheck
    CI=true NO_COLOR=1 npm test
    runHook postCheck
  '';

  installPhase = ''
    runHook preInstall
    cp -r dist "$out"
    runHook postInstall
  '';
})
