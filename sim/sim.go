package sim

import (
	"log"
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/floodsub"
	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
	exprand "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
)

// TODO: documentation
type Config struct {
	// Different seeds produce different simulation runs
	Seed *uint64 `toml:"seed,omitempty"`

	// Duration for which simulation must run
	RunDuration *time.Duration `toml:"run_duration"`

	// Total number of nodes in the network
	TotalPeers *int `toml:"total_peers"`

	// Duration for which messages are marked as seen
	SeenTTL *time.Duration `toml:"seen_ttl,omitempty"`

	// Expected time to generate the next block
	BlockInterval *time.Duration `toml:"block_interval"`
}

func GetDefaultConfig() *Config {
	seed := uint64(42)
	seenTTL := 2 * time.Minute
	return &Config{
		Seed:    &seed,
		SeenTTL: &seenTTL,
	}
}

func Simulate(cfg *Config, logger *zap.Logger) (*core.Stats, error) {
	var err error

	// Fix seed for reproducible runs
	// multiple simulations can run in parallel
	// however, a single simulation cannot be parallelized if reproducibility is desired
	rng := exprand.NewSource(42)

	// triggers events in chronological order
	log.Printf("Starting a simulator to be run for %v\n", *cfg.RunDuration)
	sched, err := core.NewScheduler(*cfg.RunDuration)
	if err != nil {
		return nil, err
	}

	// construct the static network topology
	topology, err := core.NewGraph(*cfg.TotalPeers, rng)
	if err != nil {
		return nil, err
	}

	// latency simulator
	net, err := pubsub.NewNetwork(sched, *cfg.SeenTTL, rng, logger)
	if err != nil {
		return nil, err
	}

	oracle, err := core.NewBlockGenerator(sched, *cfg.BlockInterval, rng, logger)
	if err != nil {
		return nil, err
	}

	// spawn and connect the nodes to their neighbors
	log.Printf("Spawning %v new nodes in the network\n", *cfg.TotalPeers)
	err = spawnNewNodes(sched, topology, net, oracle, *cfg.SeenTTL, rng, logger)
	if err != nil {
		return nil, err
	}

	sched.Run()
	stats := net.GetFinalStats()
	return &stats, nil
}

func spawnNewNodes(
	sched *core.Scheduler,
	topology graph.Undirected,
	net *pubsub.Network,
	oracle *core.OracleBlockGenerator,
	seenTTL time.Duration,
	rng exprand.Source,
	logger *zap.Logger,
) error {
	pubSubNodes := []*pubsub.Node{}
	nodeIt := topology.Nodes()
	for nodeIt.Next() {
		nodeID := nodeIt.Node().ID()
		pubSubNode, err := floodsub.SpawnNewNode(sched, net, oracle, seenTTL, nodeID, rng, logger)
		if err != nil {
			return err
		}
		pubSubNodes = append(pubSubNodes, pubSubNode)
	}

	for _, pubSubNode := range pubSubNodes {
		// Connect with its peers
		// The connections are made in only one direction (send paths)
		//   the reverse direction is handled by its neighbor
		neighborIt := topology.From(int64(pubSubNode.ID()))
		for neighborIt.Next() {
			neighborID := neighborIt.Node().ID()
			pubSubNode.AddPeer(neighborID)
		}
	}

	return nil
}
