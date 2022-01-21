package gossipsub

import (
	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/pubsub"
)

type MessageCache struct {
	msgs map[pubsub.MsgID]pubsub.Message

	// each element of the outer array represents a period of history
	// higher indices represent older histories
	history [][]pubsub.MsgID
}

func NewMessageCache(historyLength int) *MessageCache {
	return &MessageCache{
		msgs:    map[pubsub.MsgID]pubsub.Message{},
		history: make([][]pubsub.MsgID, historyLength),
	}
}

func (mcache *MessageCache) Add(msg pubsub.Message) {
	msgID := pubsub.MsgID{
		From:  msg.From(),
		Seqno: msg.Seqno(),
	}
	mcache.msgs[msgID] = msg
	mcache.history[0] = append(mcache.history[0], msgID)
}

func (mcache *MessageCache) GetMessage(msgID pubsub.MsgID) (pubsub.Message, bool) {
	msg, ok := mcache.msgs[msgID]
	return msg, ok
}

func (mcache *MessageCache) GetGossipIDs(historyGossip int) *core.Set {
	gossipIDs := core.NewSet()
	// only pick historyGossip (< historyLength) items since some windows were shifted after publsihing ihaves
	for _, window := range mcache.history[:historyGossip] {
		for _, msgID := range window {
			gossipIDs.Add(msgID)
		}
	}
	return gossipIDs
}

func (mcache *MessageCache) Shift() {
	oldest := mcache.history[len(mcache.history)-1]
	for _, msgID := range oldest {
		// msgID not present in more recent windows since a message is processed only once (owing to seen cache)
		delete(mcache.msgs, msgID)
	}

	// move the windows to the higher indices
	for i := len(mcache.history) - 2; i >= 0; i-- {
		mcache.history[i+1] = mcache.history[i]
	}
	mcache.history[0] = []pubsub.MsgID{}
}
