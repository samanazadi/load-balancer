package algorithm

import (
	"github.com/samanazadi/load-balancer/internal/models/node"
	"net/url"
	"testing"
)

func TestNextIndex(t *testing.T) {
	rr := RoundRobin{
		lastUsedIndex: -1,
		Nodes:         make([]*node.Node, 4),
	}

	for i := 0; i < 10; i++ {
		rr.nextIndex()
		want := i % 4
		got := rr.lastUsedIndex
		if got != want {
			t.Errorf("RoundRobin.nextIndex() caused lastUsedIndex = %d, want = %d", got, want)
		}
	}
}

func TestGetNextEligibleNode(t *testing.T) {
	// test data preparation
	urls := []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
		"http://localhost:8004",
	}
	alives := []bool{false, true, true, false}
	var nodes []*node.Node
	for i := range urls {
		uu, _ := url.Parse(urls[i])
		n := node.Node{
			URL: uu,
		}
		n.SetAlive(alives[i])
		nodes = append(nodes, &n)
	}
	rr := RoundRobin{
		lastUsedIndex: -1,
		Nodes:         nodes,
	}
	wants := []string{
		"http://localhost:8002",
		"http://localhost:8003",
		"http://localhost:8002",
	}

	for _, want := range wants {
		n := rr.GetNextEligibleNode(nil)
		got := n.URL.String()
		if got != want {
			t.Errorf("RoundRobin.GetNextEligibleNode() = %s, want %s", got, want)
		}
	}
}

func TestSetNodes(t *testing.T) {
	urls := []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
	}
	var nodes []*node.Node
	for _, u := range urls {
		uu, _ := url.Parse(u)
		n := node.Node{
			URL: uu,
		}
		nodes = append(nodes, &n)
	}
	rr := RoundRobin{}

	rr.SetNodes(nodes)
	if len(rr.Nodes) != len(urls) {
		t.Errorf("RoundRobin.SetNodes(%d nodes) caused %d nodes", len(urls), len(rr.Nodes))
	}
	for i, u := range urls {
		if rr.Nodes[i].URL.String() != u {
			t.Errorf("RoundRobin.SetNodes(node.Node{URL: %s}) not added", u)
		}
	}
}

func TestNewRoundRobin(t *testing.T) {
	alg := NewRoundRobin()
	rr, ok := alg.(*RoundRobin)

	if !ok {
		t.Errorf("NewRoundRobin() doesn't create a RoundRobin")
	}
	if got := rr.lastUsedIndex; got != -1 {
		t.Errorf("NewRoundRobin().lastUsedIndex = %d, want -1", got)
	}
}

func BenchmarkRRGetNextEligibleNode(b *testing.B) {
	// setup
	const count = 100
	nodes := make([]*node.Node, 0, count)
	for i := 0; i < count; i++ {
		n := &node.Node{
			URL: nil,
		}
		n.SetAlive(i%2 == 0)
		nodes = append(nodes, n)
	}

	rr := NewRoundRobin()
	rr.SetNodes(nodes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr.GetNextEligibleNode(nil)
	}
}
