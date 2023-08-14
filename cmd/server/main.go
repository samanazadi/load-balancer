package main

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/app"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"log"
	"net/http"
	"strconv"
)

func main() {
	// Logging
	logging.Init()

	// config
	config, err := configs.New()
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// checker
	var checker app.ConnectionChecker
	switch config.Checker.Name {
	case app.TCP:
		checker = app.TCPChecker{
			Timeout: config.HealthCheck.Passive.Timeout,
		}
		logging.Logger.Println("checker: TCP checker")
	case app.HTTP:
		path, keyPhrase := app.HTTPCheckerParamDecode(config.Checker.Params)
		checker = app.HTTPChecker{
			Path:      path,
			KeyPhrase: keyPhrase,
			Timeout:   config.HealthCheck.Passive.Timeout,
		}
		logging.Logger.Println("checker: HTTP checker")
	default:
		logging.Logger.Fatalf("invalid checker: %s", config.Checker.Name)
	}

	// strategy
	var strategy app.Strategy
	switch config.Strategy.Name {
	case app.RR:
		strategy = app.NewRoundRobin()
		logging.Logger.Println("strategy: round-robin")
	default:
		logging.Logger.Fatalf("invalid strategy: %s", config.Strategy.Name)
	}

	lb := app.NewLoadBalancer(config.Nodes, checker, strategy,
		config.HealthCheck.Active.MaxRetry, config.HealthCheck.Active.RetryDelay, config.HealthCheck.Passive.Period)
	http.Handle("/", lb)
	log.Printf("Load balancer started at port %d", config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Port), nil); err != nil {
		logging.Logger.Fatalf("cannot start load balancer: %s", err.Error())
	}
}
