package main

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/checker"
	"github.com/samanazadi/load-balancer/internal/container"
	"github.com/samanazadi/load-balancer/internal/logging"
	"github.com/samanazadi/load-balancer/internal/strategy"
	"log"
	"net/http"
	"strconv"
)

func main() {
	lb := container.NewLoadBalancer(configs.Config.Nodes, checker.HTTPChecker{}, strategy.RR)
	lb.StartPassiveHealthCheck()
	http.Handle("/", lb)
	log.Printf("Load balancer started at port %d", configs.Config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(configs.Config.Port), nil); err != nil {
		logging.Logger.Fatalf("Cannot start load balancer: %s", err.Error())
	}
}
