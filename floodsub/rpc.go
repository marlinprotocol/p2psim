package floodsub

import (
	"github.com/marlinprotocol/p2psim/pubsub"
)

type RPCMsg struct {
	size int64
	msgs []pubsub.Message
}

func NewRPCMsg(blockMsg *pubsub.BlockMsg) *RPCMsg {
	return &RPCMsg{
		size: blockMsg.GetSize(),
		msgs: []pubsub.Message{blockMsg},
	}
}

func (rpcMsg *RPCMsg) GetSize() int64 {
	return rpcMsg.size
}

func (rpcMsg *RPCMsg) GetMessages() []pubsub.Message {
	return rpcMsg.msgs
}
