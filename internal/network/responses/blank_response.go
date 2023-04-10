package responses

import "github.com/mbpolan/openmcs/internal/network"

// BlankResponse sends a sequence of zero bytes to the client.
type BlankResponse struct {
	padding int
}

// NewBlankResponse creates a response with zero byte padding.
func NewBlankResponse(padding int) *BlankResponse {
	return &BlankResponse{padding: padding}
}

// Write writes the contents of the message to a stream.
func (p *BlankResponse) Write(w *network.ProtocolWriter) error {
	for i := 0; i < p.padding; i++ {
		err := w.WriteByte(0x00)
		if err != nil {
			return err
		}
	}

	return nil
}
