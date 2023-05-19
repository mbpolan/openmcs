package interaction

import (
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/model"
)

// WeaponTabInterface is the set of interfaces for equipped weapons.
type WeaponTabInterface struct {
	config config.WeaponTabInterfaceConfig
}

// newWeaponTabInterface creates a new weapon tab interface manager.
func newWeaponTabInterface(cfg config.WeaponTabInterfaceConfig) *WeaponTabInterface {
	return &WeaponTabInterface{
		config: cfg,
	}
}

// IDForWeaponStyle returns an interface ID for a weapon based on its attack style.
func (t *WeaponTabInterface) IDForWeaponStyle(style model.WeaponStyle) int {
	switch style {
	case model.WeaponStyleNone:
		return 0
	case model.WeaponStyle2HSword:
		return t.config.TwoHandedSword
	case model.WeaponStyleAxe:
		return t.config.Axe
	case model.WeaponStyleBow:
		return t.config.Bow
	case model.WeaponStyleBlunt:
		return t.config.Blunt
	case model.WeaponStyleClaw:
		return t.config.Claws
	case model.WeaponStyleCrossbow:
		return t.config.Crossbow
	case model.WeaponStyleGun:
		return t.config.Gun
	case model.WeaponStylePickaxe:
		return t.config.Pickaxe
	case model.WeaponStylePoleArm:
		return t.config.PoleArm
	case model.WeaponStylePoleStaff:
		return t.config.PoleStaff
	case model.WeaponStyleScythe:
		return t.config.Scythe
	case model.WeaponStyleSlashSword:
		return t.config.SlashSword
	case model.WeaponStyleSpear:
		return t.config.Spear
	case model.WeaponStyleSpiked:
		return t.config.Spiked
	case model.WeaponStyleStabSword:
		return t.config.StabSword
	case model.WeaponStyleStaff:
		return t.config.Staff
	case model.WeaponStyleThrown:
		return t.config.Thrown
	case model.WeaponStyleWhip:
		return t.config.Whip
	}

	return 0
}
