package container

import (
	"context"
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

func (lb *LoadBalancer) StartPassiveHealthCheck(period int) {
	lb.serverPool.startPassiveHealthCheck(period)
}

func NewLoadBalancer(nodeURLStrings []string, chkr checker.ConnectionChecker, stgy strategy.Strategy,
	maxRetry int, retryDelay int, period int) *LoadBalancer {
	lb := &LoadBalancer{}

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
			logging.Logger.Printf("Active health check, node down, %d retires: %s (%s)",
				retries, nodeURL, e.Error())
			if retries < maxRetry {
				// same node, more retries after some delay
				retryDelay := time.Millisecond * time.Duration(retryDelay)
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
	stgy.SetNodes(nodes)
	lb.strategy = stgy

	lb.StartPassiveHealthCheck(period)

	return lb
}

func getRetryCountFromContext(r *http.Request) int {
	if retryCount, ok := r.Context().Value(RetryCount).(int); ok {
		return retryCount
	}
	return 1
}
