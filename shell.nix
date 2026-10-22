{ pkgs ? import <nixpkgs> {} }:

let unstable = import (builtins.fetchTarball {
  url = "https://github.com/NixOS/nixpkgs/archive/nixos-unstable.tar.gz";
}) { };

in pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gcc
    glibc.dev
    air
    sqlc
    pgloader
    sqlitebrowser
    goose
    # Use postgresql_17 from unstable overlay
    (unstable.postgresql_17)
  ];

  shellHook = ''
    export CGO_ENABLED=1
    export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
  '';
}
