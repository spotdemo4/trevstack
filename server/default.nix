{
  buildGoModule,
  docs,
  go-tools,
  lib,
  web,
}:
buildGoModule (final: {
  pname = "trevstack-server";
  version = "0.15.0";

  src = ./.;
  goSum = ./go.sum;
  proxyVendor = true;
  vendorHash = "sha256-ywh8KpD5fE0oaSge7zwa0xvRaE3BShRWXhoS8hPtCcU=";

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
    go vet ./...
    staticcheck ./...
    go fix -diff ./...
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
