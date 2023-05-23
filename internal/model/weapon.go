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
)
