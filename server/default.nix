{
  mkGoModule,
  docs,
  lib,
  web,
}:
mkGoModule (final: {
  pname = "trevstack-server";
  version = "1.2.0";

  src = ./.;
  goSum = ./go.sum;
  vendorHash = "sha256-P0LuUP2K/1Zw2Sv9GQZVEpDb/HgYeS1wJRpDVxDHdJU=";

  ldflags = [ "-X main.version=${final.version}" ];

  postConfigure = ''
    cp -r ${docs} docs
    cp -r ${web} web
  '';

  doCheck = true;
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
