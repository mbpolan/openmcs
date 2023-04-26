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

func ReadChatCommandRequest(r *network.ProtocolReader) (*ChatCommandRequest, error) {
	// read 1 byte for the string length
	b, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	// read the string itself
	text, err := r.String()
	if err != nil {
		return nil, err
	}

	// validate the string length matches the size
	if len(text) != int(b)-1 {
		return nil, fmt.Errorf("expected chat command string length %d, got %d", b-1, len(text))
	}

	return &ChatCommandRequest{
		Text: text,
	}, nil
}
