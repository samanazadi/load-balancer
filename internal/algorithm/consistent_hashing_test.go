package algorithm

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/models/node"
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
