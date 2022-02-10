package episub

import (
	"time"

	"github.com/marlinprotocol/p2psim/pubsub"
)

/**

Utilizes plumtree protocol to maintain the tree overlay
Source: https://www.gsd.inesc-id.pt/~ler/docencia/rcs1617/papers/srds07.pdf

Protocol Summary
================

Basic Idea
----------

There are a few important metrics deciding the performance of any pubsub protocol (learn more in core/stats.go)

- latency
- bandwidth consumption
- fault tolerance

There is an implicit tradeoff between the latency and bandwidth consumption of a typical network overlay.
Although flooding the network with messages provides
- lower latencies by virtue of always finding the shortest path
- robustness to network joins and leaves since otherwise the previously traversable paths are no longer available
  and new paths formed, also termed as fault tolerance
flooding the network has the disadvantage of possessing enormous bandwidth requirements thus limiting scalability

The plumtree protocol is a combination of two distinct mechanisms

- the protocol constructs and maintains a spanning tree through which messages can be sent without any redundancy
- the tree is maintained in a way to optimize message latency
- however, the tree structure is not fault tolerant since any single disconnection leads to disconencted network
- thus, the protocol provides a second mechanism that periodically broadcasts messages in a mesh-like overlay
  to ensure the redundancy required by fault tolerance (this is the periodic gossip)

Architecture
------------

- Peers are maintained by a membershipship service extraneous to the protocol
- The periodic gossip contains message summaries and not complete messages to reduce bandwidth consumption
- The neighbors that are part of the tree are called eager peers because the received messages are relayed to
  these peers immediately(eagerly)
- The messages are stored and broadcasted periodically, similar to gossipsub, to lazy peers
- The tree overlay is constructed initially as a mesh and then the edges are pruned slowly to make a tree
- The tree overlay could change with time as and when either some suboptimal paths are detected or
  there are changes in the network with nodes joining and leaving

Notes
-----

- The tree paths are not completely "optimal" paths but they're a best effort minization of latency
- The lazy peers mesh is not symmetric in general

*/

var (
	TreeTimeout   = 1 * time.Minute
	GraftTimeout  = 1 * time.Minute
	Threshold     = 2
	HistoryLength = 5
	HistoryGossip = 3
)

type TreeConfig struct {
	TreeTimeout   *time.Duration `toml:"tree_timeout,omitempty"`
	GraftTimeout  *time.Duration `toml:"graft_timeout,omitempty"`
	Threshold     *int           `toml:"threshold,omitempty"`
	HistoryLength *int           `toml:"history_length,omitempty"`
	HistoryGossip *int           `toml:"history_gossip,omitempty"`
}

type TreeOverlay struct {
	// assigned when initializing the pubsub node
	node *pubsub.Node

	mcache *MessageCache

	eagerPushPeers *Set
	lazyPushPeers  *Set
}

func GetDefaultConfig() *TreeConfig {
	return &TreeConfig{
		TreeTimeout:   &TreeTimeout,
		GraftTimeout:  &GraftTimeout,
		Threshold:     &Threshold,
		HistoryLength: &HistoryLength,
		HistoryGossip: &HistoryGossip,
	}
}

func NewTreeOverlay() *TreeOverlay {

}

// For unknown messages,
// - add to the list of missing messages
// - start a timer to wait for receiving the message via eager pushes
//   on timer fire, the protocol tries to retrieve the missing message (and simultaneously repair the tree)
func HandleIHave(ihave *IHave) {

}

func (tree *TreeOverlay) Broadcast(msg pubsub.Message, srcID int64) {
	tree.mache.Add(msg)

	seqno += 1
	msgID := MsgID{
		From:  localID,
		Seqno: seqno,
	}
	tree.eagerPush(msg, srcID)
}

func (tree *TreeOverlay) eagerPush(msg pubsub.Message, srcID int64) {
	tree.eagerPushPeers.Traverse(func(iEagerPeerID interface{}) {
		eagerPeerID := iEagerPeerID.(int64)
		if eagerPeerID != msg.From() && eagerPeerID != srcID {
			tree.node.SendRPC(eagerPeerID, NewDataMsg(msg))
		}
	})
}

func (tree *TreeOverlay) lazyPush() {
	gossipSet := tree.mcache.GetGossipIDs(tree.cfg.HistoryGossip)
	tree.lazyPushPeers.Traverse(func(iLazyPeerID interface{}) {
		lazyPeerID := iLazyPeerID.(int64)
		tree.node.SendRPC(lazyPeerID)
	})
}
