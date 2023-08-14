package container

import (
	"github.com/samanazadi/load-balancer/configs"
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
			node.SetAlive(alive)
			if !alive {
				logging.Logger.Printf("Passive health check, node down: %s", node.URL.String())
			}
		}()
	}
	wg.Wait()
}

func (p *ServerPool) startPassiveHealthCheck() {
	go func() {
		logging.Logger.Printf("Passive health check daemon started")
		period := time.Second * time.Duration(configs.Config.HealthCheck.Passive.Period)
		t := time.NewTicker(period)
		for {
			select {
			case <-t.C:
				logging.Logger.Println("Passive health check is starting...")
				p.passiveHealthCheck()
				logging.Logger.Println("Passive health check completed")
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
