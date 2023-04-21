package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
	"github.com/mbpolan/openmcs/internal/util"
)

const PrivateChatResponseHeader byte = 0xC4

// PrivateChatResponse is sent by the server when a player receives a private chat message.
type PrivateChatResponse struct {
	id         int
	username   string
	senderType model.PlayerType
	text       string
}

// NewPrivateChatResponse creates a new private chat message response.
func NewPrivateChatResponse(id int, username string, senderType model.PlayerType, text string) *PrivateChatResponse {
	return &PrivateChatResponse{
		id:         id,
		username:   username,
		senderType: senderType,
		text:       text,
	}
}

// Write writes the contents of the message to a stream.
func (p *PrivateChatResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(PrivateChatResponseHeader)
	if err != nil {
		return err
	}

	// encode the sender's name and chat message
	encoded := util.EncodeChat(p.text)

	// write 1 byte for the size of the packet, which includes the encoded chat message and all following bytes
	size := len(encoded) + 13
	err = w.WriteUint8(byte(size))
	if err != nil {
		return err
	}

	// encode the sender's name and write it as 8 bytes
	name := util.EncodeName(p.username)
	err = w.WriteUint64(name)
	if err != nil {
		return err
	}

	// write 4 bytes for the message sequence id
	err = w.WriteUint32(uint32(p.id))
	if err != nil {
		return err
	}

	// write 1 byte for the sender's player type
	pType := common.PlayerTypeCode(p.senderType)
	err = w.WriteUint8(pType)
	if err != nil {
		return err
	}

	// write the bytes of the encoded chat message
	_, err = w.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}
