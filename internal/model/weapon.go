package model

// AttackStyle enumerates possible weapon attack styles an entity may use.
type AttackStyle int

const (
	AttackStyleChop AttackStyle = iota
	AttackStyleSlash
	AttackStyleLunge
	AttackStyleBlock
	AttackStylePunch
	AttackStyleKick
	AttackStylePound
	AttackStylePummel
	AttackStyleSpike
	AttackStyleImpale
	AttackStyleSmash
	AttackStyleJab
	AttackStyleSwipe
	AttackStyleFend
	AttackStyleBash
	AttackStyleReap
	AttackStyleFlick
	AttackStyleLash
	AttackStyleDeflect
	AttackStyleAccurate
	AttackStyleRapid
	AttackStyleLongRange
	AttackStyleArmAndFire
	AttackStyleFocus
	AttackStyleStab
)

// InitAttackStyleMap returns a map of weapon styles and default attack styles.
func InitAttackStyleMap() map[WeaponStyle]AttackStyle {
	m := map[WeaponStyle]AttackStyle{}
	m[WeaponStyle2HSword] = AttackStyleChop
	m[WeaponStyleAxe] = AttackStyleChop
	m[WeaponStyleBow] = AttackStyleAccurate
	m[WeaponStyleBlunt] = AttackStylePound
	m[WeaponStyleClaw] = AttackStyleChop
	m[WeaponStyleCrossbow] = AttackStyleAccurate
	m[WeaponStyleGun] = AttackStyleArmAndFire
	m[WeaponStylePickaxe] = AttackStyleSpike
	m[WeaponStylePoleArm] = AttackStyleJab
	m[WeaponStylePoleStaff] = AttackStyleBash
	m[WeaponStyleScythe] = AttackStyleReap
	m[WeaponStyleSlashSword] = AttackStyleChop
	m[WeaponStyleSpear] = AttackStyleLunge
	m[WeaponStyleSpiked] = AttackStylePound
	m[WeaponStyleStabSword] = AttackStyleStab
	m[WeaponStyleStaff] = AttackStyleBash
	m[WeaponStyleThrown] = AttackStyleAccurate
	m[WeaponStyleWhip] = AttackStyleFlick
	m[WeaponStyleUnarmed] = AttackStylePunch

	return m
}
