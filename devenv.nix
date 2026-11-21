{ pkgs, lib, config, inputs, ... }:

let
  unstable = import
    (builtins.fetchTarball "https://github.com/NixOS/nixpkgs/archive/nixos-unstable.tar.gz")
    {
      inherit (pkgs) system;
      config.allowUnfree = true;
    };
in {
    packages = [
        pkgs.git
        pkgs.bun
        pkgs.awscli2
        unstable.opentofu
    ];

    dotenv.enable = true;

    languages.go.enable = true;
    languages.go.lsp.enable = true;

    languages.typescript.enable = true;
    languages.typescript.lsp.enable = true;

    enterShell = ''
        git --version
        go mod tidy
        bun install
        tofu --version
        aws --version
    '';

    enterTest = ''
        echo "Running tests"
        git --version | grep --color=auto "${pkgs.git.version}"
    '';

    processes = {
        web-client.exec = "npm run web";
        web-server.exec = "go run ./packages/server/main.go";
    };
}
