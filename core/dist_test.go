package core

import (
	"math"
	"testing"

	exprand "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

func TestConst(t *testing.T) {
	constVal := float64(132)
	constDist := &ConstantDist{
		Value: constVal,
	}
	tolerance := 1e-6
	for i := 0; i < 100; i++ {
		if val := constDist.Rand(); math.Abs(val-constDist.Mean()) > tolerance {
			t.Errorf("Got %v, expected %v", val, constDist.Mean())
		}
	}
}

func TestUniform(t *testing.T) {
	rng := exprand.NewSource(132)
	dist := &distuv.Uniform{
		Min: 0.0,
		Max: 1.0,
		Src: rng,
	}
	// figure out the total number of values in these ranges
	// 0-0.25, 0.25-0.5, 0.5-0.75, 0.75-1.0
	samples := 16384
	divisions := 4
	counter := make([]int, divisions)
	for i := 0; i < samples; i++ {
		index := int(math.Floor(dist.Rand() * float64(divisions)))
		if index >= divisions {
			index = divisions - 1
		}
		counter[index]++
	}
	tolerance := 1024
	for i := 0; i < divisions; i++ {
		if int(math.Abs(float64(counter[i]*divisions-samples))) > tolerance {
			t.Errorf("Got %v, expected %v", counter[i]*divisions, samples)
		}
	}
}

func TestBernoulli(t *testing.T) {
	for _, prob := range []float64{0.0, 0.1, 0.5, 1.0} {
		rng := exprand.NewSource(429)
		dist := &distuv.Bernoulli{
			P:   prob,
			Src: rng,
		}
		total := 1_000
		ones := 0
		for i := 0; i < total; i++ {
			if math.Round(dist.Rand()) == 1 {
				ones++
			}
		}
		ratio := float64(ones) / float64(total)
		// round to one decimal place
		approx := math.Round(ratio*float64(10)) / 10
		if approx != prob {
			t.Errorf("Got %v, expected %v", approx, prob)
		}
	}
}

func TestExponential(t *testing.T) {
	for _, rate := range []float64{1.0 / 15, 1.0 / 60, 1.0 / 600} {
		rng := exprand.NewSource(1430)
		dist := &distuv.Exponential{
			Rate: rate,
			Src:  rng,
		}
		totalSamples := 10_000
		expectedTime := float64(totalSamples) / rate
		simTime := float64(0)
		for i := 0; i < totalSamples; i++ {
			simTime += dist.Rand()
		}
		tolerance := 1e3 / rate
		if math.Abs(simTime-expectedTime) > tolerance {
			t.Errorf("Got %v, expected %v", simTime, expectedTime)
		}
	}
}
