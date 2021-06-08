{
  inputs = {
    nixpkgs.url = "github:NickCao/nixpkgs/nixos-unstable";
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
          packages = { inherit (pkgs) meow woff; };
          checks = packages // pkgs.lib.mapAttrs' (k: v: pkgs.lib.nameValuePair "${k}-image" v.image) packages;
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
            passthru = {
              image = final.dockerTools.buildLayeredImage {
                name = "gitlab.com/nickcao/meow";
                contents = [ final.cacert ];
                config.Entrypoint = [ "${final.meow}/bin/meow" ];
              };
            };
          };
          woff = final.buildGoModule
            {
              name = "woff";
              src = ./woff;
              vendorSha256 = "sha256-br1k0TLegGnDkUk8p8cybjHkLAo/oJcvNGpG/ndbhLA=";
              passthru = {
                image = final.dockerTools.buildLayeredImage {
                  name = "gitlab.com/nickcao/woff";
                  contents = [ final.cacert ];
                  config.Entrypoint = [ "${final.woff}/bin/woff" ];
                };
              };
            };
        };
    };
}
