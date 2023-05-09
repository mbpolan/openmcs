package telemetry

import (
	"context"
	"fmt"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

// prometheusTelemetry is a metrics server backed by Prometheus.
type prometheusTelemetry struct {
	// bindAddress is the address the HTTP server will listen on.
	bindAddress string
	// enabled is used to determine if metrics collection is active.
	enabled bool
	// server is the HTTP server that provides endpoints for scraping Prometheus metrics.
	server *http.Server
	// usersOnlineGauge counts the current online player count.
	usersOnlineGauge prometheus.Gauge
}

// newPrometheusTelemetry creates a server for exposing Prometheus metrics.
func newPrometheusTelemetry(cfg *config.Config) (Telemetry, error) {
	bindAddress := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port)

	// expose an endpoint for serving up metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    bindAddress,
		Handler: mux,
	}

	// create metrics
	usersOnlineGauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "users_online_total",
		Help: "The total number of users connected to the server",
	})

	return &prometheusTelemetry{
		bindAddress:      bindAddress,
		server:           server,
		usersOnlineGauge: usersOnlineGauge,
	}, nil
}

// Start sets up the metrics server and begins exposing metrics. You must call this method before attempting to
// instrument the server. If the server cannot be started, a fatal error is logged and the process terminated.
func (p *prometheusTelemetry) Start() {
	go func() {
		p.enabled = true
		logger.Infof("metrics server listening on %s", p.bindAddress)

		err := p.server.ListenAndServe()
		if err != nil {
			logger.Fatalf("failed to start Prometheus metrics server: %s", err)
		}
	}()
}

// Stop gracefully terminates the metrics server.
func (p *prometheusTelemetry) Stop() error {
	p.enabled = false

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := p.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

// RecordPlayerConnected tracks that a player has connected to the server.
func (p *prometheusTelemetry) RecordPlayerConnected() {
	if !p.enabled {
		return
	}

	p.usersOnlineGauge.Inc()
}

// RecordPlayerDisconnected tracks that a player has disconnected from the server.
func (p *prometheusTelemetry) RecordPlayerDisconnected() {
	if !p.enabled {
		return
	}

	p.usersOnlineGauge.Dec()
}
