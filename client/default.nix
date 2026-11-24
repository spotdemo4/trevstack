{
  clippy,
  lib,
  trev,
}:
trev.libs.mkRustPackage (final: {
  pname = "trevstack-client";
  version = "4.3.2";

  src = ./.;
  cargoLock.lockFile = ./Cargo.lock;

  # nixpkgs' cargoSetupHook makes crates such as zstd-sys link system libraries
  # through pkg-config, keep building the vendored ones, which cross-compile and
  # link statically
  dontCargoUseSystemLibs = true;

  doCheck = true;
  checkPhase = ''
    runHook preCheck
    cargo test --offline
    runHook postCheck
  '';

  # also check the dependencies for clippy, the lints run in the client-lint flake check
  cargoArtifactsArgs = {
    nativeCheckInputs = [ clippy ];
    doCheck = true;
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
    downloadPage = "https://trev.zip/template/stack/releases";
    changelog = "https://trev.zip/template/stack/releases/tag/client/v${final.version}";
    sourceProvenance = [ lib.sourceTypes.fromSource ];
    maintainers = [ lib.maintainers.spotdemo4 ];
  };
})
