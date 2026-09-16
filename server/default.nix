{
  buildGoModule,
  docs,
  go-tools,
  lib,
  web,
}:
buildGoModule (final: {
  pname = "trevstack-server";
  version = "0.16.0";

  src = ./.;
  goSum = ./go.sum;
  proxyVendor = true;
  vendorHash = "sha256-NvnvZT80lmjQl6uvSxKGNisO8C1GAkfuXHkbUrTeevk=";

  postConfigure = ''
    cp -r ${docs} docs
    cp -r ${web} web
  '';

  nativeCheckInputs = [
    go-tools
  ];
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
    homepage = "https://trev.zip/llc/stack";
    changelog = "https://trev.zip/llc/stack/releases";
    downloadPage = "https://trev.zip/llc/stack/releases/tag/v${final.version}";
  };
})
