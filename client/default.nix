{
  clippy,
  lib,
  mkRustPackage,
}:
mkRustPackage (final: {
  pname = "trevstack-client";
  version = "2.2.0";

  src = ./.;
  cargoLock.lockFile = ./Cargo.lock;

  doCheck = true;
  checkPhase = ''
    runHook preCheck
    cargo test --offline
    runHook postCheck
  '';

  # also check the dependencies for clippy, the lints run in the clippy flake check
  cargoArtifactsArgs = {
    nativeCheckInputs = [ clippy ];
    checkPhase = ''
      runHook preCheck
      cargo test --offline
      cargo clippy --offline --all-targets
      runHook postCheck
    '';
  };

  doInstallCheck = true;
  installCheckPhase = ''
    runHook preInstallCheck
    test -x "$out/bin/client"
    "$out/bin/client" --help >/dev/null
    runHook postInstallCheck
  '';

  meta = {
    mainProgram = "client";
    description = "full-stack template";
    license = lib.licenses.mit;
    platforms = lib.platforms.all;
    homepage = "https://trev.zip/template/stack";
    changelog = "https://trev.zip/template/stack/releases";
    downloadPage = "https://trev.zip/template/stack/releases/tag/v${final.version}";
  };
})
