package main

import (
	"flag"
	"fmt"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/server"
	"github.com/mbpolan/openmcs/internal/telemetry"
	"os"
	"os/signal"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config-dir", ".", "directory where server config.yaml is located")
	flag.Parse()

	if configPath == "" {
		fmt.Printf("-config-dir is required")
		os.Exit(1)
	}

	// load server configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("failed to load configuration: %s", err)
		os.Exit(1)
	}

	// set up logging
	err = logger.Setup(logger.Options{
		LogLevel: cfg.Server.LogLevel,
	})
	if err != nil {
		fmt.Printf("failed to initialize logger: %s", err)
		os.Exit(1)
	}

	// prepare the telemetry provider
	tel, err := telemetry.Setup(cfg)
	if err != nil {
		logger.Fatalf("failed to set up telemetry provider: %s", err)
	}

	// start the telemetry provider if enabled
	if cfg.Metrics.Enabled {
		tel.Start()
	}

	// setup up the game server
	srv, err := server.New(server.Options{
		Config:    cfg,
		Telemetry: tel,
	})
	if err != nil {
		logger.Fatalf("failed to prepare server: %s", err)
	}

	// prepare signal handlers
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		logger.Infof("received interrupt, stopping server")

		srv.Stop()
		_ = tel.Stop()
	}()

	// start the server
	err = srv.Run()
	if err != nil {
		logger.Fatalf("failed to start server: %s", err)
	}
}
