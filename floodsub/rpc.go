package floodsub

import (
	"github.com/marlinprotocol/p2psim/pubsub"
)

type RPCMsg struct {
	size int64
	msgs []pubsub.Message
}

func NewDataMsg(msg pubsub.Message) *RPCMsg {
	return &RPCMsg{
		size: msg.GetSize(),
		msgs: []pubsub.Message{msg},
	}
}

func (rpcMsg *RPCMsg) GetSize() int64 {
	return rpcMsg.size
}

func (rpcMsg *RPCMsg) GetMessages() []pubsub.Message {
	return rpcMsg.msgs
}
