package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"time"

	"github.com/marlinprotocol/p2psim/core"
	"github.com/marlinprotocol/p2psim/sim"
	toml "github.com/pelletier/go-toml"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

/**

See https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md
  for more info. on the pubsub/floodsub/gossipsub protocols

TODO:

- make topology configurable
- make latency configurable
- log events
  - separates stat computation logic from simulation
  - allowing us to compute better statistics such as the 90th percentile delay and so on
- support multiple simulations in a single run
- make random seed configurable
- make logger configurable
- add support for topics
- support for fanout topics in gossip

*/

var (
	// errors
	NoCfgErr = errors.New("Must configure the application using a config file!")
)

var (
	// command line flags
	configFileFlag = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Load TOML based configuration from `FILE`",
	}
)

func main() {
	flags := []cli.Flag{
		configFileFlag,
	}
	app := cli.App{
		Name:    "p2psim",
		Usage:   "Marlin P2P simulator",
		Action:  p2psim,
		Flags:   flags,
		Version: "v1",
		After: func(ctx *cli.Context) error {
			log.Println("Exiting application ...")
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln("Application terminated with error", zap.Error(err))
	}
}

func p2psim(ctx *cli.Context) error {
	log.Println("Starting application ...")
	var err error

	// Load configuration file passed by the user
	cfg, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	// Initialize the global logger
	// logger, err = zap.NewProductionConfig()
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.OutputPaths = []string{
		"build/events.log",
	}
	logger, err := loggerCfg.Build()
	if err != nil {
		return err
	}
	defer logger.Sync()

	// Run the simulation
	stats, err := sim.Simulate(cfg, logger)
	if err != nil {
		return err
	}

	// Print final stats to stdout
	printStats(stats)
	log.Println("Printing statistics ...")
	return nil
}

// Extracts configuration required for simulation
// options specified in the config file take preference over default options
func loadConfig(ctx *cli.Context) (*sim.Config, error) {
	var err error

	// Must specify a config file
	if !ctx.IsSet(configFileFlag.Name) {
		return nil, NoCfgErr
	}
	filepath := ctx.String(configFileFlag.Name)
	log.Printf("Loading config file at location %v\n", filepath)

	// Extract configuration options from the config file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Decode toml config
	// the configuration schema is specified by sim.Config
	// below code disallows specification of extraneous config options
	cfg := &sim.Config{}
	if err = toml.NewDecoder(bufio.NewReader(file)).Strict(true).Decode(cfg); err != nil {
		return nil, err
	}

	// success
	return cfg, nil
}

func printStats(stats *core.Stats) {
	log.Println("Mean packet count:", stats.PacketCountPerMsg)
	log.Println("Mean traffic:", stats.TrafficPerMsg)
	log.Println("Mean delay:", time.Duration(stats.DelayMsPerMsg.Value)*time.Millisecond)
	log.Println("Delivered Percent:", stats.DeliveredPart)
}
