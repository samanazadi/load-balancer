package strategy

import (
	"github.com/samanazadi/load-balancer/internal/node"
	"net/http"
)

const (
	RR = iota
)

type Strategy interface {
	GetNextEligibleNode(*http.Request) *node.Node // based on enabled, alive, argument and implementation logic (RR, least connection, ...)
}
