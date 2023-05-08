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

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *ChangeModesRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 1 byte for the public chat mode
	b, err := r.Uint8()
	if err != nil {
		return err
	}

	publicChatMode, ok := common.ChatModeCodes[b]
	if !ok {
		return fmt.Errorf("invalid public chat mode byte: %d", b)
	}

	// read 1 byte for the private chat mode
	b, err = r.Uint8()
	if err != nil {
		return err
	}

	privateChatMode, ok := common.ChatModeCodes[b]
	if !ok {
		return fmt.Errorf("invalid private chat mode byte: %d", b)
	}

	// read 1 byte for the interaction mode
	b, err = r.Uint8()
	if err != nil {
		return err
	}

	interactionMode, ok := common.InteractionModeCodes[b]
	if !ok {
		return fmt.Errorf("invalid interaction mode byte: %d", b)
	}

	p.PublicChat = publicChatMode
	p.PrivateChat = privateChatMode
	p.Interaction = interactionMode
	return nil
}
