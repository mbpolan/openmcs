package common

import "github.com/mbpolan/openmcs/internal/model"

// ChatModeCodes maps protocol codes to chat modes.
var ChatModeCodes = map[byte]model.ChatMode{
	0x00: model.ChatModePublic,
	0x01: model.ChatModeFriends,
	0x02: model.ChatModeOff,
	0x03: model.ChatModeHide,
}

// InteractionModeCodes maps protocol codes to interaction modes.
var InteractionModeCodes = map[byte]model.InteractionMode{
	0x00: model.InteractionModePublic,
	0x01: model.InteractionModeFriends,
	0x02: model.InteractionModeOff,
}

// ChatModeCode returns a protocol identifier for a chat mode.
func ChatModeCode(mode model.ChatMode) byte {
	for k, v := range ChatModeCodes {
		if v == mode {
			return k
		}
	}

	return 0x00
}

// InteractionModeCode returns a protocol identifier for an interaction mode.
func InteractionModeCode(mode model.InteractionMode) byte {
	for k, v := range InteractionModeCodes {
		if v == mode {
			return k
		}
	}

	return 0x00
}
