package core

// Simulations make extensive use of random numbers
// Some examples
// - network latency is not deterministic and depends on external factors
// - peers are connected at random
// - blocks in a blockchain are generated at random intervals

type Dist interface {
	// Methods are named to be consistent with gonum's distuv interface
	Rand() float64
	Mean() float64
}

type ConstantDist struct {
	Value float64
}

type LatencyDist struct {
	// Modelling latency is not simple
	// latency depends on numerous factors such as
	// - processing delay
	// - queueing delay
	// - transmission delay
	// - propagation delay
	// To make matters simple, let's just assume that there is only normal latency and spiked latency
	// latencies are measured in milliseconds
	SpikeDist    Dist
	BaseLatency  float64
	SpikeLatency float64
}

func (constant *ConstantDist) Rand() float64 {
	return constant.Value
}

func (constant *ConstantDist) Mean() float64 {
	return constant.Value
}

func (latency *LatencyDist) Rand() float64 {
	return latency.BaseLatency + latency.SpikeDist.Rand()*latency.SpikeLatency
}

func (latency *LatencyDist) Mean() float64 {
	return latency.BaseLatency + latency.SpikeDist.Mean()*latency.SpikeLatency
}
