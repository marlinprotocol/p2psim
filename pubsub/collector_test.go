package pubsub

import (
	"errors"
	"math"
	"testing"
	"time"
)

type CollectorRPC struct {
	size int64
	msg  *CollectorMsg
}

type CollectorMsg struct {
	from  int64
	seqno int64
}

func (rpcMsg *CollectorRPC) GetSize() int64 {
	return rpcMsg.size
}

func (rpcMsg *CollectorRPC) GetMessages() []Message {
	return []Message{rpcMsg.msg}
}

func (msg *CollectorMsg) GetSize() int64 {
	return 0
}

func (msg *CollectorMsg) From() int64 {
	return msg.from
}

func (msg *CollectorMsg) Seqno() int64 {
	return msg.seqno
}

func TestNegTTL(t *testing.T) {
	_, err := NewStatCollector(-1 * time.Second)
	if !errors.Is(err, NegTTLErr) {
		t.Error("Cannot specify a negative TTL")
	}
}

func TestHalfRecv(t *testing.T) {
	collector, err := NewStatCollector(time.Hour)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	nodeIDs := []int64{16, 8, 24}
	for _, nodeID := range nodeIDs {
		collector.AddNode(nodeID)
	}

	tolerance := 1e-6
	rpcMsgSize := int64(24 * 1024)
	rpcMsg := &CollectorRPC{
		size: rpcMsgSize,
		msg: &CollectorMsg{
			from:  nodeIDs[0],
			seqno: 36,
		},
	}

	delay := 100
	sendTime := time.Time{}
	recvTime := sendTime.Add(time.Duration(delay) * time.Millisecond)

	// send a message to only one node and not the other
	collector.CollectSendStats(nodeIDs[0], rpcMsg, sendTime)
	collector.CollectRecvStats(nodeIDs[1], rpcMsg, recvTime)
	stats := collector.GetFinalStats()

	// delivered to 50% of the nodes
	deliveredPart := 50.0

	// check stats are as expected
	if math.Abs(stats.DeliveredPart.Value-deliveredPart) > tolerance {
		t.Errorf("delivery percentage value: %v", stats.DeliveredPart.Value)
	}

	if math.Abs(stats.DelayMsPerMsg.Value-float64(delay)) > tolerance {
		t.Errorf("delay ms value: %v", stats.DelayMsPerMsg.Value)
	}
}

// A -- 100ms --> B -- 200ms --> C
// retransmit an identical message from A
func TestForward(t *testing.T) {
	collector, _ := NewStatCollector(time.Hour)

	nodeIDs := []int64{22, 11, 34}
	for _, nodeID := range nodeIDs {
		collector.AddNode(nodeID)
	}

	tolerance := 1e-6
	rpcMsgSize := int64(1_000)
	packetSize := rpcMsgSize + RPCOverhead

	firstDelay := 100
	secondDelay := 200
	epoch := time.Time{}

	for i := 0; i < 2; i++ {
		rpcMsg := &CollectorRPC{
			size: rpcMsgSize,
			msg: &CollectorMsg{
				from:  nodeIDs[0],
				seqno: 36,
			},
		}

		sendTime := epoch.Add(time.Duration(i * (firstDelay + secondDelay)))
		forwardTime := sendTime.Add(time.Duration(firstDelay) * time.Millisecond)
		recvTime := forwardTime.Add(time.Duration(secondDelay) * time.Millisecond)

		collector.CollectSendStats(nodeIDs[0], rpcMsg, sendTime)
		collector.CollectRecvStats(nodeIDs[1], rpcMsg, forwardTime)
		// B does not forward the message the second time since it has already seen the message earlier
		if i == 0 {
			collector.CollectSendStats(nodeIDs[1], rpcMsg, forwardTime)
			collector.CollectRecvStats(nodeIDs[2], rpcMsg, recvTime)
		}
	}

	stats := collector.GetFinalStats()

	// delays: 100ms, 300ms
	// avg delay: 200ms
	packetCount := int64(3)
	traffic := packetCount * packetSize
	avgDelay := float64(firstDelay) + float64(secondDelay)/2
	deliveredPart := 100.0

	if math.Abs(stats.PacketCountPerMsg.Value-float64(packetCount)) > tolerance {
		t.Errorf("packet count value: %v", stats.PacketCountPerMsg.Value)
	}

	if math.Abs(stats.TrafficPerMsg.Value-float64(traffic)) > tolerance {
		t.Errorf("traffic value: %v", stats.TrafficPerMsg.Value)
	}

	if math.Abs(stats.DelayMsPerMsg.Value-avgDelay) > tolerance {
		t.Errorf("delay value: %v", stats.DelayMsPerMsg.Value)
	}

	if math.Abs(stats.DeliveredPart.Value-deliveredPart) > tolerance {
		t.Errorf("delivered part value: %v", stats.DeliveredPart.Value)
	}
}

// Ensure atleast some messages are retired
func TestCacheEviction(t *testing.T) {
	collector, _ := NewStatCollector(time.Second)

	nodeIDs := []int64{13, 40}
	for _, nodeID := range nodeIDs {
		collector.AddNode(nodeID)
	}

	tolerance := 1e-6
	rpcMsgSize := int64(1_000)
	packetSize := rpcMsgSize + RPCOverhead

	delay := 100
	epoch := time.Time{}

	for i := 0; i < 100; i++ {
		rpcMsg := &CollectorRPC{
			size: rpcMsgSize,
			msg: &CollectorMsg{
				from:  nodeIDs[0],
				seqno: int64(7 + 2*i),
			},
		}

		sendTime := epoch.Add(time.Duration(i*delay) * time.Millisecond)
		recvTime := sendTime.Add(time.Duration(delay) * time.Millisecond)

		collector.CollectSendStats(nodeIDs[0], rpcMsg, sendTime)
		collector.CollectRecvStats(nodeIDs[1], rpcMsg, recvTime)
	}

	stats := collector.GetFinalStats()

	// only one packet sent from A to B for each message
	packetCount := 1
	traffic := packetSize
	avgDelay := float64(delay)
	deliveredPart := 100.0

	if math.Abs(stats.PacketCountPerMsg.Value-float64(packetCount)) > tolerance {
		t.Errorf("packet count value: %v", stats.PacketCountPerMsg.Value)
	}

	if math.Abs(stats.TrafficPerMsg.Value-float64(traffic)) > tolerance {
		t.Errorf("traffic value: %v", stats.TrafficPerMsg.Value)
	}

	if math.Abs(stats.DelayMsPerMsg.Value-avgDelay) > tolerance {
		t.Errorf("delay value:%v count: %v", stats.DelayMsPerMsg.Value, stats.DelayMsPerMsg.Count)
	}

	if math.Abs(stats.DeliveredPart.Value-deliveredPart) > tolerance {
		t.Errorf("delivered part value: %v", stats.DeliveredPart.Value)
	}
}
