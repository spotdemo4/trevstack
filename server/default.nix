{
  mkGoModule,
  lib,
  web,
}:
mkGoModule (final: {
  pname = "trevstack-server";
  version = "3.2.0";

  src = ./.;
  goSum = ./go.sum;
  modules = ./gomod2nix.toml;

  tags = [ "prod" ];
  ldflags = [ "-X main.version=${final.version}" ];

  postConfigure = ''
    cp -r ${web} web
  '';

  doCheck = true;
  checkPhase = ''
    runHook preCheck
    go test ./...
    go test -tags=prod ./...
    runHook postCheck
  '';

  meta = {
    mainProgram = "server";
    description = "full-stack template";
    license = lib.licenses.mit;
    platforms = lib.platforms.all;
    homepage = "https://trev.zip/template/stack";
    downloadPage = "https://trev.zip/template/stack/releases";
    changelog = "https://trev.zip/template/stack/releases/tag/server/v${final.version}";
    sourceProvenance = [ lib.sourceTypes.fromSource ];
    maintainers = [ lib.maintainers.spotdemo4 ];
  };
})
