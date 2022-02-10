package gossipsub

import (
	"errors"
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
	exprand "golang.org/x/exp/rand"
)

var (
	InvDegErr  = errors.New("Configured degrees do not follow the required constraints!")
	InvHistErr = errors.New("Configured the message cache incorrectly!")
)

var (
	// default config params
	HeartbeatInterval = 1 * time.Minute
	D                 = 6
	Dlow              = 4
	Dhigh             = 12
	HistoryLength     = 5
	HistoryGossip     = 3
	Dlazy             = 6
)

type Router struct {
	// gossipsub config params
	cfg *Config

	rng exprand.Source

	// assigned while initializing the pubsub node
	node *pubsub.Node

	// set of peers in the the mesh
	// underlying type => int64 (peer ID)
	mesh *core.Set

	// For gossipping IHave messages
	// To respond to IWant messages in reply
	mcache *MessageCache
}

type Config struct {
	// Interval between consecutive heartbeats
	// Heartbeats are triggers for periodic gossip
	HeartbeatInterval *time.Duration `toml:"heartbeat_interval,omitempty"`

	// Desired degree for the mesh.
	// Currently, the network is static and hence the mesh as well (not using peer scoring from v1.1)
	D *int `toml:"D,omitempty"`

	// Ideal lower bound on the degree of the mesh
	Dlow *int `toml:Dlow,omitempty`

	// Upper bound on the degree of the mesh
	Dhigh *int `toml:"Dhigh,omitempty"`

	// Number of heartbeat events for which the message cache remembers seen messages
	HistoryLength *int `toml:"history_length,omitempty"`

	// Determines the number of heartbeat intervals for which the messages arrived are gossipped about
	HistoryGossip *int `toml:"history_gossip,omitempty"`

	// Number of peers the application gossips to
	Dlazy *int `toml:"Dlazy,omitempty"`
}

func GetDefaultConfig() *Config {
	return &Config{
		HeartbeatInterval: &HeartbeatInterval,
		D:                 &D,
		Dlow:              &Dlow,
		Dhigh:             &Dhigh,
		HistoryLength:     &HistoryLength,
		HistoryGossip:     &HistoryGossip,
		Dlazy:             &Dlazy,
	}
}

func NewRouter(cfg *Config, rng exprand.Source) *Router {
	return &Router{
		cfg:    cfg,
		rng:    rng,
		node:   nil,
		mesh:   core.NewSet(),
		mcache: NewMessageCache(*cfg.HistoryLength),
	}
}

func (router *Router) Start(node *pubsub.Node, logger *zap.Logger) error {
	var err error

	if !(*router.cfg.HistoryGossip <= *router.cfg.HistoryLength) {
		return InvHistErr
	}

	router.node = node

	// Add neighbors to mesh
	// NOTE: Joining here since the network is static
	err = router.join()
	if err != nil {
		return err
	}

	// Start timer for heartbeats
	err = core.StartTicker(node.Sched, *router.cfg.HeartbeatInterval, router, logger)
	if err != nil {
		return err
	}

	return nil
}

func (router *Router) join() error {
	if !(0 <= *router.cfg.Dlow &&
		*router.cfg.Dlow <= *router.cfg.D &&
		*router.cfg.D <= *router.cfg.Dhigh &&
		0 <= *router.cfg.Dlazy) {
		return InvDegErr
	}

	// Add upto D neighbors to the mesh
	for _, neighborID := range router.getRandomNeighbors(*router.cfg.D, filterAll()) {
		// send graft
		router.node.SendRPC(neighborID, NewControlMsg([]pubsub.Message{}, nil, nil, &Graft{}, nil))

		// add locally
		router.mesh.Add(neighborID)
	}

	return nil
}

func (router *Router) PublishMsg(srcID int64, msg pubsub.Message) {
	// add message to cache
	router.mcache.Add(msg)

	// publish to all our peers in the mesh
	router.mesh.Traverse(func(iNeighborID interface{}) {
		neighborID := iNeighborID.(int64)
		if neighborID != srcID && neighborID != msg.From() {
			router.node.SendRPC(neighborID, NewDataMsg(msg))
		}
	})
}

func (router *Router) HandleRPC(srcID int64, rpcMsg pubsub.RPC) {
	control := rpcMsg.(*RPCMsg).control
	if control == nil {
		return
	}

	iwant := router.handleIHave(control.ihave)
	msgs := router.handleIWant(control.iwant)
	prune := router.handleGraft(srcID, control.graft)
	router.handlePrune(srcID, control.prune)

	if iwant == nil && len(msgs) == 0 && prune == nil {
		return
	}

	replyMsg := NewControlMsg(msgs, nil, iwant, nil, prune)
	router.node.SendRPC(srcID, replyMsg)
}

func (router *Router) handleIHave(ihave *IHave) *IWant {
	if ihave == nil {
		return nil
	}

	// retrieve messages that were not receieved in the fast path from the mesh
	missing := core.NewSet()
	ihave.msgIDs.Traverse(func(iMsgID interface{}) {
		msgID := iMsgID.(pubsub.MsgID)
		if !router.node.SeenMsgs.SeenMsg(msgID) {
			missing.Add(msgID)
		}
	})
	if missing.Len() == 0 {
		return nil
	}

	return &IWant{
		msgIDs: missing,
	}
}

