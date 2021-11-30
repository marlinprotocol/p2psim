package floodsub

import (
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
	exprand "golang.org/x/exp/rand"
)

type Router struct {
	// initialized while initializing the pubsub node
	node *pubsub.Node
}

func SpawnNewNode(
	sched *core.Scheduler,
	net *pubsub.Network,
	oracle *core.OracleBlockGenerator,
	seenTTL time.Duration,
	localID int64,
	rng exprand.Source,
	logger *zap.Logger,
) (*pubsub.Node, error) {
	router := &Router{}
	// No heartbeats registered in floodsub
	return pubsub.SpawnNewNode(sched, net, oracle, seenTTL, router, localID, rng, logger)
}

func (router *Router) PublishMsg(srcID int64, msg pubsub.Message) {
	for _, neighborID := range router.node.GetNeighbors() {
		if msg.From() != neighborID && srcID != neighborID {
			// do not resend the message back or to the originator of the message
			router.node.SendRPC(neighborID, NewRPCMsg(msg.(*pubsub.BlockMsg)))
		}
	}
}

func (router *Router) Attach(node *pubsub.Node)                 { router.node = node }
func (router *Router) AddPeer(peerID int64)                     {}
func (router *Router) HandleRPC(srcID int64, rpcMsg pubsub.RPC) {}
