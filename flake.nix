{
  description = "GG - terminal games in Go";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
    ...
  }: let
    version = builtins.substring 0 8 (self.lastModifiedDate or self.lastModified or "19700101");
    supportedSystems = ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];

    # See: https://ayats.org/blog/no-flake-utils
    forAllSystems = f:
      nixpkgs.lib.genAttrs supportedSystems (
        system: f nixpkgs.legacyPackages.${system}
      );

    gg = pkgs:
      pkgs.buildGoModule {
        inherit version;
        pname = "gg";
        src = ./.;
        doCheck = true;

        # vendorHash = pkgs.lib.fakeHash;
        vendorHash = "sha256-BhMv2rA80lHxmUqot5qH0soeVjIGV+jNdUQEtDU40MQ=";

        checkPhase = ''
          go test ./...
        '';
      };
  in {
    formatter = forAllSystems (pkgs: pkgs.alejandra);

    checks = forAllSystems (pkgs: {
      fmt =
        pkgs.runCommand "gg-check-fmt" {
          nativeBuildInputs = [pkgs.go];
        } ''
          cd ${./.}
          not_formatted="$(gofmt -l .)"
          if [ -n "$not_formatted" ]; then
            echo "Code is not gofmt'd:"
            echo "$not_formatted"
            exit 1
          fi

          touch $out
        '';
    });

    # Provide some binary packages for selected system types.
    packages = forAllSystems (pkgs: {
      default = gg pkgs;
    });

    devShells = forAllSystems (pkgs: {
      default = pkgs.mkShell {
        buildInputs = [
          pkgs.go
          pkgs.gopls
        ];

        shellHook = ''
          export GOPROXY=https://proxy.golang.org,direct
          export GOSUMDB=sum.golang.org
        '';
      };
    });
  };
}
