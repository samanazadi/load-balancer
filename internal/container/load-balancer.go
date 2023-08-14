package container

import (
	"context"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/checker"
	"github.com/samanazadi/load-balancer/internal/logging"
	"github.com/samanazadi/load-balancer/internal/node"
	"github.com/samanazadi/load-balancer/internal/strategy"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const RetryCount = iota

// LoadBalancer is a server pool along a strategy
type LoadBalancer struct {
	serverPool ServerPool
	strategy   strategy.Strategy
}

func (lb *LoadBalancer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if n := lb.strategy.GetNextEligibleNode(r); n != nil {
		n.ReverseProxy.ServeHTTP(rw, r)
		return
	}
	logging.Logger.Println("No node is available")
	http.Error(rw, "Service not available", http.StatusServiceUnavailable)
}

func (lb *LoadBalancer) StartPassiveHealthCheck() {
	lb.serverPool.startPassiveHealthCheck()
}

func NewLoadBalancer(nodeURLStrings []string) *LoadBalancer {
	lb := &LoadBalancer{}

	var chkr checker.ConnectionChecker
	switch configs.Config.Checker.Name {
	case checker.TCP:
		chkr = checker.TCPChecker{}
		logging.Logger.Println("Checker: TCP checker")
	case checker.HTTP:
		chkr = checker.HTTPChecker{}
		logging.Logger.Println("Checker: HTTP checker")
	}

	nodes := make([]*node.Node, 0, len(nodeURLStrings))
	for _, nodeURLString := range nodeURLStrings {
		nodeURL, err := url.Parse(nodeURLString)
		if err != nil {
			logging.Logger.Printf("Cannot parse node URL: %s", nodeURLString)
			continue
		}

		rp := httputil.NewSingleHostReverseProxy(nodeURL)
		rp.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) { // Active health check
			retries := getRetryCountFromContext(r)
			logging.Logger.Printf("Active health check, node down, %d retires: %s (%s)", retries, nodeURL, e.Error())
			if retries < configs.Config.HealthCheck.Active.MaxRetry {
				// same node, more retries after some delay
				retryDelay := time.Millisecond * time.Duration(configs.Config.HealthCheck.Active.RetryDelay)
				select {
				case <-time.After(retryDelay):
					ctx := context.WithValue(r.Context(), RetryCount, retries+1)
					rp.ServeHTTP(rw, r.WithContext(ctx))
				}
				return
			}

			// max retries exceeded
			logging.Logger.Printf("Active health check, node down, retires exceeded: %s", nodeURL)
			lb.serverPool.setNodeAlive(nodeURL, false)
			newCtx := context.WithValue(r.Context(), RetryCount, 1)
			lb.ServeHTTP(rw, r.WithContext(newCtx))
		}

		n := node.Node{
			URL:               nodeURL,
			ReverseProxy:      *rp,
			ConnectionChecker: chkr,
		}
		n.SetAlive(true)
		nodes = append(nodes, &n)
		logging.Logger.Printf("Node added: %s", nodeURLString)
	}

	lb.serverPool = newServerPool(nodes)

	var stgy strategy.Strategy
	switch configs.Config.Strategy.Name {
	case strategy.RR:
		stgy = strategy.NewRoundRobin(nodes)
		logging.Logger.Println("Strategy: round-robin")
	}
	lb.strategy = stgy

	lb.StartPassiveHealthCheck()

	return lb
}

func getRetryCountFromContext(r *http.Request) int {
	if retryCount, ok := r.Context().Value(RetryCount).(int); ok {
		return retryCount
	}
	return 1
}
