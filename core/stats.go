package core

// Final result of a simulation goes into `Stats`
// Various stats thus collected from the network simulation help compare protocol performance
//
// Important metrics that decide the performance of a p2p network protocol
//
// Average message delay: Lower message latency leads to faster consensus among peers
// Ex: The floodsub protocol has the potential for minimal latencies since the protocol simply forwards the incoming
//   message to all its peers
//
// Bandwidth consumption: lower latencies in isolation do not mean much
//   if the wire bandwidth is saturated dropping messages frequently
// Ex: Although the floodsub protocol can potentially have the lowest latencies possible, the protocol floods the
//   network resulting in an efficient usage of bandwidth
//
// Reach of the network: messages must reach all corners of the network
// Ex: A protocol that simply doesnt forward any message has a zero latency and bandwidth usage.
//   However, such a protocol is useless since no one hears of the message. Consensus should be among all peers.
//
// Control messages are not counted as messages and counts as overhead for data messages
// NOTE: We do not consider the replies in the RPC protocol while computing the stats
// NOTE: Headers are not included in the bytes transferred

type MeanStat struct {
	Count int64
	Value float64
}

type Stats struct {
	// Mean number of packets transferred per message
	PacketCountPerMsg MeanStat

	// Mean number of bytes transferred per message
	TrafficPerMsg MeanStat

	// Mean delay per message
	DelayMsPerMsg MeanStat

	// Mean percentage of nodes that received the message
	DeliveredPart MeanStat
}

// mean of nth value is (sum of n-1 nums + nth num) / n
// sum of n-1 nums is (mean of n-1) * (n-1)
// substituting, we can arrive at the below expr is correct
func (stat *MeanStat) AddValue(value float64) {
	stat.Count++
	stat.Value += (value - stat.Value) / float64(stat.Count)
}

func (stat *MeanStat) AddMeanStat(other *MeanStat) {
	stat.Count += other.Count
	stat.Value += (other.Value - stat.Value) * float64(other.Count) / float64(stat.Count)
}
