package main

import (
	"math"
	"path/filepath"
	"sort"
	"testing"
)

func TestPSNR_OfExampleYuvEqualSize(t *testing.T) {
	expectedValues := []float64{37.64, 36.63, 37.24}
	expectedDtsPtsValues := []DtsPts{{0, 3}, {1, 1}, {2, 2}}
	expectedDtsPtsValuesSortedByPts := []DtsPts{{1, 1}, {2, 2}, {0, 3}}
	inF, _ := filepath.Abs("input.yuv")
	outF, _ := filepath.Abs("output.yuv")
	psnrValues := calculatePsnr(inF, outF, 144, 176)
	dtsValues := readDtsPts("dts_pts_test_values.txt")
	if len(psnrValues) != len(expectedValues) {
		t.Fatalf("Insufficient amount of values returned: %d vs. %d expected", len(psnrValues), len(expectedValues))
	}
	for i := 0; i < len(psnrValues); i++ {
		if math.Abs(expectedValues[i]-psnrValues[i]) > 0.005 {
			t.Errorf("PSNR different from frame %d: %3.2f instead of %3.2f", i, psnrValues[i], expectedValues[i])
		}
	}
	for i := 0; i < len(dtsValues); i++ {
		if dtsValues[i] != expectedDtsPtsValues[i] {
			t.Errorf("DTS value different for frame %d: %d instead of %d", i, dtsValues[i], expectedDtsPtsValues[i])
		}
	}
	if len(dtsValues) != 3 {
		t.Errorf("Length of retrieved DTS values not 3: %d", len(dtsValues))
	}
	sort.Sort(ByDts(dtsValues))
	if len(dtsValues) != 3 {
		t.Errorf("Length of retrieved DTS values not 3: %d", len(dtsValues))
	}
	// Order should not have changed
	for i := 0; i < len(dtsValues); i++ {
		if dtsValues[i] != expectedDtsPtsValues[i] {
			t.Errorf("Sorted by DTS: DTS value different for frame %d: %d instead of %d", i, dtsValues[i], expectedDtsPtsValues[i])
		}
	}
	sort.Sort(ByPts(dtsValues))
	if len(dtsValues) != 3 {
		t.Errorf("Length of retrieved DTS values not 3: %d", len(dtsValues))
	}
	// Order should not have changed
	for i := 0; i < len(dtsValues); i++ {
		if dtsValues[i] != expectedDtsPtsValuesSortedByPts[i] {
			t.Errorf("Sorted by PTS: DTS value different for frame %d: %d instead of %d", i, dtsValues[i], expectedDtsPtsValues[i])
			t.Error(dtsValues, expectedDtsPtsValuesSortedByPts)
		}
	}
}
