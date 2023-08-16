package main

import (
	"flag"
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

	// command line flags
	var cfgPath string //configs directory path
	flag.StringVar(&cfgPath, "c", "/etc/load-balancer", "configs directory")
	flag.Parse()

	// config
	cfg, err := configs.New(cfgPath)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// checker
	logging.Logger.Printf("checker: %s", cfg.Checker.Name)
	chk, err := checker.New(cfg)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// algorithm
	logging.Logger.Printf("algorithm: %s", cfg.Algorithm.Name)
	alg, err := algorithm.New(cfg)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// load balancer
	lb := app.New(cfg, chk, alg)
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
