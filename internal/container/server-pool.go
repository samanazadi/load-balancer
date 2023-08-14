package container

import (
	"github.com/samanazadi/load-balancer/internal/logging"
	"github.com/samanazadi/load-balancer/internal/node"
	"net/url"
	"sync"
	"time"
)

type ServerPool struct {
	nodes []*node.Node
}

func (p *ServerPool) passiveHealthCheck() {
	var wg sync.WaitGroup
	for _, node := range p.nodes {
		wg.Add(1)
		node := node
		go func() {
			defer wg.Done()
			alive := node.CheckNodeAlive()
			if alive != node.IsAlive() {
				prev := "down"
				if node.IsAlive() {
					prev = "up"
				}

				now := "down"
				if alive {
					now = "up"
				}

				logging.Logger.Printf("passive health check, %s: %s -> %s", node.URL.String(), prev, now)
			}
			node.SetAlive(alive)
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
	for _, node := range p.nodes {
		if node.URL.String() == nodeURL.String() {
			node.SetAlive(alive)
			return
		}
	}
}

func newServerPool(nodes []*node.Node) ServerPool {
	return ServerPool{
		nodes: nodes,
	}
}
