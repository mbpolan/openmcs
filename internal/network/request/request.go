package request

import "github.com/mbpolan/openmcs/internal/network"

// A Request is a command sent by the player's client to perform an action or convey a data packet.
type Request interface {
	// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
	Read(r *network.ProtocolReader) error
}
