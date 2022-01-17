package gossipsub

import (
	"testing"

	"github.com/marlinprotocol/p2psim/pubsub"
)

type MCacheMsg struct {
	from  int64
	seqno int64
}

func (msg *MCacheMsg) GetSize() int64 {
	return 0
}

func (msg *MCacheMsg) From() int64 {
	return msg.from
}

func (msg *MCacheMsg) Seqno() int64 {
	return msg.seqno
}

func TestShift(t *testing.T) {
	mcache := NewMessageCache(2)
	mcache.Add(&MCacheMsg{
		from:  0,
		seqno: 0,
	})
	mcache.Shift()
	mcache.Add(&MCacheMsg{
		from:  0,
		seqno: 1,
	})
	mcache.Shift()
	mcache.Add(&MCacheMsg{
		from:  0,
		seqno: 2,
	})

	// Check existence
	if _, exists := mcache.GetMessage(pubsub.MsgID{
		From:  0,
		Seqno: 0,
	}); exists {
		t.Error("The first message must have been evicted after shifting twice!")
	}

	if _, exists := mcache.GetMessage(pubsub.MsgID{
		From:  0,
		Seqno: 1,
	}); !exists {
		t.Error("The second message must not have been evicted after shifting!")
	}

	if _, exists := mcache.GetMessage(pubsub.MsgID{
		From:  0,
		Seqno: 2,
	}); !exists {
		t.Error("The third message was just added and must exist!")
	}

	// Check gossip IDs in both the windows
	bothWindows := mcache.GetGossipIDs(2)
	if bothWindows.Len() != 2 ||
		!bothWindows.Exists(pubsub.MsgID{
			From:  0,
			Seqno: 1,
		}) ||
		!bothWindows.Exists(pubsub.MsgID{
			From:  0,
			Seqno: 2,
		}) {
		t.Error("Incorrect gossipIDs repr!")
	}

	firstWindow := mcache.GetGossipIDs(1)
	if firstWindow.Len() != 1 ||
		!firstWindow.Exists(pubsub.MsgID{
			From:  0,
			Seqno: 2,
		}) {
		t.Error("Incorrect gossipIDs repr!")
	}

	noWindows := mcache.GetGossipIDs(0)
	if noWindows.Len() != 0 {
		t.Error("Incorrect gossipIDs repr!")
	}
}
