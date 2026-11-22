{
  pkgs,
  self,
}:
{
  lint = {
    src = self.packages.${pkgs.stdenv.hostPlatform.system}.web;
    packages = with pkgs; [
      oxfmt
      oxlint
    ];
    script = ''
      oxfmt --check
      oxlint --deny-warnings
    '';
  };
}
