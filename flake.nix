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
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ self.overlay naersk.overlay rust-overlay.overlay ];
          };
        in
        rec {
          packages = pkgs.nickcao.fn;
          checks = packages;
        }
      ) //
    {
      overlay = final: prev:
        let
          toolchain = final.rust-bin.nightly.latest.default;
          naersk = final.naersk.override { cargo = toolchain; rustc = toolchain; };
        in
        {
          nickcao.fn = rec {
            meow = naersk.buildPackage {
              src = ./meow;
            };
            meow-image = final.dockerTools.buildLayeredImage {
              name = "gitlab.com/nickcao/meow";
              contents = [ final.cacert ];
              config.Entrypoint = [ "${meow}/bin/meow" ];
            };
          };
        };
    };
}
