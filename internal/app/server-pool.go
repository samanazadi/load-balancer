package app

import (
	"github.com/samanazadi/load-balancer/internal/checker"
	"github.com/samanazadi/load-balancer/internal/models/node"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/url"
	"sync"
	"time"
)

// ServerPool manage servers
type ServerPool struct {
	nodes             []*node.Node
	ConnectionChecker *checker.ConnectionChecker
}

func (p *ServerPool) passiveHealthCheck() {
	var wg sync.WaitGroup
	for _, n := range p.nodes {
		wg.Add(1)
		n := n
		go func() {
			defer wg.Done()
			alive := (*p.ConnectionChecker).Check(n.URL)
			if alive != n.IsAlive() {
				logging.Logger.Printf("passive health check, %s: %s -> %s",
					n.URL.String(), Alive(n.IsAlive()), Alive(alive))
			}
			n.SetAlive(alive)
		}()
	}
	wg.Wait()
}

func Alive(alive bool) string {
	if alive {
		return "up"
	}
	return "down"
}

func (p *ServerPool) startPassiveHealthCheck(period int) {
	go func() {
		logging.Logger.Printf("passive health check daemon started")
		period := time.Second * time.Duration(period)
		t := time.NewTicker(period)
		for {
			select {
			case <-t.C:
				logging.Logger.Println("passive health check is starting...")
				p.passiveHealthCheck()
				logging.Logger.Println("passive health check completed")
			}
		}
	}()
}

func (p *ServerPool) setNodeAlive(nodeURL *url.URL, alive bool) {
	for _, n := range p.nodes {
		if n.URL.String() == nodeURL.String() {
			n.SetAlive(alive)
			return
		}
	}
}

func newServerPool(nodes []*node.Node, chk *checker.ConnectionChecker) ServerPool {
	return ServerPool{
		nodes:             nodes,
		ConnectionChecker: chk,
	}
}
