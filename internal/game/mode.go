// Package game provides game state, mode management, and the main game loop.
package game

// GameMode represents the current game mode (Survival, Creative, Spectator).
type GameMode int

const (
	// ModeSurvival is the default mode with normal gameplay mechanics.
	ModeSurvival GameMode = 0
	// ModeCreative enables flying, instant break, infinite items, and no damage.
	ModeCreative GameMode = 1
	// ModeSpectator enables noclip flying, prevents interaction, and makes the player invisible.
	ModeSpectator GameMode = 2
)

// modeProperties stores the immutable capability flags for each game mode.
type modeProperties struct {
	canTakeDamage    bool
	canBreakInstant  bool
	hasInfiniteItems bool
	canFly           bool
	isNoClip         bool
}

// modeTable maps each GameMode to its property set. Adding a new mode only
// requires one entry here rather than scattered switch statements.
var modeTable = map[GameMode]modeProperties{
	ModeSurvival: {
		canTakeDamage:    true,
		canBreakInstant:  false,
		hasInfiniteItems: false,
		canFly:           false,
		isNoClip:         false,
	},
	ModeCreative: {
		canTakeDamage:    false,
		canBreakInstant:  true,
		hasInfiniteItems: true,
		canFly:           true,
		isNoClip:         false,
	},
	ModeSpectator: {
		canTakeDamage:    false,
		canBreakInstant:  false,
		hasInfiniteItems: false,
		canFly:           true,
		isNoClip:         true,
	},
}

// ModeManager tracks the current game mode and exposes capability queries.
type ModeManager struct {
	CurrentMode GameMode
}

// NewModeManager creates a ModeManager initialised to Survival mode.
func NewModeManager() *ModeManager {
	return &ModeManager{CurrentMode: ModeSurvival}
}

// SetMode switches to the given game mode.
func (m *ModeManager) SetMode(mode GameMode) {
	m.CurrentMode = mode
}

// props returns the property set for the current mode.
// Falls back to Survival defaults for unknown modes.
func (m *ModeManager) props() modeProperties {
	p, ok := modeTable[m.CurrentMode]
	if !ok {
		return modeTable[ModeSurvival]
	}
	return p
}

// CanTakeDamage reports whether the player receives damage in the current mode.
func (m *ModeManager) CanTakeDamage() bool {
	return m.props().canTakeDamage
}

// CanBreakInstantly reports whether blocks break instantly in the current mode.
func (m *ModeManager) CanBreakInstantly() bool {
	return m.props().canBreakInstant
}

// HasInfiniteItems reports whether the player has unlimited items in the current mode.
func (m *ModeManager) HasInfiniteItems() bool {
	return m.props().hasInfiniteItems
}

// CanFly reports whether the player can fly in the current mode.
func (m *ModeManager) CanFly() bool {
	return m.props().canFly
}

// IsNoClip reports whether the player passes through blocks in the current mode.
func (m *ModeManager) IsNoClip() bool {
	return m.props().isNoClip
}
