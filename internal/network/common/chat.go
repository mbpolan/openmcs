package common

import "github.com/mbpolan/openmcs/internal/model"

// ChatEffectCodes maps protocol identifiers to chat effects.
var ChatEffectCodes = map[byte]model.ChatEffect{
	0x00: model.ChatEffectNone,
	0x01: model.ChatEffectWave,
	0x02: model.ChatEffectWave2,
	0x03: model.ChatEffectShake,
	0x04: model.ChatEffectScroll,
	0x05: model.ChatEffectSlide,
}

// ChatColorCodes maps protocol identifiers to chat colors.
var ChatColorCodes = map[byte]model.ChatColor{
	0x00: model.ChatColorYellow,
	0x01: model.ChatColorRed,
	0x02: model.ChatColorGreen,
	0x03: model.ChatColorCyan,
	0x04: model.ChatColorPurple,
	0x05: model.ChatColorWhite,
	0x06: model.ChatColorFlash1,
	0x07: model.ChatColorFlash2,
	0x08: model.ChatColorFlash3,
	0x09: model.ChatColorGlow1,
	0x0A: model.ChatColorGlow2,
	0x0B: model.ChatColorGlow3,
}

func ChatEffectCode(c model.ChatEffect) byte {
	for k, v := range ChatEffectCodes {
		if v == c {
			return k
		}
	}

	return 0
}

func ChatColorCode(c model.ChatColor) byte {
	for k, v := range ChatColorCodes {
		if v == c {
			return k
		}
	}

	return 0
}
