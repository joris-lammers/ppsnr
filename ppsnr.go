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
	"sort"
	"strconv"
	"strings"
	"sync"
)

var flagWidth int
var flagHeight int
var flagYuvRef string
var flagYuvCompr string
var flagDtsFile string
var flagVerbose bool
var flagCodingOrder bool

type DtsPts struct {
	DTS, PTS int64
}

// Sorting by the DTS value
// Implements sort.Interface
type ByDts []DtsPts

func (d ByDts) Len() int           { return len(d) }
func (d ByDts) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ByDts) Less(i, j int) bool { return d[i].DTS < d[j].DTS }

// Sorting by PTS value
// Implements sort.Interface
type ByPts []DtsPts

func (d ByPts) Len() int           { return len(d) }
func (d ByPts) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ByPts) Less(i, j int) bool { return d[i].PTS < d[j].PTS }

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -w WIDTH -h HEIGHT -r REF_YUV -c COMPR_YUV [-v] [- d PTSDTS] [-coding-order]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&flagWidth, "w", 1280, "Width of video")
	flag.IntVar(&flagHeight, "h", 720, "Height of video")
	flag.StringVar(&flagYuvRef, "r", "input.yuv", "Reference YUV")
	flag.StringVar(&flagYuvCompr, "c", "output.yuv", "Compressed/Output YUV")
	flag.StringVar(&flagDtsFile, "d", "", "File containing PTSDTS values for each picture as csv (PTS,DTS)")
	flag.BoolVar(&flagVerbose, "v", false, "Verbose output")
	flag.BoolVar(&flagCodingOrder, "coding-order", false, "Assume YUV in coding order")
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

func readDtsPts(fileName string) []DtsPts {
	dts, err := os.Open(fileName)
	if err != nil {
		// No DTS file provided
		return []DtsPts{}
	}
	dtsPtsValues := []DtsPts{}
	fileReader := bufio.NewScanner(dts)
	for fileReader.Scan() {
		splitted := strings.Split(fileReader.Text(), ",")
		ptsAsInt64, errConv := strconv.ParseInt(splitted[0], 10, 64)
		if errConv != nil {
			fmt.Printf("Error converting PTS: %v\n", errConv)
		}
		dtsAsInt64, errConv := strconv.ParseInt(splitted[1], 10, 64)
		if errConv != nil {
			fmt.Printf("Error converting DTS: %v\n", errConv)
		}

		dtsPtsValues = append(dtsPtsValues, DtsPts{dtsAsInt64, ptsAsInt64})
	}
	return dtsPtsValues
}

func main() {
	flag.Parse()
	if flagVerbose {
		fmt.Printf("Number of CPU cores %d\n", runtime.NumCPU())
	}
	dtsAndPts := readDtsPts(flagDtsFile)
	psnrValues := calculatePsnr(flagYuvRef, flagYuvCompr, flagWidth, flagHeight)

	if flagCodingOrder {
		sort.Sort(ByDts(dtsAndPts))
	} else {
		sort.Sort(ByPts(dtsAndPts))
	}

	for frameNr, p := range psnrValues {
		var dts, pts int64 = -1, -1
		if frameNr < len(dtsAndPts) {
			dts = dtsAndPts[frameNr].DTS
			pts = dtsAndPts[frameNr].PTS
		}
		fmt.Printf("PSNR for frame %d DTS %10d PTS %10d is %3.2f\n", frameNr, dts, pts, p)
	}
}
