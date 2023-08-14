package app

import (
	"net/http"
)

const (
	RR = "rr"
)

// Strategy is a balancing strategy like round-robin and consistent hashing
type Strategy interface {
	GetNextEligibleNode(*http.Request) *Node // based on alive, argument and implementation logic (RR, ...)
	SetNodes([]*Node)
}
