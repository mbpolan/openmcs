package common

import "github.com/mbpolan/openmcs/internal/model"

// ClientTabIndices maps a client tab to its protocol code.
var ClientTabIndices = map[model.ClientTab]byte{
	model.ClientTabEquippedItem: 0x00,
	model.ClientTabSkills:       0x01,
	model.ClientTabQuests:       0x02,
	model.ClientTabInventory:    0x03,
	model.ClientTabEquipment:    0x04,
	model.ClientTabPrayers:      0x05,
	model.ClientTabSpells:       0x06,
	model.ClientTabFriendsList:  0x08,
	model.ClientTabIgnoreList:   0x09,
	model.ClientTabLogout:       0x0A,
	model.ClientTabSettings:     0x0B,
	model.ClientTabControls:     0x0C,
	model.ClientTabMusic:        0x0D,
}
