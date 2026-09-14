{
  buildNpmPackage,
  importNpmLock,
  lib,
  nodejs_24,
  oxfmt,
  oxlint,
}:
buildNpmPackage (final: {
  pname = "trevstack-web";
  version = "0.13.1";

  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./.npmrc
      ./.oxfmtrc.json
      ./.oxlintrc.json
      ./package-lock.json
      ./package.json
      ./tsconfig.json
      ./vite.config.ts
      ./index.css
      ./index.html
      ./index.tsx
      ./connect
      ./layout
      ./lib
      ./public
      ./routes
    ];
  };
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
