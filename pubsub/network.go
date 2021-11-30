package pubsub

import (
	"log"
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"go.uber.org/zap"
	exprand "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

// Simulates network latency and acts as an intermediary for sending and receiving messages

const (
	// latency measured in ms
	// latency on spike = base latency + spike latency
	SpikeProb    = 0.1
	BaseLatency  = 100.0
	SpikeLatency = 100.0
)

type Network struct {
	sched       *core.Scheduler
	nodes       map[int64]RPCHandler
	latencyDist core.Dist
	collector   *StatCollector
	logger      *zap.Logger
}

type MuxLink struct {
	net     *Network
	localID int64
}

type RPCEvent struct {
	net    *Network
	srcID  int64
	dstID  int64
	rpcMsg RPC
}

func NewNetwork(sched *core.Scheduler, seenTTL time.Duration, rng exprand.Source, logger *zap.Logger) (*Network, error) {
	// latency distribution
	latencyDist := &core.LatencyDist{
		SpikeDist: &distuv.Bernoulli{
			P:   SpikeProb,
			Src: rng,
		},
		BaseLatency:  BaseLatency,
		SpikeLatency: SpikeLatency,
	}

	collector, err := NewStatCollector(seenTTL)
	if err != nil {
		return nil, err
	}

	net := &Network{
		sched:       sched,
		nodes:       map[int64]RPCHandler{},
		latencyDist: latencyDist,
		collector:   collector,
		logger:      logger,
	}

	return net, nil
}

func (net *Network) HandleRPC(srcID int64, dstID int64, rpcMsg RPC) {
	if node, exists := net.nodes[dstID]; exists {
		net.logger.Debug(
			"Received RPC message",
			zap.Time("CurTime", net.sched.CurTime),
			zap.Int64("srcID", int64(srcID)),
			zap.Int64("dstID", int64(dstID)),
		)
		net.collector.CollectRecvStats(dstID, rpcMsg, net.sched.CurTime)
		node.HandleRPC(srcID, rpcMsg)
	}
}

// Called after simulation run and only once
func (net *Network) GetFinalStats() core.Stats {
	if !net.sched.IsStopped() {
		log.Fatalln("Cannot retrieve the stats while the scheduler is running!")
	}
	return net.collector.GetFinalStats()
}

func (net *Network) SendRPC(srcID int64, dstID int64, rpcMsg RPC) {
	net.collector.CollectSendStats(srcID, rpcMsg, net.sched.CurTime)
	latency := time.Duration(net.latencyDist.Rand()) * time.Millisecond
	net.sched.Schedule(latency, &RPCEvent{
		net:    net,
		srcID:  srcID,
		dstID:  dstID,
		rpcMsg: rpcMsg,
	})
}

func (net *Network) AddNode(nodeID int64, rpcHandler RPCHandler) *MuxLink {
	net.nodes[nodeID] = rpcHandler
	net.collector.AddNode(nodeID)
	return &MuxLink{
		net:     net,
		localID: nodeID,
	}
}

func (link *MuxLink) SendRPC(remoteID int64, rpcMsg RPC) {
	link.net.SendRPC(link.localID, remoteID, rpcMsg)
}

func (rpcEvent *RPCEvent) Trigger() {
	rpcEvent.net.HandleRPC(rpcEvent.srcID, rpcEvent.dstID, rpcEvent.rpcMsg)
}
