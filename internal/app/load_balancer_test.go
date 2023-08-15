package app

import (
	"fmt"
	"testing"
)

func TestLBSetNodeAlive(t *testing.T) {
	nodes, tests := createFakeNodes()
	lb := LoadBalancer{
		serverPool: ServerPool{nodes: nodes},
	}

	for _, test := range tests {
		name := fmt.Sprintf("%sTo%s", AliveToString(test.alive), AliveToString(test.setTo))
		t.Run(name, func(t *testing.T) {
			lb.SetNodeAlive(test.url, test.setTo)
			for _, n := range lb.serverPool.nodes { // search for corresponding node
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
