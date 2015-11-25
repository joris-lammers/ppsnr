# PPSNR
[![Build Status](https://travis-ci.org/joris-lammers/ppsnr.svg)](https://travis-ci.org/joris-lammers/ppsnr)

Short for **Parallel PSNR**. Basically it calculates the PSNR between two YUV420
video sequences by distributing the calculation per frame across the available
CPU cores.

This is written in Go for following reasons:
- Easy to cross-compile and have standalone binaries for multiple platforms.
- As a personal experiment to work with Go and general and the goroutines in
particular.

The project contains two example YUV files that are also used by the unit test.
These YUV files were taken from [the H.264 reference 
software](http://iphome.hhi.de/suehring/tml/)

Running the binary for the example YUV files (QCIF resolution) is done as:
```
ppsnr -r input.yuv -c output.yuv -h 176 -w 144
```

Getting help about the command line arguments is done using the `--help` argument

Enjoy!