package app

import (
	"context"
	"github.com/samanazadi/load-balancer/internal/algorithm"
	"github.com/samanazadi/load-balancer/internal/checker"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const RetryCount = iota

// LoadBalancer is a server pool along an algorithm
type LoadBalancer struct {
	serverPool ServerPool
	algorithm  *algorithm.Algorithm
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

func NewLoadBalancer(nodeURLStrings []string, checker checker.ConnectionChecker, alg *algorithm.Algorithm,
	maxRetry int, retryDelay int, period int) *LoadBalancer {
	lb := &LoadBalancer{}

	nodes := make([]*node.Node, 0, len(nodeURLStrings))
	for _, nodeURLString := range nodeURLStrings {
		nodeURL, err := url.Parse(nodeURLString)
		if err != nil {
			logging.Logger.Printf("cannot parse node URL: %s", nodeURLString)
			continue
		}

		rp := httputil.NewSingleHostReverseProxy(nodeURL)
		rp.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) { // Active health check
			retries := getRetryCountFromContext(r)
			logging.Logger.Printf("active health check, node down, %d retires: %s (%s)",
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
			logging.Logger.Printf("active health check, node down, retires exceeded: %s", nodeURL)
			lb.serverPool.setNodeAlive(nodeURL, false)
			newCtx := context.WithValue(r.Context(), RetryCount, 1)
			lb.ServeHTTP(rw, r.WithContext(newCtx))
		}

		n := node.Node{
			URL:          nodeURL,
			ReverseProxy: *rp,
		}
		n.SetAlive(true)
		nodes = append(nodes, &n)
		logging.Logger.Printf("node added: %s", nodeURLString)
	}

	lb.serverPool = newServerPool(nodes, checker)
	(*alg).SetNodes(nodes)
	lb.algorithm = alg

	lb.StartPassiveHealthCheck(period)

	return lb
}

func getRetryCountFromContext(r *http.Request) int {
	if retryCount, ok := r.Context().Value(RetryCount).(int); ok {
		return retryCount
	}
	return 1
}
