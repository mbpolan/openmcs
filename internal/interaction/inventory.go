package interaction

import (
	"github.com/mbpolan/openmcs/internal/config"
)

// InventoryTabInterface is the interface used for displaying a player's inventory.
type InventoryTabInterface struct {
	// ID is the identifier for the parent interface.
	ID int
	// SlotsID is the identifier for the interface responsible for displaying inventory slots.
	SlotsID int
	config  config.InventoryTabInterfaceConfig
}

// newInventoryTabInterface creates a new inventory tab interface manager.
func newInventoryTabInterface(cfg config.InventoryTabInterfaceConfig) *InventoryTabInterface {
	return &InventoryTabInterface{
		ID:      cfg.ID,
		SlotsID: cfg.Slots,
		config:  cfg,
	}
}
