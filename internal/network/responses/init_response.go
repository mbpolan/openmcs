package responses

import "github.com/mbpolan/openmcs/internal/network"

// InitFailureCode enumerates various error conditions that result in a failed initialization.
type InitFailureCode byte

const (
	InitInvalidUsername InitFailureCode = 0x03
)

// InitResponse is sent by the responses in response to a requests's initialization request.
type InitResponse struct {
	code byte
}

// NewFailedInitResponse creates a response indicating the player's login was rejected.
func NewFailedInitResponse(code InitFailureCode) *InitResponse {
	return &InitResponse{code: byte(code)}
}

// Write writes the contents of the message to a stream.
func (p *InitResponse) Write(w *network.ProtocolWriter) error {
	// write 8 bytes (ignored by requests)
	_, err := w.Write(make([]byte, 8))
	if err != nil {
		return err
	}

	// write the result code first
	err = w.WriteByte(p.code)
	if err != nil {
		return err
	}

	return w.Flush()
}
