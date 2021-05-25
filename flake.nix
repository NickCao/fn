{
  inputs = {
    nixpkgs.url = "github:NickCao/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
    naersk = {
      url = "github:nmattia/naersk";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
  outputs = { self, nixpkgs, flake-utils, rust-overlay, naersk }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = import nixpkgs { inherit system; overlays = [ self.overlay rust-overlay.overlay naersk.overlay ]; }; in
        rec {
          packages = { inherit (pkgs) meow; };
          checks = packages // pkgs.lib.mapAttrs' (k: v: pkgs.lib.nameValuePair "${k}-image" v.image) packages;
          devShell = pkgs.mkShell { inputsFrom = builtins.attrValues packages; };
        }
      ) //
    {
      overlay = final: prev:
        let
          toolchain = final.rust-bin.nightly.latest.default;
          naersk = final.naersk.override { cargo = toolchain; rustc = toolchain; };
        in
        {
          meow = naersk.buildPackage {
            src = ./meow;
            passthru = {
              image = final.dockerTools.buildLayeredImage {
                name = "gitlab.com/nickcao/meow";
                contents = [ final.cacert ];
                config.Entrypoint = [ "${final.meow}/bin/meow" ];
              };
            };
          };
        };
    };
}
