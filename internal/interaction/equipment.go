package interaction

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
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

// Update refreshes the interface for a player.
func (t *EquipmentTabInterface) Update(p *model.Player) []response.Response {
	var responses []response.Response

	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Attack.Stab, t.formatStat("Stab", p.CombatStats.Attack.Stab)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Attack.Slash, t.formatStat("Slash", p.CombatStats.Attack.Slash)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Attack.Crush, t.formatStat("Crush", p.CombatStats.Attack.Crush)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Attack.Magic, t.formatStat("Magic", p.CombatStats.Attack.Magic)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Attack.Range, t.formatStat("Range", p.CombatStats.Attack.Range)))

	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Defense.Stab, t.formatStat("Stab", p.CombatStats.Defense.Stab)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Defense.Slash, t.formatStat("Slash", p.CombatStats.Defense.Slash)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Defense.Crush, t.formatStat("Crush", p.CombatStats.Defense.Crush)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Defense.Magic, t.formatStat("Magic", p.CombatStats.Defense.Magic)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Bonuses.Defense.Range, t.formatStat("Range", p.CombatStats.Defense.Range)))

	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Other.Strength, t.formatStat("Strength", p.CombatStats.Strength)))
	responses = append(responses, response.NewSetInterfaceTextResponse(t.config.Other.Prayer, t.formatStat("Prayer", p.CombatStats.Prayer)))

	return responses
}

// formatStat formats a combat statistic.
func (t *EquipmentTabInterface) formatStat(name string, n int) string {
	if n > 0 {
		return fmt.Sprintf("%s: +%d", name, n)
	}

	return fmt.Sprintf("%s: %d", name, n)
}
