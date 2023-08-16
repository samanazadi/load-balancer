package algorithm

import (
	"fmt"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"hash/crc32"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func BenchmarkCHGetNextEligibleNode(b *testing.B) {
	// setup
	const count = 100
	nodes := make([]*node.Node, 0, count)
	cfg := &configs.Config{}
	cfg.Port = 8000
	cfg.Algorithm.Params = map[string]any{"replicas": 100.0, "hashFunc": "crc32"}

	for i := 0; i < count; i++ {
		u, _ := url.Parse("http://localhost:" + strconv.Itoa(i))
		n := &node.Node{
			URL: u,
		}
		n.SetAlive(i%2 == 0)
		nodes = append(nodes, n)
	}

	ch, _ := NewConsistentHashing(cfg)
	ch.SetNodes(nodes)

	r := httptest.NewRequest("GET", "localhost:"+strconv.Itoa(cfg.Port), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.GetNextEligibleNode(r)
	}
}

func TestCHSetNodes(t *testing.T) {
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

	tests := []int{1, 10, 50, 100, 300, 1000}
	for _, replicas := range tests {
		t.Run(fmt.Sprintf("%dReplicas", replicas), func(t *testing.T) {
			ch := ConsistentHashing{
				Replicas: replicas,
				HashFunc: crc32.ChecksumIEEE,
			}

			// size of nodes
			ch.SetNodes(nodes)
			if len(ch.Nodes) != len(urls) {
				t.Errorf("ConsistentHashing.SetNodes(%d nodes) caused %d nodes", len(urls), len(ch.Nodes))
			}

			// check for each node
			for i, u := range urls {
				if ch.Nodes[i].URL.String() != u {
					t.Errorf("ConsistentHashing.SetNodes(node.Node{URL: %s}) not added", u)
				}
			}

			// vnodes
			if len(ch.VNodes) != len(urls)*(replicas) {
				t.Errorf("ConsistentHashing.SetNodes(%d nodes) caused %d vnodes", len(urls), len(ch.VNodes))
			}

			// actual nodes
			if len(ch.ActualNodes) != len(urls)*(replicas) {
				t.Errorf("ConsistentHashing.SetNodes(%d nodes) caused %d actual nodes", len(urls), len(ch.ActualNodes))
			}
		})
	}
}
