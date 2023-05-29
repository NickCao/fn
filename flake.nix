{
  inputs = {
    nixpkgs.url = "github:NickCao/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = import nixpkgs { inherit system; overlays = [ self.overlays.default ]; }; in
        rec {
          packages = { inherit (pkgs) meow; };
          checks = packages;
          devShells.default = pkgs.mkShell { inputsFrom = builtins.attrValues packages; };
        }
      ) //
    {
      overlays.default = final: prev: {
        meow = final.rustPlatform.buildRustPackage {
          name = "meow";
          src = ./meow;
          cargoLock = {
            lockFile = ./meow/Cargo.lock;
          };
        };
      };
    };
}
