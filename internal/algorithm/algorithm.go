package algorithm

import (
	"github.com/samanazadi/load-balancer/internal/models/node"
	"net/http"
)

const (
	RR = "rr"
)

// Algorithm is a balancing algorithm like round-robin and consistent hashing
type Algorithm interface {
	GetNextEligibleNode(*http.Request) *node.Node // based on alive, argument and implementation logic (RR, ...)
	SetNodes([]*node.Node)
}
