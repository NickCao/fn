{
  inputs = {
    nixpkgs.url = "github:NickCao/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };
  outputs = { self, nixpkgs, flake-utils, rust-overlay }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = import nixpkgs { inherit system; overlays = [ self.overlay rust-overlay.overlay ]; }; in
        rec {
          packages = { inherit (pkgs) meow woff bark quark; };
          checks = packages;
          devShell = pkgs.mkShell { inputsFrom = builtins.attrValues packages; };
        }
      ) //
    {
      overlay = final: prev:
        let
          toolchain = final.rust-bin.nightly.latest.default;
          platform = final.makeRustPlatform { cargo = toolchain; rustc = toolchain; };
        in
        {
          meow = platform.buildRustPackage {
            name = "meow";
            src = ./meow;
            nativeBuildInputs = [ final.pkg-config ];
            buildInputs = [ final.openssl ];
            cargoLock = {
              lockFile = ./meow/Cargo.lock;
            };
          };
          woff = final.buildGoModule {
            name = "woff";
            src = ./woff;
            vendorSha256 = "sha256-JndbBs8D1kMvOnHPRLk2nmVd9KNH964BGcDUu+49anU=";
          };
          bark = platform.buildRustPackage {
            name = "bark";
            src = ./bark;
            nativeBuildInputs = [ final.pkg-config ];
            buildInputs = [ final.openssl ];
            cargoLock = {
              lockFile = ./bark/Cargo.lock;
            };
          };
          quark = final.buildGoModule {
            name = "quark";
            src = ./quark;
            vendorSha256 = "sha256-2tZS03xt/IrjBKDSfUK6WT+l2I6Lyj6IYH2cuzhqwwY=";
          };
        };
    };
}
