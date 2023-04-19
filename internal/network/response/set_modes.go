package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
)

const SetModesResponseHeader byte = 0xCE

// SetModesResponse is sent to a client to update its public, private and interaction modes.
type SetModesResponse struct {
	publicChat  model.ChatMode
	privateChat model.ChatMode
	interaction model.InteractionMode
}

// NewSetModesResponse creates a new set mode response.
func NewSetModesResponse(publicChat, privateChat model.ChatMode, interaction model.InteractionMode) *SetModesResponse {
	return &SetModesResponse{
		publicChat:  publicChat,
		privateChat: privateChat,
		interaction: interaction,
	}
}

// Write writes the contents of the message to a stream.
func (p *SetModesResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(SetModesResponseHeader)
	if err != nil {
		return err
	}

	// write 1 byte for the public chat mode
	err = w.WriteUint8(common.ChatModeCode(p.publicChat))
	if err != nil {
		return err
	}

	// write 1 byte for the private chat mode
	err = w.WriteUint8(common.ChatModeCode(p.privateChat))
	if err != nil {
		return err
	}

	// write 1 byte for the interaction mode
	err = w.WriteUint8(common.InteractionModeCode(p.interaction))
	if err != nil {
		return err
	}

	return nil
}
