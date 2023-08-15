package main

import (
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/algorithm"
	"github.com/samanazadi/load-balancer/internal/app"
	"github.com/samanazadi/load-balancer/internal/checker"
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
	var chk checker.ConnectionChecker
	switch cfg.Checker.Name {
	case checker.TCPType:
		chk = checker.TCP{
			Timeout: cfg.HealthCheck.Passive.Timeout,
		}
		logging.Logger.Println("checker: TCP checker")
	case checker.HTTPType:
		path, keyPhrase := checker.HTTPCheckerParamDecode(cfg.Checker.Params)
		chk = checker.HTTP{
			Path:      path,
			KeyPhrase: keyPhrase,
			Timeout:   cfg.HealthCheck.Passive.Timeout,
		}
		logging.Logger.Println("checker: HTTP checker")
	default:
		logging.Logger.Fatalf("invalid checker: %s", cfg.Checker.Name)
	}

	// algorithm
	logging.Logger.Printf("algorithm: %s", cfg.Algorithm.Name)
	alg, err := algorithm.New(cfg)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// load balancer
	lb := app.NewLoadBalancer(cfg.Nodes, chk, alg,
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
