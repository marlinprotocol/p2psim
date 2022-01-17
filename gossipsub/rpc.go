package gossipsub

import (
	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/pubsub"
)

type RPCMsg struct {
	size    int64
	msgs    []pubsub.Message
	control *ControlMessage
}

type ControlMessage struct {
	ihave *IHave
	iwant *IWant
	graft *Graft
	prune *Prune
}

type IHave struct {
	// Set of MsgID
	msgIDs *core.Set
}

type IWant struct {
	// Set of MsgID
	msgIDs *core.Set
}

type Graft struct{}

type Prune struct{}

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
