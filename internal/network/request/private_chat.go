package request

import (
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

const PrivateChatRequestHeader byte = 0x7E

// PrivateChatRequest is sent by the client when a player sends a private chat message to another player.
type PrivateChatRequest struct {
	Recipient string
	Text      string
}

func ReadPrivateChatRequest(r *network.ProtocolReader) (*PrivateChatRequest, error) {
	// read 1 byte containing packet size
	size, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	// read 8 bytes containing the recipient's encoded name
	name, err := r.Uint64()
	if err != nil {
		return nil, err
	}

	// decode the recipient's username
	recipient, err := util.DecodeName(name)
	if err != nil {
		return nil, err
	}

	// read the encoded chat message, subtracting 8 bytes from the size since we already read past them
	length := int(size - 8)
	rawText := make([]byte, length)
	for i := 0; i < length; i++ {
		rawText[i], err = r.Uint8()
		if err != nil {
			return nil, err
		}
	}

	// decode the message
	text, err := util.DecodeChat(rawText)
	if err != nil {
		return nil, err
	}

	return &PrivateChatRequest{
		Recipient: recipient,
		Text:      text,
	}, nil
}
