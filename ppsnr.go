package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
)

var flagWidth int
var flagHeight int
var flagYuvRef string
var flagYuvCompr string

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -w WIDTH -h HEIGHT -r REF_YUV -c COMPR_YUV\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&flagWidth, "w", 1280, "Width of video")
	flag.IntVar(&flagHeight, "h", 720, "Height of video")
	flag.StringVar(&flagYuvRef, "r", "input.yuv", "Reference YUV")
	flag.StringVar(&flagYuvCompr, "c", "output.yuv", "Compressed/Output YUV")
}

func getInAndOutFrames(inFile, outFile *os.File, frameSize int) (inFrames, outFrames int) {
	inFileStat, _ := inFile.Stat()
	inFrames = int(inFileStat.Size() / int64(frameSize))
	outFileStat, _ := outFile.Stat()
	outFrames = int(outFileStat.Size() / int64(frameSize))
	return
}

func psnr(frameNr int, YR, YC []byte) (psnrValue float64) {
	var noise float64
	for n := 0; n < len(YR); n++ {
		yr := float64(YR[n])
		yc := float64(YC[n])
		noise += (yr - yc) * (yr - yc)
	}
	if noise == 0 {
		psnrValue = 100
	} else {
		psnrValue = 10 * math.Log10(255.0*255.0*float64(len(YR))/noise)
	}
	return
}

func main() {
	flag.Parse()
	frameSize := (flagWidth * flagHeight * 3) / 2
	fmt.Printf("Number of CPU cores %d\n", runtime.NumCPU())
	fmt.Printf("YUV frame size is %d bytes\n", frameSize)
	inFile, _ := os.Open(flagYuvRef)
	outFile, _ := os.Open(flagYuvCompr)
	inFrames, outFrames := getInAndOutFrames(inFile, outFile, frameSize)
	var framesToCompare int
	if inFrames < outFrames {
		framesToCompare = inFrames
	} else {
		framesToCompare = outFrames
	}

	fmt.Printf("In: %d frames, Out: %d frames => Compare %d frames\n", inFrames, outFrames, framesToCompare)

	for n := 0; n < framesToCompare; n++ {
		YR := make([]byte, (frameSize*2)/3)
		YC := make([]byte, (frameSize*2)/3)
		inFile.Read(YR)
		outFile.Read(YC)
		p := psnr(n, YR, YC)
		inFile.Seek(int64(frameSize/3), 1)
		outFile.Seek(int64(frameSize/3), 1)
		fmt.Printf("PSNR for frame %d is %3.2f\n", n, p)
	}
}
