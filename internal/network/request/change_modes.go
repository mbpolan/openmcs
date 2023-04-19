package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
)

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

	publicChatMode, ok := common.ChatModeCodes[b]
	if !ok {
		return nil, fmt.Errorf("invalid public chat mode byte: %d", b)
	}

	// read 1 byte for the private chat mode
	b, err = r.Uint8()
	if err != nil {
		return nil, err
	}

	privateChatMode, ok := common.ChatModeCodes[b]
	if !ok {
		return nil, fmt.Errorf("invalid private chat mode byte: %d", b)
	}

	// read 1 byte for the interaction mode
	b, err = r.Uint8()
	if err != nil {
		return nil, err
	}

	interactionMode, ok := common.InteractionModeCodes[b]
	if !ok {
		return nil, fmt.Errorf("invalid interaction mode byte: %d", b)
	}

	return &ChangeModesRequest{
		PublicChat:  publicChatMode,
		PrivateChat: privateChatMode,
		Interaction: interactionMode,
	}, nil
}
