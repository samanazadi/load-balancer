package app

import (
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/url"
	"sync"
	"time"
)

type ServerPool struct {
	nodes []*Node
}

func (p *ServerPool) passiveHealthCheck() {
	var wg sync.WaitGroup
	for _, n := range p.nodes {
		wg.Add(1)
		n := n
		go func() {
			defer wg.Done()
			alive := n.CheckNodeAlive()
			if alive != n.IsAlive() {
				prev := "down"
				if n.IsAlive() {
					prev = "up"
				}

				now := "down"
				if alive {
					now = "up"
				}

				logging.Logger.Printf("passive health check, %s: %s -> %s", n.URL.String(), prev, now)
			}
			n.SetAlive(alive)
		}()
	}
	wg.Wait()
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

func newServerPool(nodes []*Node) ServerPool {
	return ServerPool{
		nodes: nodes,
	}
}