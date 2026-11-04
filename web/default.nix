{
  lib,
  buildNpmPackage,
  importNpmLock,
  nodejs_26,
}:
buildNpmPackage (final: {
  pname = "trevstack-web";
  version = "1.6.10";

  # the root oxlint config that ours extends lives outside this directory
  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../.oxlintrc.json
      ./.
    ];
  };
  sourceRoot = "${final.src.name}/web";
  nodejs = nodejs_26;
  npmConfigHook = importNpmLock.npmConfigHook;
  npmDeps = importNpmLock {
    npmRoot = ./.;
  };

  doCheck = true;
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
