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

// ReadInitRequest parses the packet from the connection stream.
func ReadInitRequest(r *network.ProtocolReader) (*InitRequest, error) {
	// first and only byte is the hash of the player's name
	hash, err := r.Uint8()
	if err != nil {
		return nil, fmt.Errorf("expected name hash packet: %s", err)
	}

	return &InitRequest{
		NameHash: hash,
	}, nil
}
