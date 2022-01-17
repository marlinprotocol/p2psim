package core

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Dummy node for simulating heartbeats
type TickerNode struct{}

func (tickerNode *TickerNode) HandleTick() {}
func (tickerNode *TickerNode) ID() int64   { return 0 }

func TestNegTick(t *testing.T) {
	var err error

	nullLogger := zap.L()

	sched, _ := NewScheduler(5 * time.Second)

	// negative tick interval
	err = StartTicker(sched, -1*time.Second, &TickerNode{}, nullLogger)
	if !errors.Is(err, NegTickErr) {
		t.Error("Cannot support negative interval ticks!")
	}

	// zero tick interval
	err = StartTicker(sched, 0, &TickerNode{}, nullLogger)
	if !errors.Is(err, NegTickErr) {
		t.Error("Cannot support zero interval ticks!")
	}
}

func TestTicks(t *testing.T) {
	nullLogger := zap.L()

	ticks := int64(4)
	// ticks+1 since the scheduler does not schedule events falling on end time by design
	sched, _ := NewScheduler(time.Duration(ticks+1) * time.Second)

	// start ticking at 1 second intervals
	// first tick occurs after the first second
	err := StartTicker(sched, time.Second, &TickerNode{}, nullLogger)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	sched.Run()
	if sched.NumTriggered != ticks {
		t.Error("Heartbeat triggered an incorrect number of times!")
	}
}
