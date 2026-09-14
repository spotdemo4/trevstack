{
  buildGoModule,
  go-tools,
  lib,
  web,
}:
buildGoModule (final: {
  pname = "trevstack-server";
  version = "0.13.1";

  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./go.mod
      ./go.sum
      (lib.fileset.fileFilter (file: file.hasExt "go" || file.hasExt "sql" || file.hasExt "yaml") ./.)
    ];
  };
  goSum = ./go.sum;
  proxyVendor = true;
  vendorHash = "sha256-tO89XUOF+MclDs01ynXNJaErHquEVeNeqTVdtasAj7k=";

  postConfigure = ''
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
