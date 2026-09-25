{
  clippy,
  lib,
  rustPlatform,
}:
rustPlatform.buildRustPackage (final: {
  pname = "trevstack-client";
  version = "1.0.1";

  src = ./.;
  cargoLock.lockFile = ./Cargo.lock;

  nativeCheckInputs = [
    clippy
  ];
  checkPhase = ''
    runHook preCheck
    cargo test --offline
    cargo clippy --offline --all-targets -- -D warnings
    runHook postCheck
  '';

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
