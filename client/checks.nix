{
  pkgs,
  self,
}:
{
  lint = {
    src = self.packages.${pkgs.stdenv.hostPlatform.system}.client;
    packages = with pkgs; [
      rustfmt
      clippy
    ];
    script = ''
      cargo fmt --check
      cargo clippy --offline --all-targets -- -D warnings
    '';
  };
}
