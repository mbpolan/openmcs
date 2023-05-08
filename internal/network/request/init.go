package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const InitRequestHeader = 0x0E

// InitRequest is sent by the request when a new connection is first established.
type InitRequest struct {
	// NameHash contains a hashed value representing the player's username.
	NameHash byte
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *InitRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 1 byte for the hash of the player's name
	hash, err := r.Uint8()
	if err != nil {
		return fmt.Errorf("expected name hash packet: %s", err)
	}

	p.NameHash = hash
	return nil
}
