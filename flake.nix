{
    description = "StyleTTS2 finetuning repo";

    nixConfig = {
        extra-substituters = [
            "https://cuda-maintainers.cachix.org"
        ];
        extra-trusted-public-keys = [
            "cuda-maintainers.cachix.org-1:0dq3bujKpuEPMCX6U4WylrUDZ9JyUG0VpVZa7CNfq5E="
        ];
    };

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs?ref=23.11";
        nixpkgs-unstable.url = "github:NixOS/nixpkgs?ref=nixos-unstable";
        nixpkgs-unfree.url = "github:SomeoneSerge/nixpkgs-unfree";
        flake-utils.url = "github:numtide/flake-utils";
    };

    inputs.nixpkgs-unfree.inputs.nixpkgs.follows = "nixpkgs";

    outputs = { self, nixpkgs, nixpkgs-unstable, flake-utils, ... }:
        flake-utils.lib.eachDefaultSystem (system:
            let
                pkgs = import nixpkgs {
                    inherit system;
                    config = {
                        allowUnfree = true;
                        segger-jlink.acceptLicense = true;
                        acceptCudaLicense = true;
                        cudaSupport = true;
                        # cudaVersion = "12";
                    };
                    inherit (pkgs.cudaPackages) cudatoolkit;
                    inherit (pkgs.linuxPackages) nvidia_x11;
                };

                pkgs_unstable = import nixpkgs-unstable {
                    inherit system;
                    config = {
                        allowUnfree = true;
                        segger-jlink.acceptLicense = true;
                        acceptCudaLicense = true;
                        cudaSupport = true;
                        # cudaVersion = "12";
                    };
                    inherit (pkgs.cudaPackages) cudatoolkit;
                    inherit (pkgs.linuxPackages) nvidia_x11;
                };

                pythonEnv = pkgs.python311.withPackages (p: with p; [
                    jupyter
                    ipython

                    tqdm
                    typing
                    typing-extensions
                    phonemizer
                    pydub
                    pysrt

                    youtube-dl
                    ctranslate2
                    pandas
                    faster-whisper
                    # numpy

                    torch
                    torchvision
                    torchaudio

                    # whisperx
                ]);

            in
            {
                devShells.default = pkgs.mkShellNoCC {
                    buildInputs = with pkgs;[
                        git gitRepo gnupg autoconf curl
                        procps gnumake util-linux m4 gperf unzip
                        cudatoolkit linuxPackages.nvidia_x11
                        libGLU libGL
                        xorg.libXi xorg.libXmu freeglut
                        xorg.libXext xorg.libX11 xorg.libXv xorg.libXrandr zlib 
                        ncurses5 stdenv.cc binutils

                        pythonEnv
                    ];

                    packages = with pkgs; [
                        pkgs_unstable.deepfilternet
                        ffmpeg
                        espeak-ng
                        poetry
                        zlib
                    ];

                    shellHook = ''
                        export PYTHONPATH="${pythonEnv}/${pythonEnv}/bin/python"/${pythonEnv.sitePackages}
                        export CUDA_PATH=${pkgs.cudatoolkit}
                        export LD_LIBRARY_PATH=${pkgs.linuxPackages.nvidia_x11}/lib:${pkgs.ncurses5}/lib:${pkgs.zlib}/lib:${pkgs.stdenv.cc.cc.lib}/lib
                        export EXTRA_LDFLAGS="-L/lib -L${pkgs.linuxPackages.nvidia_x11}/lib"
                        export EXTRA_CCFLAGS="-I/usr/include"
                    '';
                };
            }
        );
}
