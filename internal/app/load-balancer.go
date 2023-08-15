package app

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/algorithm"
	"github.com/samanazadi/load-balancer/internal/checker"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/http"
	"net/url"
)

// LoadBalancer is a server pool along an algorithm
type LoadBalancer struct {
	serverPool ServerPool
	algorithm  *algorithm.Algorithm
}

func (lb *LoadBalancer) SetNodeAlive(url *url.URL, alive bool) {
	lb.serverPool.setNodeAlive(url, alive)
}

// ServeHTTP route request based on algorithm
func (lb *LoadBalancer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if n := (*lb.algorithm).GetNextEligibleNode(r); n != nil {
		n.ReverseProxy.ServeHTTP(rw, r)
		return
	}
	logging.Logger.Println("no node is available")
	http.Error(rw, "Service not available", http.StatusServiceUnavailable)
}

// StartPassiveHealthCheck starts passive health check daemon
func (lb *LoadBalancer) StartPassiveHealthCheck(period int) {
	lb.serverPool.startPassiveHealthCheck(period)
}

func New(cfg *configs.Config, chk *checker.ConnectionChecker, alg *algorithm.Algorithm) *LoadBalancer {
	lb := &LoadBalancer{}
	nodes := make([]*node.Node, 0, len(cfg.Nodes))

	for _, nodeURLString := range cfg.Nodes {
		nodeURL, err := url.Parse(nodeURLString)
		if err != nil {
			logging.Logger.Printf("cannot parse node URL: %s", nodeURLString)
			continue
		}
		nodes = append(nodes, node.New(nodeURL, true, cfg, lb))
		logging.Logger.Printf("node added: %s", nodeURLString)
	}

	lb.serverPool = newServerPool(nodes, chk)
	(*alg).SetNodes(nodes)
	lb.algorithm = alg

	lb.StartPassiveHealthCheck(cfg.HealthCheck.Passive.Period)

	return lb
}
