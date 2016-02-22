// This utility allows you to calculate the PSNR for every frame between a
// reference YUV420 (8-bit) file and another YUV
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"sync"
)

var flagWidth int
var flagHeight int
var flagYuvRef string
var flagYuvCompr string
var flagDtsFile string
var flagVerbose bool

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -w WIDTH -h HEIGHT -r REF_YUV -c COMPR_YUV [-v] [- d DTS]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&flagWidth, "w", 1280, "Width of video")
	flag.IntVar(&flagHeight, "h", 720, "Height of video")
	flag.StringVar(&flagYuvRef, "r", "input.yuv", "Reference YUV")
	flag.StringVar(&flagYuvCompr, "c", "output.yuv", "Compressed/Output YUV")
	flag.StringVar(&flagDtsFile, "d", "", "File containing DTS values for each picture")
	flag.BoolVar(&flagVerbose, "v", false, "Verbose output")
}

// NewWorkersPool starts n workers
// This function starts n workers (goroutines) and is ready to receive work.
// It returns a work and a wait!
func NewWorkersPool(n int) (chan<- func(int), *sync.WaitGroup) {
	work := make(chan func(int), n)
	var wait sync.WaitGroup
	wait.Add(n)
	for ; n > 0; n-- {
		// idiom: passing a parameter to the "anonymous closure" function
		go func(id int) {
			for x := range work {
				x(id)
			}
			wait.Done()
		}(n)
	}
	return work, &wait
}

func getInAndOutFrames(inFile, outFile *os.File, frameSize int) (inFrames, outFrames int) {
	inFileStat, _ := inFile.Stat()
	inFrames = int(inFileStat.Size() / int64(frameSize))
	outFileStat, _ := outFile.Stat()
	outFrames = int(outFileStat.Size() / int64(frameSize))
	return
}

func psnr(YR, YC []byte) (psnrValue float64) {
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

func yuvFrameSize(w, h int) int {
	return (w * h * 3) / 2
}

func yuvYSize(w, h int) int {
	return w * h
}

func yuvUVSize(w, h int) int {
	return (w * h) / 2
}

func calculatePsnr(refFN, comprFN string, w, h int) (psnrValues []float64) {
	frameSize := yuvFrameSize(w, h)

	inFile, _ := os.Open(refFN)
	outFile, _ := os.Open(comprFN)
	inFrames, outFrames := getInAndOutFrames(inFile, outFile, frameSize)
	var framesToCompare int
	if inFrames < outFrames {
		framesToCompare = inFrames
	} else {
		framesToCompare = outFrames
	}

	work, wait := NewWorkersPool(runtime.NumCPU())
	psnrValues = make([]float64, framesToCompare)

	for n := 0; n < framesToCompare; n++ {
		YR := make([]byte, yuvYSize(w, h))
		YC := make([]byte, yuvYSize(w, h))
		inFile.Read(YR)
		outFile.Read(YC)
		frameNr := n
		work <- func(id int) {
			// The id argument can be used to know wich goroutine is executing the work
			psnrValues[frameNr] = psnr(YR, YC)
		}
		// Skip the UV
		inFile.Seek(int64(yuvUVSize(w, h)), 1)
		outFile.Seek(int64(yuvUVSize(w, h)), 1)
	}
	close(work)
	wait.Wait()

	return
}

func readDts(fileName string) []int64 {
	dts, err := os.Open(fileName)
	if err != nil {
		// No DTS file provided
		return []int64{}
	}
	dtsValues := []int64{}
	fileReader := bufio.NewReader(dts)
	for line, err := fileReader.ReadString('\n'); err == nil; line, err = fileReader.ReadString('\n') {
		dtsAsInt64, errConv := strconv.ParseInt(line[:len(line)-1], 10, 64)
		if errConv != nil {
			fmt.Printf("%v", errConv)
		}
		dtsValues = append(dtsValues, dtsAsInt64)
	}
	return dtsValues
}

func main() {
	flag.Parse()
	if flagVerbose {
		fmt.Printf("Number of CPU cores %d\n", runtime.NumCPU())
	}
	dts := readDts(flagDtsFile)
	psnrValues := calculatePsnr(flagYuvRef, flagYuvCompr, flagWidth, flagHeight)

	for frameNr, p := range psnrValues {
		var d int64 = -1
		if frameNr < len(dts) {
			d = dts[frameNr]
		}
		fmt.Printf("PSNR for frame %d DTS %10d is %3.2f\n", frameNr, d, p)
	}
}
