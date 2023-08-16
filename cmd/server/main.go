package main

import (
	"context"
	"errors"
	"flag"
	"github.com/samanazadi/load-balancer/configs"
	"github.com/samanazadi/load-balancer/internal/algorithm"
	"github.com/samanazadi/load-balancer/internal/app"
	"github.com/samanazadi/load-balancer/internal/checker"
	"github.com/samanazadi/load-balancer/pkg/logging"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	// logging
	logging.Init()
	logging.Logger.Printf("starting load balancer...")

	// command line flags
	var cfgPath string //configs directory path
	flag.StringVar(&cfgPath, "c", "/etc/load-balancer", "configs directory")
	flag.Parse()
	logging.Logger.Printf("configs directory: %s", cfgPath)

	// config
	cfg, err := configs.New(cfgPath)
	if err != nil {
		logging.Logger.Fatal(err)
	}
	logging.Logger.Printf("configs loaded")

	// checker
	logging.Logger.Printf("checker: %s", cfg.Checker.Name)
	chk, err := checker.New(cfg)
	if err != nil {
		logging.Logger.Fatal(err)
	}
	logging.Logger.Printf("checker created")

	// algorithm
	logging.Logger.Printf("algorithm: %s", cfg.Algorithm.Name)
	alg, err := algorithm.New(cfg)
	if err != nil {
		logging.Logger.Fatal(err)
	}
	logging.Logger.Printf("algorithm created")

	// load balancer
	stopPHC := make(chan bool, 1) // passive health check
	defer close(stopPHC)
	donePHC := make(chan bool, 1) // passive health check
	defer close(donePHC)
	lb := app.New(cfg, chk, alg, stopPHC, donePHC)
	logging.Logger.Println("load balancer created")

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: lb,
	}
	go func() {
		logging.Logger.Printf("load balancer started at port %d", cfg.Port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logging.Logger.Printf("cannot start load balancer: %s", err.Error())
		}
	}()

	// graceful shutdown
	sigs := make(chan os.Signal, 1)
	defer close(sigs)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	stopPHC <- true

	logging.Logger.Print("awaiting load balancer to stop")
	err = server.Shutdown(context.Background())
	if err != nil {
		logging.Logger.Printf("load balancer stopped with error: %s", err)
	} else {
		logging.Logger.Print("load balancer stopped")
	}

	logging.Logger.Print("awaiting passive health check to stop")
	<-donePHC
	logging.Logger.Print("passive health check stopped")
}
