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
		rawText[i], err = r.Uint8()
		if err != nil {
			return nil, err
		}
	}

	text, err := decodeText(rawText)

	return &PlayerChatRequest{
		Effect: effect,
		Color:  color,
		Text:   text,
	}, nil
}

func decodeText(raw []byte) (string, error) {
	var text []byte

	lastCh := byte(0x00)
	for i := 0; i < len(raw); i++ {
		ch := raw[i]

		// each byte contains up to two distinct characters
		ch -= 0x80
		hi := ch >> 4
		lo := ch & 0x0F

		// if the last byte had a continuation, form a single code point from the two parts
		// if the high bits are a value greater than 13, treat the entire byte as a single code point
		// otherwise treat the high bits as a single code point
		code := -1
		if lastCh > 0 {
			code = int(((lastCh << 4) | hi) - 0xC3)
			text = append(text, util.ValidChatChars[code])

			lastCh = 0x00
		} else if hi >= 13 {
			code = int(ch - 0xC3)
			text = append(text, util.ValidChatChars[code])

			// skip the rest of the byte
			continue
		} else {
			code = int(hi)
			text = append(text, util.ValidChatChars[code])
		}

		// if the low bits are a value greater than 13, store them and expect the next byte to complete the code point
		// otherwise treat the low bits as a single code point
		if lo >= 13 {
			lastCh = lo
		} else {
			code = int(lo)
			text = append(text, util.ValidChatChars[code])
		}
	}

	// if some bits are left over, form a code point from them
	if lastCh > 0x00 {
		code := int(lastCh)
		text = append(text, util.ValidChatChars[code])
	}

	return string(text), nil
}
