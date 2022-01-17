package pubsub

import (
	"errors"
	"time"

	"github.com/marlinprotocol/p2psim/core"
)

var (
	NegTTLErr = errors.New("Cannot specify a negative seen TTL duration!")
)

const (
	// Arbitrarily chosen
	RPCOverhead = 64

	// Close to the real payload
	MaxPayloadSize = 1460
)

// Contains logic relevant to stat collection on sending and receiving messages over the network
// See stats.go for more information

type StatCollector struct {
	// duration after which messages are retired
	seenTTL time.Duration

	// compute final stats here
	curStats core.Stats

	// count of retired messages
	msgCount int64

	// packets are counted during the send event
	totalPacketCount int64

	// bytes are counted during the send event
	totalBytesTransferred int64

	// populated the first time the message is encountered
	// entries retired on expiry
	originTimePerMsg map[MsgID]time.Time

	// each entry represents the mean delay in milliseconds over each node
	// entries retired on expiry
	delayMsPerMsg map[MsgID]*core.MeanStat

	// MsgID -> set of non-receivers
	// Typically, we expect the message to be delivered to be everyone eventually
	// => storing the set of nodes that didnt receive the message is easier on memory
	// Messages automatically retired on expiry
	remNodesPerMsg map[MsgID]*core.Set

	// messages sorted by the non-decreasing order of their origin times
	chronoMsgs []*ChronoMsg

	// list of all nodes populated at the origin of each message
	nodeIDs *core.Set
}

type ChronoMsg struct {
	msgID      MsgID
	originTime time.Time
}

// Custom eviction handlers assume that the message is retired at the time of eviction
//   and calculate the appropriate stats
func NewStatCollector(seenTTL time.Duration) (*StatCollector, error) {
	if seenTTL < 0 {
		return nil, NegTTLErr
	}

	collector := &StatCollector{
		curStats:              core.Stats{},
		totalPacketCount:      0,
		totalBytesTransferred: 0,
		originTimePerMsg:      map[MsgID]time.Time{},
		delayMsPerMsg:         map[MsgID]*core.MeanStat{},
		remNodesPerMsg:        map[MsgID]*core.Set{},
		chronoMsgs:            []*ChronoMsg{},
		nodeIDs:               core.NewSet(),
		seenTTL:               seenTTL,
	}
	return collector, nil
}

// Called after simulation run is complete and typically only once
// Returns the final stats after collecting stats from send and receive operations
func (collector *StatCollector) GetFinalStats() core.Stats {
	// retire all messages
	collector.msgCount += int64(len(collector.originTimePerMsg))

	// Collect packet count stats
	collector.curStats.PacketCountPerMsg = core.MeanStat{
		Count: collector.msgCount,
		Value: float64(collector.totalPacketCount) / float64(collector.msgCount),
	}

	// Collect traffic stats
	collector.curStats.TrafficPerMsg = core.MeanStat{
		Count: collector.msgCount,
		Value: float64(collector.totalBytesTransferred) / float64(collector.msgCount),
	}

	// Collect message delays
	for _, meanDur := range collector.delayMsPerMsg {
		collector.curStats.DelayMsPerMsg.AddMeanStat(meanDur)
	}

	// Collect reachbility stats
	for _, remNodes := range collector.remNodesPerMsg {
		// We subtract one in the denominator to exclude the originator of the message
		remRatio := float64(remNodes.Len()) / float64(collector.nodeIDs.Len()-1)
		deliveredRatio := 1.0 - remRatio
		collector.curStats.DeliveredPart.AddValue(100.0 * deliveredRatio)
	}

	// We make a copy to clear the curStats field
	stats := collector.curStats
	collector.clear()
	return stats
}

// Reset the stats
func (collector *StatCollector) clear() {
	collector.curStats = core.Stats{}
	collector.msgCount = 0
	collector.totalPacketCount = 0
	collector.totalBytesTransferred = 0
	collector.originTimePerMsg = map[MsgID]time.Time{}
	collector.delayMsPerMsg = map[MsgID]*core.MeanStat{}
	collector.remNodesPerMsg = map[MsgID]*core.Set{}
	collector.chronoMsgs = []*ChronoMsg{}
	collector.nodeIDs = core.NewSet()
}

