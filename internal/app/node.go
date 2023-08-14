package app

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

// Node is a single backend server
type Node struct {
	URL               *url.URL
	alive             bool
	ReverseProxy      httputil.ReverseProxy
	ConnectionChecker ConnectionChecker
	mux               sync.RWMutex
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

func (n *Node) CheckNodeAlive() bool {
	return n.ConnectionChecker.Check(n.URL)
}
