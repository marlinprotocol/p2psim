package sim

import (
	"math"
	"testing"
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
)

func TestBasicConfig(t *testing.T) {
	var tolerance float64

	seed := uint64(42)
	dur := time.Hour
	numPeers := 1024
	seenTTL := 2 * time.Minute
	blockInterval := 15 * time.Second
	cfg := &Config{
		Seed:          &seed,
		RunDuration:   &dur,
		TotalPeers:    &numPeers,
		SeenTTL:       &seenTTL,
		BlockInterval: &blockInterval,
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
