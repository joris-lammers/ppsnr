package main

import (
	"math"
	"path/filepath"
	"testing"
)

func TestPSNR_OfExampleYuvEqualSize(t *testing.T) {
	expectedValues := []float64{37.64, 36.63, 37.24}
	inF, _ := filepath.Abs("input.yuv")
	outF, _ := filepath.Abs("output.yuv")
	psnrValues := calculatePsnr(inF, outF, 144, 176)
	if len(psnrValues) != len(expectedValues) {
		t.Fatalf("Insufficient amount of values returned: %d vs. %d expected", len(psnrValues), len(expectedValues))
	}
	for i := 0; i < len(psnrValues); i++ {
		if math.Abs(expectedValues[i]-psnrValues[i]) > 0.005 {
			t.Errorf("PSNR different from frame %d: %3.2f instead of %3.2f", i, psnrValues[i], expectedValues[i])
		}
	}
}
