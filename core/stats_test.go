package core

import (
	"math"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/stat"
)

func TestMeanCalc(t *testing.T) {
	rng := rand.New(rand.NewSource(1729))
	meanStat := MeanStat{}
	xs := []float64{}
	for i := 0; i < 1_000; i++ {
		value := rng.Float64()
		meanStat.AddValue(value)
		xs = append(xs, value)
	}
	meanValue := stat.Mean(xs, nil)
	tolerance := 1e-6
	if math.Abs(meanValue-meanStat.Value) > tolerance {
		t.Errorf("Got %v, expected %v", meanStat.Value, meanValue)
	}
}
