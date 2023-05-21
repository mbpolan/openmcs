package interaction

import (
	"github.com/mbpolan/openmcs/internal/config"
)

// Manager provides access to various client-side interfaces and other interaction mechanisms.
type Manager struct {
	// CharacterDesigner is the interface for editing a player's appearance.
	CharacterDesigner *SimpleInterface
	// EquipmentTab is the interface for a player's equipment.
	EquipmentTab *EquipmentTabInterface
	// InventoryTab is the interface for a player's inventory.
	InventoryTab *InventoryTabInterface

	config config.InterfacesConfig
}

// New creates a new manager for interfaces and interactions.
func New(cfg config.InterfacesConfig) *Manager {
	return &Manager{
		config:            cfg,
		CharacterDesigner: newSimpleInterface(cfg.CharacterDesigner.ID),
		EquipmentTab:      newEquipmentTabInterface(cfg.Equipment),
		InventoryTab:      newInventoryTabInterface(cfg.Inventory),
	}
}
