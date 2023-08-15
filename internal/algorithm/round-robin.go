package algorithm

import (
	"github.com/samanazadi/load-balancer/internal/models/node"
	"net/http"
	"sync"
)

type RoundRobin struct {
	lastUsedIndex int
	mux           sync.RWMutex // for protecting lastUsedIndex from multiple access
	Nodes         []*node.Node
}

func (rr *RoundRobin) nextIndex() int {
	rr.mux.Lock()
	defer rr.mux.Unlock()
	rr.lastUsedIndex++
	rr.lastUsedIndex = rr.lastUsedIndex % len(rr.Nodes)
	return rr.lastUsedIndex
}

func (rr *RoundRobin) GetNextEligibleNode(*http.Request) *node.Node {
	next := rr.nextIndex()
	last := next + len(rr.Nodes)
	for i := next; i < last; i++ {
		index := i % len(rr.Nodes)
		if rr.Nodes[index].IsAlive() {
			if i != next {
				// store new current index (some unavailable nodes found)
				rr.mux.Lock()
				rr.lastUsedIndex = index
				rr.mux.Unlock()
			}
			return rr.Nodes[index]
		}
	}
	return nil // no available node
}

func (rr *RoundRobin) SetNodes(nodes []*node.Node) {
	rr.Nodes = nodes
}

func NewRoundRobin() Strategy {
	return &RoundRobin{
		lastUsedIndex: -1,
	}
}
