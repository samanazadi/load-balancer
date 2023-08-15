package algorithm

import (
	"github.com/samanazadi/load-balancer/internal/models/node"
	"net/http"
)

const (
	RR = "rr"
)

// Strategy is a balancing strategy like round-robin and consistent hashing
type Strategy interface {
	GetNextEligibleNode(*http.Request) *node.Node // based on alive, argument and implementation logic (RR, ...)
	SetNodes([]*node.Node)
}
