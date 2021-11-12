with import <nixpkgs> {};

buildGoPackage {
  name = "switch-listing";

  src = ./.;

  goPackagePath = "github.com/juliosueiras/switch-listing";
}
