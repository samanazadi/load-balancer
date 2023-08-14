package main

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/container"
	"github.com/samanazadi/load-balancer/internal/logging"
	"log"
	"net/http"
	"strconv"
)

const roundRobin = "RR"

func main() {
	lb := container.NewLoadBalancer(configs.Config.Nodes)
	lb.StartPassiveHealthCheck()
	http.Handle("/", lb)
	log.Printf("Load balancer started at port %d", configs.Config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(configs.Config.Port), nil); err != nil {
		logging.Logger.Fatalf("Cannot start load balancer: %s", err.Error())
	}
}
