{
  description = "GG - terminal games in Go";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

    templ.url = "github:a-h/templ";
    templ.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = {
    self,
    nixpkgs,
    ...
  } @ inputs: let
    version = builtins.substring 0 8 (self.lastModifiedDate or self.lastModified or "19700101");
    supportedSystems = ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];

    templ = system: inputs.templ.packages.${system}.templ;

    # See: https://ayats.org/blog/no-flake-utils
    forAllSystems = f:
      nixpkgs.lib.genAttrs supportedSystems (
        system:
          f {
            inherit system;
            pkgs = nixpkgs.legacyPackages.${system};
          }
      );

    gg = forAllSystems ({
      system,
      pkgs,
      ...
    }: {
      default = pkgs.buildGoModule {
        inherit version;
        pname = "gg";
        src = ./.;
        doCheck = true;

        preBuild = ''
          ${templ system}/bin/templ generate
        '';

        # vendorHash = pkgs.lib.fakeHash;
        vendorHash = "sha256-BhMv2rA80lHxmUqot5qH0soeVjIGV+jNdUQEtDU40MQ=";

        checkPhase = ''
          go test ./...
        '';
      };
    });
  in {
    formatter = forAllSystems ({
      pkgs,
      system,
    }:
      pkgs.alejandra);

    checks = forAllSystems ({
      pkgs,
      system,
    }: {
      fmt =
        pkgs.runCommand "gg-fmt-check" {
          nativeBuildInputs = [pkgs.go];
        } ''
          cd ${./.}
          if [ -n "$(gofmt -l .)" ]; then
            echo "Code is not gofmt'd:"
            gofmt -l .
            exit 1
          fi
          touch $out
        '';
    });

    # Provide some binary packages for selected system types.
    packages = forAllSystems ({
      pkgs,
      system,
    }:
      gg.${system});

    devShells = forAllSystems ({
      pkgs,
      system,
    }: {
      default = pkgs.mkShell {
        buildInputs = [
          pkgs.go
          pkgs.gopls
          pkgs.gotools
          pkgs.go-tools
          (templ system)
        ];

        shellHook = ''
          export GOPROXY=https://proxy.golang.org,direct
          export GOSUMDB=sum.golang.org
        '';
      };
    });
  };
}
