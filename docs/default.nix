{
  buildNpmPackage,
  importNpmLock,
  nodejs_24,
  oxfmt,
  oxlint,
}:
buildNpmPackage (final: {
  pname = "trevstack-docs";
  version = "0.17.1";

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
    runHook postCheck
  '';

  installPhase = ''
    runHook preInstall
    cp -r dist "$out"
    runHook postInstall
  '';
})
