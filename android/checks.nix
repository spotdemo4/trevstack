{
  ktlint,
}:
{
  lint = {
    root = ./.;
    filter = file: file.hasExt "kt" || file.hasExt "kts";
    include = [ ./.editorconfig ];
    packages = [ ktlint ];
    # the Android lint checks run when the package builds
    script = ''
      ktlint --relative "**/*.kt" "**/*.kts" "!app/src/gen/**"
    '';
  };
}
