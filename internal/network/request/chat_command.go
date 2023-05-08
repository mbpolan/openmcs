package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const ChatCommandRequestHeader byte = 0x67

// ChatCommandRequest is sent by the client when a player enters a chat command.
type ChatCommandRequest struct {
	Text string
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *ChatCommandRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 1 byte for the string length
	b, err := r.Uint8()
	if err != nil {
		return err
	}

	// read the string itself
	text, err := r.String()
	if err != nil {
		return err
	}

	// validate the string length matches the size
	if len(text) != int(b)-1 {
		return fmt.Errorf("expected chat command string length %d, got %d", b-1, len(text))
	}

	p.Text = text
	return nil
}
