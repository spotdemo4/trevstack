{
  description = "Trevstack";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = {nixpkgs, ...}: let
    build-systems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
    forSystem = f:
      nixpkgs.lib.genAttrs build-systems (
        system:
          f {
            inherit system;
            pkgs = import nixpkgs {
              inherit system;
            };
          }
      );
  in {
    devShells = forSystem ({pkgs, ...}: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          git

          # Nix
          nix-update
          alejandra
        ];
      };

      ci = pkgs.mkShell {
        packages = with pkgs; [
          git
          renovate
          nix-update
        ];
      };
    });

    formatter = forSystem ({pkgs, ...}: pkgs.alejandra);
  };
}