// Called to collect stats on message/packet send
func (collector *StatCollector) CollectSendStats(srcID int64, rpcMsg RPC, curTime time.Time) {
	newMsgAlreadyFound := false

	for _, msg := range rpcMsg.GetMessages() {
		msgID := MsgID{
			From:  msg.From(),
			Seqno: msg.Seqno(),
		}

		// Check if the message is new
		if _, exists := collector.originTimePerMsg[msgID]; !exists {
			if !newMsgAlreadyFound {
				collector.retireOldMsgs(curTime)
				newMsgAlreadyFound = true
			}

			// Useful for
			// - retiring older messages
			// - calculating latency at the receiving end
			collector.originTimePerMsg[msgID] = curTime
			collector.chronoMsgs = append(collector.chronoMsgs, &ChronoMsg{
				msgID:      msgID,
				originTime: curTime,
			})

			// Delay calculated on the receiving end
			collector.delayMsPerMsg[msgID] = &core.MeanStat{}

			// Add all the nodes (except src) to the remaining nodes set
			// Delivered percentage calculated on retiring messages
			// Messages are removed from the set whenever the appropriate node receives the message
			collector.remNodesPerMsg[msgID] = collector.excludeSource(srcID)
		}
	}

	rpcMsgSize := rpcMsg.GetSize()
	packetCount := getPacketCount(rpcMsgSize)
	// replies are not counted here
	collector.totalPacketCount += packetCount
	collector.totalBytesTransferred += packetCount*RPCOverhead + rpcMsgSize
}

// Called to collect stats on message/packet receive
func (collector *StatCollector) CollectRecvStats(dstID int64, rpcMsg RPC, curTime time.Time) {
	for _, msg := range rpcMsg.GetMessages() {
		var exists bool

		msgID := MsgID{
			From:  msg.From(),
			Seqno: msg.Seqno(),
		}

		remNodes, exists := collector.remNodesPerMsg[msgID]
		if !exists || !remNodes.Exists(dstID) {
			// This particular message is
			// - Either never seen, globally
			// - Or already seen on this particular node
			return
		}

		// Remove the receiver from the set for the delivery stat
		remNodes.Remove(dstID)

		// We expect msgID to be present since the key set ofr both remNodesPerMsg and originTimePerMsg are the same
		origTime := collector.originTimePerMsg[msgID]

		// Update mean delay
		delay := curTime.Sub(origTime).Milliseconds()
		collector.delayMsPerMsg[msgID].AddValue(float64(delay))
	}
}

func (collector *StatCollector) AddNode(nodeID int64) {
	collector.nodeIDs.Add(nodeID)
}

func (collector *StatCollector) retireOldMsgs(curTime time.Time) {
	// go back seenTTL
	oldestValidTime := curTime.Add(-1 * collector.seenTTL)
	for 0 < len(collector.chronoMsgs) && collector.chronoMsgs[0].originTime.Before(oldestValidTime) {
		msgID := collector.chronoMsgs[0].msgID

		collector.msgCount++

		delete(collector.originTimePerMsg, msgID)
		// first element garbage collected on reallocation
		collector.chronoMsgs = collector.chronoMsgs[1:]

		collector.curStats.DelayMsPerMsg.AddMeanStat(collector.delayMsPerMsg[msgID])
		delete(collector.delayMsPerMsg, msgID)

		remRatio := float64(collector.remNodesPerMsg[msgID].Len()) / float64(collector.nodeIDs.Len()-1)
		deliveredRatio := 1.0 - remRatio
		collector.curStats.DeliveredPart.AddValue(100.0 * deliveredRatio)
		delete(collector.remNodesPerMsg, msgID)
	}
}

func (collector *StatCollector) excludeSource(srcID int64) *core.Set {
	nodeSet := core.NewSet()
	collector.nodeIDs.Traverse(func(iNodeID interface{}) {
		nodeID := iNodeID.(int64)
		if nodeID != srcID {
			nodeSet.Add(nodeID)
		}
	})
	return nodeSet
}

func getPacketCount(rpcMsgSize int64) int64 {
	return (rpcMsgSize + MaxPayloadSize - 1) / MaxPayloadSize
}
