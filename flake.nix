{
  inputs = {
    nixpkgs.url = "github:NickCao/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = import nixpkgs { inherit system; overlays = [ self.overlays.default ]; }; in
        rec {
          packages = { inherit (pkgs) meow quark sirius serve; };
          checks = packages;
          devShells.default = pkgs.mkShell { inputsFrom = builtins.attrValues packages; };
        }
      ) //
    {
      overlays.default = final: prev:
        {
          meow = final.rustPlatform.buildRustPackage {
            name = "meow";
            src = ./meow;
            cargoLock = {
              lockFile = ./meow/Cargo.lock;
            };
          };
          quark = final.buildGoModule {
            name = "quark";
            src = ./quark;
            vendorSha256 = "sha256-2tZS03xt/IrjBKDSfUK6WT+l2I6Lyj6IYH2cuzhqwwY=";
          };
          sirius = final.buildGoModule {
            name = "sirius";
            src = ./sirius;
            vendorSha256 = "sha256-+/dltb04n/s5E6lkH2HlllQu5rihQQScBHZSDWwLyxY=";
          };
          serve = final.buildGoModule {
            name = "serve";
            src = ./serve;
            vendorSha256 = null;
          };
        };
    };
}
