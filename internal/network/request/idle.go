package request

import "github.com/mbpolan/openmcs/internal/network"

const PlayerIdleRequestHeader byte = 0xCA

// PlayerIdleRequest is sent by the client to indicate the player has not interacted with the game in some time.
type PlayerIdleRequest struct {
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *PlayerIdleRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	return nil
}