package floodsub

import (
	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
)

type Router struct {
	// initialized while initializing the pubsub node
	node *pubsub.Node
}

func NewRouter() *Router {
	return &Router{
		node: nil,
	}
}

func (router *Router) Start(node *pubsub.Node, logger *zap.Logger) error {
	router.node = node

	// no heartbeats registered in floodsub
	return nil
}

func (router *Router) PublishMsg(srcID int64, msg pubsub.Message) {
	router.node.NeighborIDs.Traverse(func(iNeighborID interface{}) {
		neighborID := iNeighborID.(int64)
		if msg.From() != neighborID && srcID != neighborID {
			// do not resend the message back or to the originator of the message
			router.node.SendRPC(neighborID, NewDataMsg(msg))
		}
	})
}

func (router *Router) HandleRPC(srcID int64, rpcMsg pubsub.RPC) {}
