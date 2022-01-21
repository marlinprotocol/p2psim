package sim

import (
	"math"
	"testing"
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/gossipsub"
	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
)

func TestFloodSub(t *testing.T) {
	var tolerance float64

	seed := uint64(42)
	dur := time.Hour
	numPeers := 1024
	seenTTL := 2 * time.Minute
	blockInterval := 15 * time.Second
	router := FloodSub
	cfg := &Config{
		Seed:          &seed,
		RunDuration:   &dur,
		TotalPeers:    &numPeers,
		SeenTTL:       &seenTTL,
		BlockInterval: &blockInterval,
		Router:        &router,
	}
	nullLogger := zap.L()
	stats, err := Simulate(cfg, nullLogger)
	if err != nil {
		t.Error("Unexpected error!")
	}

	numFragments := float64((pubsub.BlockSize + pubsub.MaxPayloadSize - 1) / pubsub.MaxPayloadSize)
	expectedPacketCountPerMsg := numFragments * core.AvgDeg * float64(numPeers)
	// 10% tolerance
	tolerance = 0.1 * expectedPacketCountPerMsg
	if math.Abs(stats.PacketCountPerMsg.Value-expectedPacketCountPerMsg) > tolerance {
		t.Errorf("Simulated mean packet count: %v", stats.PacketCountPerMsg.Value)
	}

	expectedTrafficPerMsg := pubsub.BlockSize * core.AvgDeg * float64(numPeers)
	// 10% tolerance
	tolerance = 0.1 * expectedTrafficPerMsg
	if math.Abs(stats.TrafficPerMsg.Value-expectedTrafficPerMsg) > tolerance {
		t.Errorf("Simulated mean traffic: %v", stats.TrafficPerMsg.Value)
	}

	// this calculation is very inaccurate but can serve as a stop gap for a more rigorous analysis
	numHops := math.Log(float64(numPeers)) / math.Log(core.AvgDeg)
	expectedMeanDelay := (pubsub.BaseLatency + pubsub.SpikeProb*pubsub.SpikeLatency) * numHops
	tolerance = 0.1 * expectedMeanDelay
	if math.Abs(stats.DelayMsPerMsg.Value-expectedMeanDelay) > tolerance {
		t.Errorf("Simulated mean delay: %v", stats.DelayMsPerMsg.Value)
	}

	tolerance = 1
	if math.Abs(stats.DeliveredPart.Value-100) > tolerance {
		t.Errorf("Simulated mean delivery percent: %v", stats.DeliveredPart.Value)
	}
}

// gossipsub without configured heartbeats behaves similar to floodsub in a static network
func TestGossipSubNoHeartbeats(t *testing.T) {
	seed := uint64(42)
	dur := time.Hour
	numPeers := 1024
	seenTTL := 2 * time.Minute
	blockInterval := 15 * time.Second
	upperD := 8
	router := GossipSub
	heartbeatInterval := 1 * time.Hour
	routerConfig := gossipsub.GetDefaultConfig()
	routerConfig.HeartbeatInterval = &heartbeatInterval
	routerConfig.Dhigh = &upperD
	cfg := &Config{
		Seed:          &seed,
		RunDuration:   &dur,
		TotalPeers:    &numPeers,
		SeenTTL:       &seenTTL,
		BlockInterval: &blockInterval,
		Router:        &router,
		GossipSub:     routerConfig,
	}
	nullLogger := zap.L()
	stats, err := Simulate(cfg, nullLogger)
	if err != nil {
		t.Error("Unexpected error!")
	}

	numFragments := float64((pubsub.BlockSize + pubsub.MaxPayloadSize - 1) / pubsub.MaxPayloadSize)
	lowerPacketCount := numFragments * float64(*routerConfig.Dlow) * float64(numPeers)
	upperPacketCount := numFragments * float64(*routerConfig.Dhigh) * float64(numPeers)
	if lowerPacketCount > stats.PacketCountPerMsg.Value || stats.PacketCountPerMsg.Value > upperPacketCount {
		t.Errorf("Simulated mean packet count: %v", stats.PacketCountPerMsg.Value)
	}

	lowerTraffic := pubsub.BlockSize * float64(*routerConfig.Dlow) * float64(numPeers)
	upperTraffic := pubsub.BlockSize * float64(*routerConfig.Dhigh) * float64(numPeers)
	if lowerTraffic > stats.TrafficPerMsg.Value || stats.TrafficPerMsg.Value > upperTraffic {
		t.Errorf("Simulated mean traffic: %v", stats.TrafficPerMsg.Value)
	}

	// this calculation is very inaccurate but can serve as a stop gap for a more rigorous analysis
	lowerNumHops := math.Log(float64(numPeers)) / math.Log(float64(*routerConfig.Dhigh))
	upperNumHops := math.Log(float64(numPeers)) / math.Log(float64(*routerConfig.Dlow))
	lowerMeanDelay := (pubsub.BaseLatency + pubsub.SpikeProb*pubsub.SpikeLatency) * lowerNumHops
	upperMeanDelay := (pubsub.BaseLatency + pubsub.SpikeProb*pubsub.SpikeLatency) * upperNumHops
	if lowerMeanDelay > stats.DelayMsPerMsg.Value || stats.DelayMsPerMsg.Value > upperMeanDelay {
		t.Errorf("Simulated mean delay: %v", stats.DelayMsPerMsg.Value)
	}

	// require more than 99% delivery guarantee
	if stats.DeliveredPart.Value < 99 {
		t.Errorf("Simulated mean delivery percent: %v", stats.DeliveredPart.Value)
	}
}

// ensure the overhead traffic from heartbeats does not have exceed floodsub
func TestGossipSubWithHeartbeats(t *testing.T) {
	seed := uint64(42)
	dur := time.Hour
	numPeers := 1024
	seenTTL := 2 * time.Minute
	blockInterval := 15 * time.Second
	upperD := 8
	router := GossipSub
	heartbeatInterval := 1 * time.Minute
	routerConfig := gossipsub.GetDefaultConfig()
	routerConfig.HeartbeatInterval = &heartbeatInterval
	routerConfig.Dhigh = &upperD
	cfg := &Config{
		Seed:          &seed,
		RunDuration:   &dur,
		TotalPeers:    &numPeers,
		SeenTTL:       &seenTTL,
		BlockInterval: &blockInterval,
		Router:        &router,
		GossipSub:     routerConfig,
	}
	nullLogger := zap.L()
	stats, err := Simulate(cfg, nullLogger)
	if err != nil {
		t.Error("Unexpected error!")
	}

	floodsubTraffic := pubsub.BlockSize * core.AvgDeg * float64(numPeers)
	if stats.TrafficPerMsg.Value > floodsubTraffic {
		t.Error("Traffic from gossipsub cannot be higher than that of floodsub!")
	}
}
