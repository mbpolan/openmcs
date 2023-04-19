package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

// chatModeCodes maps protocol codes to chat modes.
var chatModeCodes = map[byte]model.ChatMode{
	0x00: model.ChatModePublic,
	0x01: model.ChatModeFriends,
	0x02: model.ChatModeOff,
	0x03: model.ChatModeHide,
}

// interactionModeCodes maps protocol codes to interaction modes.
var interactionModeCodes = map[byte]model.InteractionMode{
	0x00: model.InteractionModePublic,
	0x01: model.InteractionModeFriends,
	0x02: model.InteractionModeOff,
}

const ChangeModesRequestHeader byte = 0x5F

// ChangeModesRequest is sent by the client when a player changes the public, private and/or interaction modes.
type ChangeModesRequest struct {
	PublicChat  model.ChatMode
	PrivateChat model.ChatMode
	Interaction model.InteractionMode
}

func ReadChangeModesRequest(r *network.ProtocolReader) (*ChangeModesRequest, error) {
	// read 1 byte for the public chat mode
	b, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	publicChatMode, ok := chatModeCodes[b]
	if !ok {
		return nil, fmt.Errorf("invalid public chat mode byte: %d", b)
	}

	// read 1 byte for the private chat mode
	b, err = r.Uint8()
	if err != nil {
		return nil, err
	}

	privateChatMode, ok := chatModeCodes[b]
	if !ok {
		return nil, fmt.Errorf("invalid private chat mode byte: %d", b)
	}

	// read 1 byte for the interaction mode
	b, err = r.Uint8()
	if err != nil {
		return nil, err
	}

	interactionMode, ok := interactionModeCodes[b]
	if !ok {
		return nil, fmt.Errorf("invalid interaction mode byte: %d", b)
	}

	return &ChangeModesRequest{
		PublicChat:  publicChatMode,
		PrivateChat: privateChatMode,
		Interaction: interactionMode,
	}, nil
}
