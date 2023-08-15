package node

import (
	"context"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

const RetryCount = iota

// Node is a single backend server
type Node struct {
	URL          *url.URL
	alive        bool
	ReverseProxy *httputil.ReverseProxy
	mux          sync.RWMutex // for protecting alive
}

func (n *Node) SetAlive(alive bool) {
	n.mux.Lock()
	defer n.mux.Unlock()
	n.alive = alive
}

func (n *Node) IsAlive() bool {
	n.mux.RLock()
	defer n.mux.RUnlock()
	return n.alive
}

type LB interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	SetNodeAlive(*url.URL, bool)
}

func New(url *url.URL, alive bool, cfg *configs.Config, lb LB) *Node {
	rp := httputil.NewSingleHostReverseProxy(url)
	rp.ErrorHandler = newReverseProxyErrorHandler(cfg, lb, url, rp)
	n := &Node{
		URL:          url,
		ReverseProxy: rp,
	}
	n.SetAlive(alive)
	return n
}

func newReverseProxyErrorHandler(cfg *configs.Config, lb LB, url *url.URL, rp *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request, error) {
	return func(rw http.ResponseWriter, r *http.Request, e error) { // Active health check
		retries := getRetryCountFromContext(r)
		logging.Logger.Printf("active health check, node down, %d retires: %s (%s)",
			retries, url, e.Error())
		if retries < cfg.HealthCheck.Active.MaxRetry {
			// same node, more retries after some delay
			retryDelay := time.Millisecond * time.Duration(cfg.HealthCheck.Active.RetryDelay)
			select {
			case <-time.After(retryDelay):
				ctx := context.WithValue(r.Context(), RetryCount, retries+1)
				rp.ServeHTTP(rw, r.WithContext(ctx))
			}
			return
		}

		// max retries exceeded
		logging.Logger.Printf("active health check, node down, retires exceeded: %s", url)
		lb.SetNodeAlive(url, false)
		newCtx := context.WithValue(r.Context(), RetryCount, 1)
		lb.ServeHTTP(rw, r.WithContext(newCtx))
	}
}

func getRetryCountFromContext(r *http.Request) int {
	if retryCount, ok := r.Context().Value(RetryCount).(int); ok {
		return retryCount
	}
	return 1
}
