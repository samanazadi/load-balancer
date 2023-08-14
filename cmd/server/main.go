package main

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/app"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/http"
	"strconv"
)

func main() {
	// logging
	logging.Init()

	// config
	cfg, err := configs.New()
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// checker
	var checker app.ConnectionChecker
	switch cfg.Checker.Name {
	case app.TCP:
		checker = app.TCPChecker{
			Timeout: cfg.HealthCheck.Passive.Timeout,
		}
		logging.Logger.Println("checker: TCP checker")
	case app.HTTP:
		path, keyPhrase := app.HTTPCheckerParamDecode(cfg.Checker.Params)
		checker = app.HTTPChecker{
			Path:      path,
			KeyPhrase: keyPhrase,
			Timeout:   cfg.HealthCheck.Passive.Timeout,
		}
		logging.Logger.Println("checker: HTTP checker")
	default:
		logging.Logger.Fatalf("invalid checker: %s", cfg.Checker.Name)
	}

	// strategy
	var strategy app.Strategy
	switch cfg.Strategy.Name {
	case app.RR:
		strategy = app.NewRoundRobin()
		logging.Logger.Println("strategy: round-robin")
	default:
		logging.Logger.Fatalf("invalid strategy: %s", cfg.Strategy.Name)
	}

	// load balancer
	lb := app.NewLoadBalancer(cfg.Nodes, checker, strategy,
		cfg.HealthCheck.Active.MaxRetry, cfg.HealthCheck.Active.RetryDelay, cfg.HealthCheck.Passive.Period)
	logging.Logger.Println("load balancer created")

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: lb,
	}

	logging.Logger.Printf("load balancer started at port %d", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		logging.Logger.Fatalf("cannot start load balancer: %s", err.Error())
	}
}
