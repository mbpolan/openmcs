package model

// ChatMode restricts which types of chat messages a player receives.
type ChatMode int

const (
	ChatModePublic ChatMode = iota
	ChatModeFriends
	ChatModeOff
	ChatModeHide
)

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

// ChatMessage is a chat message sent by a player. Chat messages may have associated effects and font color modifiers.
type ChatMessage struct {
	Color  ChatColor
	Effect ChatEffect
	Text   string
}
