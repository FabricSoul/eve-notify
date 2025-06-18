{
  description = "Dev shell for Fyne + systray apps on NixOS";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05"; # Use 24.05 or your preferred version
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
      };

      # GLFW, X11 headers, and other Fyne dependencies
      buildDeps = with pkgs; [
        go
        pkg-config
        xorg.libX11
        xorg.libXcursor
        xorg.libXrandr
        xorg.libXinerama
        xorg.libXi
        xorg.libXext
        xorg.libXxf86vm
        xorg.xinput
        xorg.libSM
        xorg.libICE
        libGL
        gtk3
        cairo
        gdk-pixbuf
        glib
        pango
        atk
        dbus
        libayatana-appindicator
        alsa-lib
      ];
    in {
      devShell = pkgs.mkShell {
        name = "fyne-dev-shell";
        packages = buildDeps;

        shellHook = ''
          echo "Fyne + systray dev shell ready!"
        '';
      };
    });
}
