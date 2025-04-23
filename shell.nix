with import <nixpkgs> {}; {
  devEnv = stdenv.mkDerivation {
    name = "dev";
    buildInputs = [ stdenv go glibc.static pkgsStatic.ncurses pkg-config ];

    CFLAGS="-I${pkgs.glibc.dev}/include \
    -I${pkgs.ncurses.dev}/include";

    LDFLAGS="-L${pkgs.glibc}/lib \
    -I${pkgs.ncurses.dev}/include";
    
    shellHook = ''
      export CGO_ENABLED=1
      export CGO_FLAGS="-I${pkgs.ncurses.dev}/include"
      export CGO_LDFLAGS="-L${pkgs.pkgsStatic.ncurses}/lib -lncursesw -static"
    '';
  };
}
