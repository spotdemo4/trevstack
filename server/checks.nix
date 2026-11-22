{
  pkgs,
  self,
}:
{
  lint = {
    src = self.packages.${pkgs.stdenv.hostPlatform.system}.server;
    packages = with pkgs; [ golangci-lint ];
    script = ''
      golangci-lint run ./...
      golangci-lint run --build-tags=prod ./...
    '';
  };
}
