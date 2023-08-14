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

const roundRobin = "RR"

func main() {
	// logging
	config, err := configs.New()
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// checker
	var chkr checker.ConnectionChecker
	switch config.Checker.Name {
	case checker.TCP:
		chkr = checker.TCPChecker{}
		logging.Logger.Println("checker: TCP checker")
	case checker.HTTP:
		chkr = checker.HTTPChecker{}
		logging.Logger.Println("checker: HTTP checker")
	}

	// strategy
	var stgy strategy.Strategy
	switch config.Strategy.Name {
	case strategy.RR:
		stgy = strategy.NewRoundRobin()
		logging.Logger.Println("strategy: round-robin")
	}

	lb := container.NewLoadBalancer(config.Nodes, chkr, stgy,
		config.HealthCheck.Active.MaxRetry, config.HealthCheck.Active.RetryDelay, config.HealthCheck.Passive.Period)
	http.Handle("/", lb)
	log.Printf("Load balancer started at port %d", config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Port), nil); err != nil {
		logging.Logger.Fatalf("cannot start load balancer: %s", err.Error())
	}
}
