# PPSNR
Short for **Parallel PSNR**. Basically it calculates the PSNR between two YUV420
video sequences by distributing the calculation per frame across the available
CPU cores.

This is written in Go for following reasons:
- Easy to cross-compile and have standalone binaries for multiple platforms
- As a personal experiment to work with Go and general and the goroutines in
particular

Enjoy!