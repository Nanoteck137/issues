{
  description = "issues";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";

    versionctl.url = "github:nanoteck137/versionctl/0.3.0";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        version = pkgs.lib.strings.fileContents "${self}/version";
        fullVersion = ''${version}-${self.dirtyShortRev or self.shortRev or "dirty"}'';

        app = pkgs.buildGoModule {
          pname = "issues";
          version = fullVersion;
          src = ./.;

          ldflags = [
            "-X github.com/nanoteck137/issues.Version=${version}"
            "-X github.com/nanoteck137/issues.Commit=${self.dirtyRev or self.rev or "no-commit"}"
          ];

          vendorHash = null;
        };
      in
      {
        packages = {
          default = app;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            air
            go
            gopls
            just

            inputs.versionctl.packages.${system}.default
          ];
        };
      }
    );
}
