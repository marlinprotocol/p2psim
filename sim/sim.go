package sim

import (
	"errors"
	"log"
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/floodsub"
	"github.com/marlinprotocol/p2psim/gossipsub"
	"github.com/marlinprotocol/p2psim/pubsub"
	"go.uber.org/zap"
	exprand "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
)

var (
	UnknownRouterErr  = errors.New("Could not recognize the requested router type!")
	UnspecDurErr      = errors.New("Did not configure a run duration!")
	UnspecNumPeerErr  = errors.New("Did not configure the total number of peers!")
	UnspecBlockDurErr = errors.New("Did not configure the block interval!")
	UnspecRouterErr   = errors.New("Did not configure the router type!")
)

const (
	FloodSub  = "floodsub"
	GossipSub = "gossipsub"
)

var (
	// Default config params
	Seed    = uint64(42)
	SeenTTL = 2 * time.Minute
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

	// The type of router to consider
	Router *string `toml:"router"`

	// Configuration options for the gossip router
	// Options enabled iff the router is specified as `gossipsub`
	GossipSub *gossipsub.Config `toml:"gossipsub,omitempty"`
}

func GetDefaultConfig() *Config {
	return &Config{
		Seed:      &Seed,
		SeenTTL:   &SeenTTL,
		GossipSub: gossipsub.GetDefaultConfig(),
	}
}

func Simulate(cfg *Config, logger *zap.Logger) (*core.Stats, error) {
	var err error

	// Fix seed for reproducible runs
	// multiple simulations can run in parallel
	// however, a single simulation cannot be parallelized if reproducibility is desired
	rng := exprand.NewSource(42)

	// triggers events in chronological order
	if cfg.RunDuration == nil {
		return nil, UnspecDurErr
	}
	log.Printf("Starting a simulator to be run for %v\n", *cfg.RunDuration)
	sched, err := core.NewScheduler(*cfg.RunDuration)
	if err != nil {
		return nil, err
	}

	// construct the static network topology
	if cfg.TotalPeers == nil {
		return nil, UnspecNumPeerErr
	}
	topology, err := core.NewGraph(*cfg.TotalPeers, rng)
	if err != nil {
		return nil, err
	}

	// latency simulator
	net, err := pubsub.NewNetwork(sched, *cfg.SeenTTL, rng, logger)
	if err != nil {
		return nil, err
	}

	if cfg.BlockInterval == nil {
		return nil, UnspecBlockDurErr
	}
	oracle, err := core.NewBlockGenerator(sched, *cfg.BlockInterval, rng, logger)
	if err != nil {
		return nil, err
	}

	// spawn and connect the nodes to their neighbors
	log.Printf("Spawning %v new nodes in the network\n", *cfg.TotalPeers)
	err = spawnNewNodes(sched, topology, net, oracle, cfg, rng, logger)
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
	cfg *Config,
	rng exprand.Source,
	logger *zap.Logger,
) error {
	pubSubNodes := []*pubsub.Node{}
	nodeIt := topology.Nodes()
	for nodeIt.Next() {
		nodeID := nodeIt.Node().ID()
		pubSubNode, err := spawnNewNode(sched, net, oracle, cfg, nodeID, rng, logger)
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

	// start the nodes
	for _, pubSubNode := range pubSubNodes {
		err := pubSubNode.Start(logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func spawnNewNode(
	sched *core.Scheduler,
	net *pubsub.Network,
	oracle *core.OracleBlockGenerator,
	cfg *Config,
	nodeID int64,
	rng exprand.Source,
	logger *zap.Logger,
) (*pubsub.Node, error) {
	if cfg.Router == nil {
		return nil, UnspecRouterErr
	}
	switch *cfg.Router {
	case FloodSub:
		router := floodsub.NewRouter()
		return pubsub.SpawnNewNode(sched, net, oracle, *cfg.SeenTTL, router, nodeID, rng, logger)
	case GossipSub:
		router := gossipsub.NewRouter(cfg.GossipSub, rng)
		return pubsub.SpawnNewNode(sched, net, oracle, *cfg.SeenTTL, router, nodeID, rng, logger)
	default:
		return nil, UnknownRouterErr
	}
}
