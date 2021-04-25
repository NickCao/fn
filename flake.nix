{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable-small";
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
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ rust-overlay.overlay ];
        };
        toolchain = pkgs.rust-bin.nightly.latest.default.override {
          targets = [ "wasm32-wasi" ];
        };
        nlib = naersk.lib.${system}.override {
          rustc = toolchain;
          cargo = toolchain;
        };
        buildCrate =
          { name
          , nativeBuildInputs ? [ ]
          , buildInputs ? [ ]
          }: nlib.buildPackage {
            inherit nativeBuildInputs buildInputs;
            pname = name;
            root = ./.;
            cargoBuildOptions = opts: opts ++ [ "-p" name ];
            cargoTestOptions = opts: opts ++ [ "-p" name ];
          };
        pushImage = image: ''
          ${pkgs.skopeo}/bin/skopeo copy --insecure-policy docker-archive:${image} docker://${image.imageName}
        '';
      in
      with pkgs; rec {
        devShell = pkgs.mkShell { inputsFrom = [ defaultPackage ]; };
        defaultPackage = packages.meow;

        packages = {
          meow = buildCrate {
            name = "meow";
          };
          image = {
            meow = dockerTools.buildLayeredImage {
              name = "quay.io/nickcao/meow";
              contents = [ cacert ];
              config.Entrypoint = [ "${packages.meow}/bin/meow" ];
            };
          };
          push = writeShellScriptBin "push" ''
            mkdir -p /var/tmp
            ${pushImage packages.image.meow}
          '';
        };
      }
    );
}
