{
  buildNpmPackage,
  importNpmLock,
  nodejs_24,
  oxfmt,
  oxlint,
}:
buildNpmPackage (final: {
  pname = "trevstack-web";
  version = "1.2.0";

  src = ./.;
  nodejs = nodejs_24;
  npmConfigHook = importNpmLock.npmConfigHook;
  npmDeps = importNpmLock {
    npmRoot = final.src;
  };

  nativeCheckInputs = [
    oxfmt
    oxlint
  ];
  checkPhase = ''
    runHook preCheck
    oxfmt --check
    oxlint --deny-warnings
    CI=true NO_COLOR=1 npm test
    runHook postCheck
  '';

  installPhase = ''
    runHook preInstall
    cp -r dist "$out"
    runHook postInstall
  '';
})
