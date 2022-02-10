package episub

import (
	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/pubsub"
)

type RPCMsg struct {
	size       int64
	discovery  *DiscoveryMessage
	membership *MembershipMessage
	broadcast  *BroadcastMessage
}

type DiscoveryMessage struct {
	getNodes *GetNodes
	nodes    *Nodes
}

type MembershipMessage struct {
	join         *Join
	forwardJoin  *ForwardJoin
	neighbor     *Neighbor
	disconnect   *Disconnect
	leave        *Leave
	shuffle      *Shuffle
	shuffleReply *SuffleReply
}

type BroadcastMessage struct {
	gossip []Gossip
	ihave  *IHave
	prune  *Prune
	graft  *Graft
}

// Initial node discovery
type GetNodes struct{}

type Nodes struct {
	peerIDs []int64
	ttl     int
}

// Membership mgmt
type Join struct {
	peerID int64
	ttl    int
}

type ForwardJoin struct {
	peerID int64
	ttl    int
}

type Neighbor struct {
	peerIDs []int64
}

type Disconnect struct{}

type Leave struct {
	srcID int64
	ttl   int
}

type Shuffle struct {
	peerID  int64
	peerIDs []int64
	ttl     int
}

type SuffleReply struct {
	peerIDs []int64
}

// Broadcast
type Gossip struct {
	msg  pubsub.Message
	hops int
}

type IHave struct {
	msgSummaries *core.Set
}

type MessageSummary struct {
	msgID pubsub.MsgID
	hops  int
}

type Prune struct{}

type Graft struct {
	msgIDs []pubsub.MsgID
}

func NewDataMsg(msg pubsub.Message) *RPCMsg {
	return &RPCMsg{
		size: msg.GetSize(),
	}
}

/// -------

func NewDataMsg(msg pubsub.Message) *RPCMsg {
	return &RPCMsg{
		size:    msg.GetSize(),
		msgs:    []pubsub.Message{msg},
		control: nil,
	}
}

func NewControlMsg(msgs []pubsub.Message, ihave *IHave, iwant *IWant, graft *Graft, prune *Prune) *RPCMsg {
	// compute size
	size := int64(0)
	for _, msg := range msgs {
		size += msg.GetSize()
	}
	if ihave != nil {
		size += int64(ihave.msgIDs.Len()) * 8
	}
	if iwant != nil {
		size += int64(iwant.msgIDs.Len()) * 8
	}
	if graft != nil {
		size++
	}
	if prune != nil {
		size++
	}

	control := &ControlMessage{
		ihave: ihave,
		iwant: iwant,
		graft: graft,
		prune: prune,
	}
	return &RPCMsg{
		size:    size,
		msgs:    msgs,
		control: control,
	}
}

func (rpcMsg *RPCMsg) GetSize() int64 {
	return rpcMsg.size
}

func (rpcMsg *RPCMsg) GetMessages() []pubsub.Message {
	return rpcMsg.msgs
}
