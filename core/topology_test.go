package core

import (
	"errors"
	"math"
	"testing"

	exprand "golang.org/x/exp/rand"
)

func TestTooFewNodes(t *testing.T) {
	var err error

	_, err = NewGraph(-1, exprand.NewSource(2))
	if !errors.Is(err, TooFewNodesErr) {
		t.Error("Cannot create a graph with negative nodes!")
	}

	_, err = NewGraph(0, exprand.NewSource(4))
	if !errors.Is(err, TooFewNodesErr) {
		t.Error("Cannot create a graph with zero nodes!")
	}

	_, err = NewGraph(1, exprand.NewSource(1))
	if !errors.Is(err, TooFewNodesErr) {
		t.Error("Need atleast two nodes to create a graph network")
	}
}

func TestGraphGen(t *testing.T) {
	for _, numNodes := range []int{10, 100, 1_000} {
		rng := exprand.NewSource(314)
		grph, err := NewGraph(numNodes, rng)
		if err != nil {
			t.Error("Unexpected error!")
		}
		expectedDegree := 16
		if expectedDegree > numNodes-1 {
			expectedDegree = numNodes - 1
		}
		nodes := GetNodeSlice(grph.Nodes())
		actualEdgeCount := 0
		for _, node := range nodes {
			actualEdgeCount += grph.From(node.ID()).Len()
		}
		expectedEdgeCount := expectedDegree * numNodes
		tolerance := numNodes
		if int(math.Abs(float64(actualEdgeCount-expectedEdgeCount))) > tolerance {
			t.Errorf("Got %v, expected %v", actualEdgeCount, expectedEdgeCount)
		}
	}
}
