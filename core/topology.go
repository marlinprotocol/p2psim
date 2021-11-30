package core

import (
	"errors"
	"math"

	exprand "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	AvgDeg = 16
)

var (
	TooFewNodesErr = errors.New("Must create atleast two nodes to create a network topology!")
)

// In a real-life network,
// - connections are constantly made and broken
// - peer nodes use various discovery mechanisms such as multicast-DNS, distributed hash tables and so .. on
//     to discover other peers
// To simplify our simulation, we
// - assume that the graph is static thorughout the simulation
// - peers are randomly connected and are connected to a specified number of peers on average

// numNodes is the order of the graph/number of vertices
// generates an undirected graph with 10 peers each
func NewGraph(numNodes int, rng exprand.Source) (graph.Undirected, error) {
	if numNodes < 2 {
		return nil, TooFewNodesErr
	}

	// create an undirected graph with `order` nodes
	grph := simple.NewUndirectedGraph()
	for i := 0; i < numNodes; i++ {
		node := grph.NewNode()
		grph.AddNode(node)
	}

	// TODO: graph generation algorithm based on the configuration
	// NOTE: The average degree is chosen arbitrarily as 16
	// NOTE: Although graphs with order atmost 16 cannot have nodes of degree 16, the algorithm already handles this
	addEdges(grph, AvgDeg, rng)
	return grph, nil
}

// Algorithm described in the paper "Efficient Generation of Networks with Given Expected Degrees"
//   accessible at http://aric.hagberg.org/papers/miller-2011-efficient.pdf
// Assume that every node has an expected degree deg
// TODO: See if there are alternatives that mimic the actual network topologies better
func addEdges(grph *simple.UndirectedGraph, deg int, rng exprand.Source) {
	// uniform distribution
	// tolerance is chosen arbitrarily so that log of that number if not too high in absolute value
	tolerance := 1e-2
	dist := &distuv.Uniform{
		Min: 0.0,
		Max: 1.0,
		Src: rng,
	}

	// take nodes into an array
	nodes := GetNodeSlice(grph.Nodes())
	numNodes := len(nodes)

	// Chung Lu algortihm to generate a simple undirected graph having an expected degree `deg`
	edgeCount := numNodes * deg
	for u := 0; u < numNodes-1; u++ {
		v := u + 1
		p := float64(deg) * float64(deg) / float64(edgeCount)
		if p > 1.0 {
			p = 1.0
		}
		for v < numNodes && p > 0 {
			if p < 1.0-tolerance {
				r := dist.Rand()
				v += int(math.Log(r) / math.Log(1.0-p))
			}
			if v < numNodes {
				// here q == p since the expected degrees of all the nodes are the same
				if r := dist.Rand(); r < 1.0 {
					grph.SetEdge(grph.NewEdge(nodes[u], nodes[v]))
					// edge v, u is automatically added since this is an undirected graph
				}
				v++
			}
		}
	}
}

func GetNodeSlice(nodeIt graph.Nodes) []graph.Node {
	nodes := []graph.Node{}
	for nodeIt.Next() {
		nodes = append(nodes, nodeIt.Node())
	}
	return nodes
}
