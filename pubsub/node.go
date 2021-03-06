package pubsub

import (
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"go.uber.org/zap"
	exprand "golang.org/x/exp/rand"
)

const (
	// Arbitrarily chosen
	BlockSize = 48 * 1024
)

// Generic Node interface
// Implements the routing protocol by handling network messages `HandleMsg`
//   `Msg` itself is a generic interface for network messages
//   apart from updating their own state, nodes are expected to forward the message using the appropriate protocols
// Nodes can register a timer to generate heartbeats on timer ticks
//   this is done by implementing `HandleBeat`
//   GetBeatInterval determines the time interval and is expected to be constant both in time and across nodes
// Apart from handling messages, some nodes generate messages (blocks in this case)
//   `HandleBlockGen` is triggered in intervals taken from a probability distribution (currently exponential)
//   nodes are expected to send the message using the appropriate protocol
type Node struct {
	Sched       *core.Scheduler
	router      Router
	NeighborIDs *core.Set // set of PeerID
	SeenMsgs    *SeenCache
	localID     int64
	link        *MuxLink
	nextSeqno   int64
}

type Router interface {
	Start(node *Node, logger *zap.Logger) error
	PublishMsg(srcID int64, msg Message)
	HandleRPC(srcID int64, rpcMsg RPC)
}

type BlockMsg struct {
	from  int64
	seqno int64
}

func SpawnNewNode(
	sched *core.Scheduler,
	net *Network,
	oracle *core.OracleBlockGenerator,
	seenTTL time.Duration,
	router Router,
	localID int64,
	rng exprand.Source,
	logger *zap.Logger,
) (*Node, error) {
	node := &Node{
		Sched:       sched,
		router:      router,
		NeighborIDs: core.NewSet(),
		SeenMsgs:    NewSeenCache(seenTTL),
		localID:     localID,
		link:        nil,
		nextSeqno:   0,
	}

	// Register ourselves as miner/block publisher
	oracle.AddPublisher(node)

	// Add the local node to the network
	node.link = net.AddNode(localID, node)

	return node, nil
}

func (node *Node) HandleRPC(srcID int64, rpcMsg RPC) {
	for _, msg := range rpcMsg.GetMessages() {
		msgID := MsgID{
			From:  msg.From(),
			Seqno: msg.Seqno(),
		}
		if node.SeenMsgs.MarkSeen(msgID, node.Sched.CurTime) {
			node.router.PublishMsg(srcID, msg)
		}
	}

	node.router.HandleRPC(srcID, rpcMsg)
}

func (node *Node) SendRPC(remoteID int64, rpcMsg RPC) {
	node.link.SendRPC(remoteID, rpcMsg)
}

func (node *Node) PublishNewBlock() {
	node.nextSeqno++
	// Since this message is generated locally, srcID has little meaning
	node.router.PublishMsg(node.localID, &BlockMsg{
		from:  node.localID,
		seqno: node.nextSeqno,
	})
}

func (node *Node) AddPeer(remoteID int64) {
	node.NeighborIDs.Add(remoteID)
}

func (node *Node) Start(logger *zap.Logger) error {
	return node.router.Start(node, logger)
}

func (node *Node) ID() int64 {
	return node.localID
}

func (blockMsg *BlockMsg) GetSize() int64 {
	return BlockSize
}

func (blockMsg *BlockMsg) From() int64 {
	return blockMsg.from
}

func (blockMsg *BlockMsg) Seqno() int64 {
	return blockMsg.seqno
}
