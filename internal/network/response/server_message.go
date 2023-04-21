package response

import "github.com/mbpolan/openmcs/internal/network"

const ServerMessageResponseHeader byte = 0xFD

// ServerMessageResponse is sent by the server to convey a game message to the client.
type ServerMessageResponse struct {
	message string
}

// NewServerMessageResponse creates a new server message response.
func NewServerMessageResponse(message string) *ServerMessageResponse {
	return &ServerMessageResponse{
		message: message,
	}
}

// Write writes the contents of the message to a stream.
func (p *ServerMessageResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(ServerMessageResponseHeader)
	if err != nil {
		return err
	}

	// write one byte containing the packet size (length of message plus one terminating byte)
	err = w.WriteUint8(byte(len(p.message) + 1))
	if err != nil {
		return err
	}

	// write the message text as a string
	err = w.WriteString(p.message)
	if err != nil {
		return err
	}

	return nil
}
