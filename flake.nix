{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
    fenix = {
      url = "github:nix-community/fenix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    naersk = {
      url = "github:nmattia/naersk";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
  outputs = { self, nixpkgs, flake-utils, fenix, naersk }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ fenix.overlay ];
        };
        toolchain = pkgs.rust-nightly.default;
        nlib = naersk.lib."${system}".override {
          rustc = toolchain.rustc;
          cargo = toolchain.cargo;
        };
      in
      rec {
        devShell = pkgs.mkShell { inputsFrom = [ defaultPackage ]; };
        defaultPackage = packages.fn;
        packages.fn = with pkgs;nlib.buildPackage {
          pname = "fn";
          root = ./.;
          buildInputs = [ openssl ];
          nativeBuildInputs = [ pkg-config ];
        };
        packages.image.meow = pkgs.dockerTools.buildLayeredImage {
          name = "meow";
          contents = [ pkgs.cacert ];
          config.Entrypoint = [ "${defaultPackage}/bin/meow" ];
        };
      }
    );
}