func (router *Router) handleIWant(iwant *IWant) []pubsub.Message {
	if iwant == nil {
		return []pubsub.Message{}
	}

	msgs := []pubsub.Message{}
	iwant.msgIDs.Traverse(func(iMsgID interface{}) {
		msgID := iMsgID.(pubsub.MsgID)
		msg, exists := router.mcache.GetMessage(msgID)
		if !exists {
			return
		}

		msgs = append(msgs, msg)
	})

	return msgs
}

func (router *Router) handleGraft(remoteID int64, graft *Graft) *Prune {
	if graft == nil {
		return nil
	}

	// already added
	// do not prune
	if router.mesh.Exists(remoteID) {
		return nil
	}

	// cannot add any more peers
	if router.mesh.Len() >= *router.cfg.Dhigh {
		return &Prune{}
	}

	// add peer to mesh
	router.mesh.Add(remoteID)
	return nil
}

func (router *Router) handlePrune(remoteID int64, prune *Prune) {
	if prune == nil {
		return
	}

	router.mesh.Remove(remoteID)
	// NOTE: number of peers in the mesh may fall below Dlow
	//   this is adjusted for periodically during the mesh maintenance in heartbeat
}

func (router *Router) HandleTick() {
	// the mesh is potentially in a bad state because of too few peers
	toGraft := router.fixMesh()

	// NOTE: do not check if the peer count is too high since the count does not go that high
	//   assert router.mesh.Len() <= *router.cfg.Dhigh

	// slow path gossip of available messages
	lazy, gossip := router.emitGossip()

	// send control messages for heartbeats
	router.sendHeartbeats(toGraft, lazy, gossip)

	// shift the cache
	router.mcache.Shift()
}

// Return the set of peers to graft
func (router *Router) fixMesh() *core.Set {
	// set of peers (int64)
	toGraft := core.NewSet()

	// verify that the node is connected to enough peers in the mesh
	if router.mesh.Len() < *router.cfg.Dlow {
		// bring the number of peers up to the ideal value
		deficit := *router.cfg.D - router.mesh.Len()
		neighborIDs := router.getRandomNeighbors(deficit, router.filterOutMesh())
		for _, neighborID := range neighborIDs {
			// cache to send the grafts with gossip
			toGraft.Add(neighborID)

			// add peer to the mesh (since grafting)
			router.mesh.Add(neighborID)
		}
	}

	return toGraft
}

// Do not send gossip to our peers in the mesh because they already have our messages
// returns
// - the peers to send the gossip to
// - messages that are in the cache
func (router *Router) emitGossip() (*core.Set, *core.Set) {
	// gossip to Dlazy peers
	neighborIDs := router.getRandomNeighbors(*router.cfg.Dlazy, router.filterOutMesh())
	gossipSet := core.NewSet()
	for _, neighborID := range neighborIDs {
		gossipSet.Add(neighborID)
	}

	// retrieve messages
	return gossipSet, router.mcache.GetGossipIDs(*router.cfg.HistoryGossip)
}

func (router *Router) sendHeartbeats(toGraft *core.Set, lazy *core.Set, gossip *core.Set) {
	// send grafts
	// sending ihaves to freshly grafted peers is not necessary
	toGraft.Traverse(func(iNeighborID interface{}) {
		neighborID := iNeighborID.(int64)
		router.node.SendRPC(neighborID, NewControlMsg([]pubsub.Message{}, nil, nil, &Graft{}, nil))
		// neighborID was already added to the mesh locally
	})

	// send ihaves to the selected peers
	lazy.Traverse(func(iNeighborID interface{}) {
		neighborID := iNeighborID.(int64)
		router.node.SendRPC(neighborID, NewControlMsg([]pubsub.Message{}, &IHave{msgIDs: gossip}, nil, nil, nil))
	})
}

func (router *Router) getRandomNeighbors(count int, filter func(int64) bool) []int64 {
	neighborIDs := []int64{}
	router.node.NeighborIDs.Traverse(func(iNeighborID interface{}) {
		neighborID := iNeighborID.(int64)
		if filter(neighborID) {
			neighborIDs = append(neighborIDs, neighborID)
		}
	})

	// shuffle our neighbors (pick random count elements from the slice)
	exprand.New(router.rng).Shuffle(len(neighborIDs), func(i, j int) {
		neighborIDs[i], neighborIDs[j] = neighborIDs[j], neighborIDs[i]
	})

	// cannot pick more than the elements already present
	if count > len(neighborIDs) {
		count = len(neighborIDs)
	}
	return neighborIDs[:count]
}

func (router *Router) filterOutMesh() func(int64) bool {
	return func(neighborID int64) bool {
		return !router.mesh.Exists(neighborID)
	}
}

func filterAll() func(int64) bool {
	return func(int64) bool {
		return true
	}
}

func (router *Router) ID() int64 {
	return router.node.ID()
}
