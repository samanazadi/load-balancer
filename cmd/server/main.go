package main

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/app"
	"github.com/samanazadi/load-balancer/pkg/logging"
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
	var chkr app.ConnectionChecker
	switch config.Checker.Name {
	case app.TCP:
		chkr = app.TCPChecker{}
		logging.Logger.Println("checker: TCP checker")
	case app.HTTP:
		chkr = app.HTTPChecker{}
		logging.Logger.Println("checker: HTTP checker")
	}

	// strategy
	var stgy app.Strategy
	switch config.Strategy.Name {
	case app.RR:
		stgy = app.NewRoundRobin()
		logging.Logger.Println("strategy: round-robin")
	}

	lb := app.NewLoadBalancer(config.Nodes, chkr, stgy,
		config.HealthCheck.Active.MaxRetry, config.HealthCheck.Active.RetryDelay, config.HealthCheck.Passive.Period)
	http.Handle("/", lb)
	log.Printf("Load balancer started at port %d", config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Port), nil); err != nil {
		logging.Logger.Fatalf("cannot start load balancer: %s", err.Error())
	}
}
