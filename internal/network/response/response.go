package response

import "github.com/mbpolan/openmcs/internal/network"

// A Response is sent by the server to confirm or instruct the client to perform an action.
type Response interface {
	// Write writes the contents of the message to a stream.
	Write(w *network.ProtocolWriter) error
}
