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
        nlib = with pkgs.rust-nightly.default;
          naersk.lib.${system}.override {
            rustc = rustc;
            cargo = cargo;
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
