package algorithm

import (
	"fmt"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"net/http"
)

const (
	RRType = "rr"
	CHType = "ch"
)

// Algorithm is a balancing algorithm like round-robin and consistent hashing
type Algorithm interface {
	GetNextEligibleNode(*http.Request) *node.Node // based on alive, argument and implementation logic (RR, ...)
	SetNodes([]*node.Node)
}

func New(cfg *configs.Config) (Algorithm, error) {
	switch cfg.Algorithm.Name {
	case RRType:
		return NewRoundRobin(), nil
	case CHType:
		return NewConsistentHashing(cfg)
	default:
		return nil, fmt.Errorf("invalid algorithm: %s", cfg.Algorithm.Name)
	}
}
