package pubsub

// Closest parallel is the RPC protocol buffer described in libp2p pubsub
type RPC interface {
	GetSize() int64
	GetMessages() []Message
}

// Closest parallel is the Message protocol buffer described in libp2p pubsub
// We do not verify peer identities using signatures since this is a simulation
// Message ID function is a struct combination of from and seqno
// NOTE: Seqno corresponds to the originator of the message and not the forwarder
type Message interface {
	From() int64
	Seqno() int64
}

// Assume default message ID function
// Combination of `from` and `seqno`
type MsgID struct {
	From  int64
	Seqno int64
}

type RPCHandler interface {
	HandleRPC(srcID int64, rpcMsg RPC)
}
