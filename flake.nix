{
  description = "GG - terminal games in Go";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      version = builtins.substring 0 8 (self.lastModifiedDate or self.lastModified or "19700101");
      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      # See: https://ayats.org/blog/no-flake-utils
      forAllSystems =
        f: nixpkgs.lib.genAttrs supportedSystems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      formatter = forAllSystems (
        pkgs:
        pkgs.treefmt.withConfig {
          runtimeInputs = [
            pkgs.nixfmt-rfc-style
            pkgs.go
          ];

          settings.no-cache = true; # Disable cache to avoid uneccessary work
          settings.on-unmatched = "info"; # Log level for files treefmt won't format
          settings.formatter.nixfmt = {
            command = "nixfmt";
            includes = [ "*.nix" ];
          };
          settings.formatter.gofmt = {
            command = "gofmt";
            excludes = [ "vendor/*" ];
            includes = [ "*.go" ];
            options = [ "-w" ];
          };
        }
      );

      # Provide some binary packages for selected system types.
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
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
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          env = {
            "GOROOT" = "${pkgs.go}/share/go";
            "GOPROXY" = "https://proxy.golang.org,direct";
            "GOSUMDB" = "sum.golang.org";
          };

          buildInputs = [
            pkgs.go
            pkgs.gopls
          ];
        };
      });
    };
}
