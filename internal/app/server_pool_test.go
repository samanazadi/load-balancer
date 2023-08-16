package app

import (
	"fmt"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"net/url"
	"testing"
)

func TestSPSetNodeAlive(t *testing.T) {
	nodes, tests := node.CreateFakeNodes()
	pool := ServerPool{
		Nodes: nodes,
	}

	for _, test := range tests {
		name := fmt.Sprintf("%sTo%s", node.AliveToString(test.Alive), node.AliveToString(test.SetTo))
		t.Run(name, func(t *testing.T) {
			pool.SetNodeAlive(test.URL, test.SetTo)
			for _, n := range pool.Nodes { // search for corresponding node
				if n.URL.String() == test.URL.String() {
					// node found
					if n.IsAlive() != test.Want {
						t.Errorf("ServerPool.setNodeAlive(%t) = %t, want %t", test.SetTo, n.IsAlive(), test.Want)
					}
					return
				}
			}
			t.Errorf("ServerPool.setNodeAlive(%t) did nothing, want %t", test.SetTo, test.Want)
		})
	}
}

func TestNewServerPool(t *testing.T) {
	urls := []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
		"http://localhost:8004",
	}
	var nodes []*node.Node
	for i := range urls {
		uu, _ := url.Parse(urls[i])
		n := node.Node{
			URL: uu,
		}
		nodes = append(nodes, &n)
	}
	pool := NewServerPool(nodes, nil)

	if len(pool.Nodes) != len(urls) {
		t.Errorf("ServerPool.newServerPool(%d nodes) caused %d nodes", len(urls), len(pool.Nodes))
	}
	for i, u := range urls {
		if pool.Nodes[i].URL.String() != u {
			t.Errorf("ServerPool.newServerPool(node.Node{URL: %s}) not added", u)
		}
	}
}
