{
  description = "go flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
  };

  outputs =
    {
      nixpkgs,
      ...
    }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
        config.allowUnfree = true;
      };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        hardeningDisable = [ "fortify" ];
        packages = with pkgs; [
          go
          gopls
          golangci-lint
          gotools
          gofumpt
          delve
        ];
      };
    };
}
