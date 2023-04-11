package main

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/server"
	"os"
	"os/signal"
)

func main() {
	err := logger.Setup()
	if err != nil {
		fmt.Printf("failed to initialize logger: %s", err)
		os.Exit(1)
	}

	// TODO: provide these parameters via flags
	srv, err := server.New(server.Options{
		AssetDir: "data",
		Address:  "127.0.0.1",
		Port:     43594,
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
	}()

	err = srv.Run()
	if err != nil {
		logger.Fatalf("failed to start server: %s", err)
	}
}
