package interaction

import (
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/model"
)

// Manager provides access to various client-side interfaces and other interaction mechanisms.
type Manager struct {
	// EquipmentTab is the interface for a player's equipment.
	EquipmentTab *EquipmentTabInterface
	// FriendListTab is the interface for a player's friend list.
	FriendListTab *SimpleInterface
	// IgnoreListTab is the interface for a player's ignore list.
	IgnoreListTab *SimpleInterface
	// LogoutTab is the interface for a player's logout tab.
	LogoutTab *SimpleInterface
	// InventoryTab is the interface for a player's inventory.
	InventoryTab *InventoryTabInterface
	// SkillsTab is the interface for a player's skills.
	SkillsTab *SimpleInterface
	config    config.InterfacesConfig
}

// New creates a new manager for interfaces and interactions.
func New(cfg config.InterfacesConfig) *Manager {
	return &Manager{
		config:        cfg,
		EquipmentTab:  newEquipmentTabInterface(cfg.Equipment),
		FriendListTab: newSimpleInterface(cfg.FriendList.ID),
		IgnoreListTab: newSimpleInterface(cfg.IgnoreList.ID),
		LogoutTab:     newSimpleInterface(cfg.Logout.ID),
		InventoryTab:  newInventoryTabInterface(cfg.Inventory),
		SkillsTab:     newSimpleInterface(cfg.Skills.ID),
	}
}

// ClientTabInterfaces returns a map of client tab interfaces to the IDs of interfaces to render on the client.
func (m *Manager) ClientTabInterfaces() map[model.ClientTab]int {
	return map[model.ClientTab]int{
		model.ClientTabSkills:      m.SkillsTab.ID,
		model.ClientTabInventory:   m.InventoryTab.ID,
		model.ClientTabEquipment:   m.EquipmentTab.ID,
		model.ClientTabFriendsList: m.FriendListTab.ID,
		model.ClientTabIgnoreList:  m.IgnoreListTab.ID,
		model.ClientTabLogout:      m.LogoutTab.ID,
	}
}
