package common

import "github.com/mbpolan/openmcs/internal/model"

// PlayerTypeCodes maps protocol identifiers to player types.
var PlayerTypeCodes = map[byte]model.PlayerType{
	0x00: model.PlayerNormal,
	0x01: model.PlayerModerator,
	0x02: model.PlayerAdmin,
}

// PlayerTypeCode returns a protocol identifier for a player type.
func PlayerTypeCode(pType model.PlayerType) byte {
	for k, v := range PlayerTypeCodes {
		if v == pType {
			return k
		}
	}

	return 0x00
}
