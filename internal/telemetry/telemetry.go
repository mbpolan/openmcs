package telemetry

import "github.com/mbpolan/openmcs/internal/config"

// Telemetry is an interface for metrics providers that instrument the game server.
type Telemetry interface {
	// Start enables the metrics server.
	Start()
	// Stop terminates the metrics server.
	Stop() error
	// RecordGameStateUpdateDuration records the duration a game state update took to complete.
	RecordGameStateUpdateDuration(duration float64)
	// RecordPlayerConnected tracks a player that was connected to the server.
	RecordPlayerConnected()
	// RecordPlayerDisconnected tracks a player that was disconnected from the server.
	RecordPlayerDisconnected()
}

// Setup creates a new metrics provider. If the provider cannot be created, an error is returned. You must call Start()
// on the returned Telemetry provider to initialize it, and call Drop() to clean up resources.
func Setup(cfg *config.Config) (Telemetry, error) {
	return newPrometheusTelemetry(cfg)
}
