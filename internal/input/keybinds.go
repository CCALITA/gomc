package input

import "sync"

// Action represents a game action that can be bound to a key.
type Action int

// Standard game actions.
const (
	MoveForward Action = iota
	MoveBackward
	MoveLeft
	MoveRight
	Jump
	Sneak
	Sprint
	Attack
	Use
	DropItem
	OpenInventory
	Pause
	ToggleDebug
	Hotbar1
	Hotbar2
	Hotbar3
	Hotbar4
	Hotbar5
	Hotbar6
	Hotbar7
	Hotbar8
	Hotbar9
)

// GLFW key constants (matching glfw.Key* values) so tests don't need GLFW.
const (
	KeyW           = 87
	KeyA           = 65
	KeyS           = 83
	KeyD           = 68
	KeyE           = 69
	KeyQ           = 81
	KeyF3          = 292
	KeySpace       = 32
	KeyLeftShift   = 340
	KeyLeftControl = 341
	KeyEscape      = 256
	Key1           = 49
	Key2           = 50
	Key3           = 51
	Key4           = 52
	Key5           = 53
	Key6           = 54
	Key7           = 55
	Key8           = 56
	Key9           = 57
)

// GLFW mouse button constants.
const (
	MouseButtonLeft  = 0
	MouseButtonRight = 1
)

// actionNames maps each Action to a human-readable name.
var actionNames = map[Action]string{
	MoveForward:   "Move Forward",
	MoveBackward:  "Move Backward",
	MoveLeft:      "Move Left",
	MoveRight:     "Move Right",
	Jump:          "Jump",
	Sneak:         "Sneak",
	Sprint:        "Sprint",
	Attack:        "Attack",
	Use:           "Use",
	DropItem:      "Drop Item",
	OpenInventory: "Open Inventory",
	Pause:         "Pause",
	ToggleDebug:   "Toggle Debug",
	Hotbar1:       "Hotbar 1",
	Hotbar2:       "Hotbar 2",
	Hotbar3:       "Hotbar 3",
	Hotbar4:       "Hotbar 4",
	Hotbar5:       "Hotbar 5",
	Hotbar6:       "Hotbar 6",
	Hotbar7:       "Hotbar 7",
	Hotbar8:       "Hotbar 8",
	Hotbar9:       "Hotbar 9",
}

// defaultBindings returns the standard WASD + modifier key mapping.
func defaultBindings() map[Action]int {
	return map[Action]int{
		MoveForward:   KeyW,
		MoveBackward:  KeyS,
		MoveLeft:      KeyA,
		MoveRight:     KeyD,
		Jump:          KeySpace,
		Sneak:         KeyLeftShift,
		Sprint:        KeyLeftControl,
		Attack:        MouseButtonLeft,  // note: button, not key
		Use:           MouseButtonRight, // note: button, not key
		DropItem:      KeyQ,
		OpenInventory: KeyE,
		Pause:         KeyEscape,
		ToggleDebug:   KeyF3,
		Hotbar1:       Key1,
		Hotbar2:       Key2,
		Hotbar3:       Key3,
		Hotbar4:       Key4,
		Hotbar5:       Key5,
		Hotbar6:       Key6,
		Hotbar7:       Key7,
		Hotbar8:       Key8,
		Hotbar9:       Key9,
	}
}

// KeyMap holds the mapping from game actions to key/button codes.
// It is safe for concurrent reads and writes.
type KeyMap struct {
	mu       sync.RWMutex
	bindings map[Action]int
}

// NewKeyMap returns a KeyMap initialised with the default bindings.
func NewKeyMap() *KeyMap {
	return &KeyMap{
		bindings: defaultBindings(),
	}
}

// GetKey returns the key code currently bound to the given action.
// Returns -1 if the action has no binding.
func (km *KeyMap) GetKey(action Action) int {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if key, ok := km.bindings[action]; ok {
		return key
	}
	return -1
}

// SetKey rebinds the given action to a new key code.
func (km *KeyMap) SetKey(action Action, key int) {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Create a new map to preserve immutability for any concurrent readers.
	updated := make(map[Action]int, len(km.bindings))
	for a, k := range km.bindings {
		updated[a] = k
	}
	updated[action] = key
	km.bindings = updated
}

// ActionName returns the human-readable name for the given action.
// Returns "Unknown" for unrecognised actions.
func ActionName(action Action) string {
	if name, ok := actionNames[action]; ok {
		return name
	}
	return "Unknown"
}
