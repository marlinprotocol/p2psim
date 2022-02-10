package episub

import (
	"time"

	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
)

/**

Implements the proximity aware epidemic pubsub protocol
Source: https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/episub.md

Protocol Summary
================

Basic idea
----------

*/

type Router struct {
	// episub config params
	cfg *Config

	tree *TreeOverlay
}

type Config struct {
	// Interval between consecutive heartbeats
	// Heartbeats are triggers for periodic gossip
	HeartbeatInterval *time.Duration `toml:"heartbeat_interval,omitempty"`

	Tree       *TreeConfig
	Membership *MembershipConfig
}

func (router *Router) Start(node *pubsub.Node, logger *zap.Logger) error {
}

func (router *Router) PublishMsg(srcID int64, msg pubsub.Message) {
}

func (router *Router) HandleRPC(srcID int64, rpcMsg pubsub.RPC) {
}
