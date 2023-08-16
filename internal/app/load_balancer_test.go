package app

import (
	"fmt"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"testing"
)

func TestLBSetNodeAlive(t *testing.T) {
	nodes, tests := node.CreateFakeNodes()
	lb := LoadBalancer{
		ServerPool: ServerPool{Nodes: nodes},
	}

	for _, test := range tests {
		name := fmt.Sprintf("%sTo%s", node.AliveToString(test.Alive), node.AliveToString(test.SetTo))
		t.Run(name, func(t *testing.T) {
			lb.SetNodeAlive(test.URL, test.SetTo)
			for _, n := range lb.ServerPool.Nodes { // search for corresponding node
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
