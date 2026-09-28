{
  mkGoModule,
  docs,
  go-tools,
  lib,
  web,
}:
mkGoModule (final: {
  pname = "trevstack-server";
  version = "1.0.3";

  src = ./.;
  goSum = ./go.sum;
  vendorHash = "sha256-EAuHMjk0E6eIyE6BIxF5vT43iu5WgWFr0YEKinWdzec=";

  postConfigure = ''
    cp -r ${docs} docs
    cp -r ${web} web
  '';

  doCheck = true;
  nativeCheckInputs = [ go-tools ];
  checkPhase = ''
    runHook preCheck
    go test ./...
    go test -tags=dev ./...
    runHook postCheck
  '';

  meta = {
    mainProgram = "server";
    description = "full-stack template";
    license = lib.licenses.mit;
    platforms = lib.platforms.all;
    homepage = "https://trev.zip/template/stack";
    changelog = "https://trev.zip/template/stack/releases";
    downloadPage = "https://trev.zip/template/stack/releases/tag/v${final.version}";
  };
})
