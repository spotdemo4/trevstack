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

  nativeCheckInputs = [ go-tools ];
  checkPhase = ''
    runHook preCheck
    export HOME=$(mktemp -d)

    go test ./...
    go test -tags=dev ./...

    go vet ./...
    go vet -tags=dev ./...

    staticcheck ./...
    staticcheck -tags=dev ./...

    go fix -diff ./...
    go fix -diff -tags=dev ./...

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
