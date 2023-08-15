package algorithm

import (
	"github.com/samanazadi/load-balancer/configs"
	"testing"
)

func TestNew(t *testing.T) {
	// RR
	cfg := &configs.Config{Algorithm: configs.Algorithm{Name: RRType}}
	alg, err := New(cfg)
	if _, ok := alg.(*RoundRobin); !ok {
		t.Errorf("algoritm.New(RRType) != RoundRobin")
	}
	if err != nil {
		t.Errorf("algoritm.New(TCPType) returns error")
	}
	// invalid type
	cfg = &configs.Config{Algorithm: configs.Algorithm{Name: "invalid"}}
	alg, err = New(cfg)
	if err == nil {
		t.Errorf("algoritm.New(invalid type) doesn't return error")
	}
}
