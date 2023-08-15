package app

import (
	"fmt"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"net/url"
	"testing"
)

type Test struct {
	url   *url.URL
	alive bool
	setTo bool
	want  bool
}

func createFakeNodes() ([]*node.Node, []*Test) {
	// data preparation
	var tests []*Test
	urls := []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
		"http://localhost:8004",
	}
	alives := []bool{true, true, false, false}
	setTos := []bool{true, false, true, false}

	var nodes []*node.Node
	for i := range urls {
		uu, _ := url.Parse(urls[i])
		n := node.Node{
			URL: uu,
		}
		nodes = append(nodes, &n)

		tests = append(tests, &Test{
			url:   uu,
			alive: alives[i],
			setTo: setTos[i],
			want:  setTos[i],
		})
	}
	return nodes, tests
}

func AliveToString(alive bool) string {
	if alive {
		return "Alive"
	}
	return "Dead"
}

func TestSPSetNodeAlive(t *testing.T) {
	nodes, tests := createFakeNodes()
	pool := ServerPool{
		nodes: nodes,
	}

	for _, test := range tests {
		name := fmt.Sprintf("%sTo%s", AliveToString(test.alive), AliveToString(test.setTo))
		t.Run(name, func(t *testing.T) {
			pool.setNodeAlive(test.url, test.setTo)
			for _, n := range pool.nodes { // search for corresponding node
				if n.URL.String() == test.url.String() {
					// node found
					if n.IsAlive() != test.want {
						t.Errorf("ServerPool.setNodeAlive(%t) = %t, want %t", test.setTo, n.IsAlive(), test.want)
					}
					return
				}
			}
			t.Errorf("ServerPool.setNodeAlive(%t) did nothing, want %t", test.setTo, test.want)
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
	pool := ServerPool{nodes: nodes}

	if len(pool.nodes) != len(urls) {
		t.Errorf("ServerPool.newServerPool(%d nodes) caused %d nodes", len(urls), len(pool.nodes))
	}
	for i, u := range urls {
		if pool.nodes[i].URL.String() != u {
			t.Errorf("ServerPool.newServerPool(node.Node{URL: %s}) not added", u)
		}
	}
}
