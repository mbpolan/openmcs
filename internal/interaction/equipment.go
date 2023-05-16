package interaction

import (
	"github.com/mbpolan/openmcs/internal/config"
)

// EquipmentTabInterface is the interface used for displaying a player's equipment.
type EquipmentTabInterface struct {
	// ID is the identifier for the parent interface.
	ID int
	// SlotsID is the identifier for the interface responsible for displaying equipment slots.
	SlotsID int
	config  config.EquipmentTabInterfaceConfig
}

// newEquipmentTabInterface creates a new equipment tab interface manager.
func newEquipmentTabInterface(cfg config.EquipmentTabInterfaceConfig) *EquipmentTabInterface {
	return &EquipmentTabInterface{
		ID:      cfg.ID,
		SlotsID: cfg.Slots,
		config:  cfg,
	}
}
