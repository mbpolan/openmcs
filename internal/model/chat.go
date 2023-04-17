package model

// ChatEffect is a visual effect applied to chat messages.
type ChatEffect int

const (
	ChatEffectNone ChatEffect = iota
	ChatEffectWave
	ChatEffectWave2
	ChatEffectShake
	ChatEffectScroll
	ChatEffectSlide
)

// ChatColor is a color applied to a chat message.
type ChatColor int

const (
	ChatColorYellow ChatColor = iota
	ChatColorRed
	ChatColorGreen
	ChatColorCyan
	ChatColorPurple
	ChatColorWhite
	ChatColorFlash1
	ChatColorFlash2
	ChatColorFlash3
	ChatColorGlow1
	ChatColorGlow2
	ChatColorGlow3
)
