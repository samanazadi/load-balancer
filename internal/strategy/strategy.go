package strategy

import (
	"github.com/samanazadi/load-balancer/internal/node"
	"net/http"
)

const (
	RR = "rr"
)

// Strategy is a balancing strategy like round-robin and consistent hashing
type Strategy interface {
	GetNextEligibleNode(*http.Request) *node.Node // based on enabled, alive, argument and implementation logic (RR, least connection, ...)
	SetNodes([]*node.Node)
}
