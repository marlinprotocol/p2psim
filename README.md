# p2psim

*This is a work in progress.*

Simulate custom p2p overlay network protocols for comparing performance across various designs. The simulation is determinsitic, configurable and runs to completion quicker than if the events were run in real-time.

## Background

[libp2p](https://docs.libp2p.io/) specifies a baseline [pubsub](https://docs.libp2p.io/concepts/publish-subscribe/) protocol and builds the floodsub, [gossipsub](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md) and the [episub](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/episub.md) protocols on top of the baseline protocol. These and similar protocols are commonly used for communication in blockchain applications at the network layer. The performance of such an application not only depends on the variations introduced to the existing protocols but also on the numerous parameters the protocols are configured by. Simulating the peer to peer overlay network communication enables protocol designers to fine tune both the algortihm as well as its parameters for application specific use cases.

[Discrete Event Simulations](https://en.wikipedia.org/wiki/Discrete-event_simulation) are better suited for the usecase in the prototype phase because of

* **Higher reproducibility** allowing better debugging of the protocol
* **Easier setup and deployment** because the application is single threaded and makes no use of actual network resources
* **Faster time to completion** since the events are simulated on logical time and not real-time

## Evaluation metrics

Some of the important metrics that help evaluate the performance are

* **Message delay**: This metric represents the mean delay for a message to reach a node. Lower delay implies quicker consensus which can inturn help reduce forks.
* **Bandwidth consumption**: This metric represents the mean bytes transferred over the network inorder to transfer a particular message. Keep in mind that messages may reach some nodes more than once and that those messages still consume bandwidth.
* **Network reachability**: This metric indicates how far the messages reach over the network. Typically, the messages reach all the nodes and henceforth most protocols have a 100% reachability.

## Build and Usage

```bash
make p2psim
```

A binary, named `p2psim` is placed in the `build` directory relative to the current directory.

The binary takes several configuration options, simulates a pubsub network protocol in accordance with the options passed via the configuration file and outputs statistics that indicate the protocol's network performance.

Supported config formats: TOML

```bash
./build/p2psim -c config.toml
```

## Configuration Schema

| Path            | Description                                                   | Type     | Example          | Default  | Additional Constraints |
|-----------------|---------------------------------------------------------------|----------|------------------|----------|------------------------|
| run\_duration   | Duration for which the simulation is run                      | duration | "1h"<br>(1 hour) | Required | Must be positive       |
| total\_peers    | Total number of nodes simulated in the network                | integer  | 1024             | Required | Must be at least 2     |
| seen\_ttl       | Duration for which the pubsub framework retains past messages | duration | "5m"<br>(5 mins) | "2m"     | Must be positive       |
| block\_interval | Expected time to generate the next block                      | duration | "15s"            | Required | Must be positive       |

## Example Configuration

```toml
run_duration = "1h"
total_peers = 1024
seen_ttl = "5m"
block_interval = "15s"
```
