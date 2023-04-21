package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
	"github.com/mbpolan/openmcs/internal/util"
)

const PlayerChatRequestHeader byte = 0x04

// PlayerChatRequest is sent by the client when a player sends chat message.
type PlayerChatRequest struct {
	Effect model.ChatEffect
	Color  model.ChatColor
	Text   string
}

func ReadPlayerChatRequest(r *network.ProtocolReader) (*PlayerChatRequest, error) {
	// read 1 byte for the packet size
	size, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	// read 1 byte for the chat effect
	effectCode, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	effect, ok := common.ChatEffectCodes[0x80-effectCode]
	if !ok {
		return nil, fmt.Errorf("unknown chat effect code: %d", effectCode)
	}

	// read 1 byte for the chat color
	colorCode, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	color, ok := common.ChatColorCodes[0x80-colorCode]
	if !ok {
		return nil, fmt.Errorf("unknown chat color code: %d", colorCode)
	}

	// read bytes corresponding to the text message itself. the packet size includes the two bytes used for chat
	// effect and color, so we subtract that from the total length. the bytes themselves are written in reverse order.
	length := int(size - 2)
	rawText := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		b, err := r.Uint8()
		if err != nil {
			return nil, err
		}

		rawText[i] = b - 0x80
	}

	text, err := util.DecodeChat(rawText)
	if err != nil {
		return nil, err
	}

	return &PlayerChatRequest{
		Effect: effect,
		Color:  color,
		Text:   text,
	}, nil
}
