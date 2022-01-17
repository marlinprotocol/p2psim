package core

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

// Generate heartbeats for protocols that require it
// No heartbeats in floodsub and hence no ticker is regisered with the scheduler
// In gossipsub, we have mesh maintenance periodically independent of other messages

var (
	NegTickErr = errors.New("Ticker interval must be positive!")
)

type TickEvent struct {
	ticker *Ticker
}

type Ticker struct {
	sched    *Scheduler
	interval time.Duration
	node     TickHandler
	logger   *zap.Logger
}

type TickHandler interface {
	HandleTick()
	ID() int64
}

func StartTicker(sched *Scheduler, interval time.Duration, node TickHandler, logger *zap.Logger) error {
	if interval <= 0 {
		return NegTickErr
	}
	ticker := &Ticker{
		sched:    sched,
		interval: interval,
		node:     node,
		logger:   logger,
	}
	ticker.scheduleTick()
	return nil
}

// helper for the trigger event
func (ticker *Ticker) Tick() {
	ticker.logger.Debug(
		"Firing a heartbeat",
		zap.Time("CurTime", ticker.sched.CurTime),
		zap.Int64("nodeID", ticker.node.ID()),
	)
	ticker.node.HandleTick()
	ticker.scheduleTick()
}

func (ticker *Ticker) scheduleTick() {
	ticker.sched.Schedule(ticker.interval, &TickEvent{
		ticker: ticker,
	})
}

// Implements the event interface
// The node is notified of the heartbeat tick
// We also schedule the next beat since heartbeats are periodic and not a one-time event
func (tickEvent *TickEvent) Trigger() {
	tickEvent.ticker.Tick()
}
