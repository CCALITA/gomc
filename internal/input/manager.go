// Package input provides keyboard and mouse input management for the game.
// It tracks key states, mouse button states, mouse position, mouse delta,
// and scroll delta. GLFW key/button constants are used as plain ints so that
// tests can run without an active GLFW context.
package input

import "sync"

// Key/button action constants matching GLFW values.
const (
	ActionRelease = 0
	ActionPress   = 1
	ActionRepeat  = 2
)

// keyState tracks the current and previous frame state for a single key.
type keyState struct {
	current  bool
	previous bool
}

// buttonState tracks the current and previous frame state for a mouse button.
type buttonState struct {
	current  bool
	previous bool
}

// Manager tracks all input state: keys, mouse buttons, cursor position,
// cursor delta, and scroll delta. It is safe for concurrent use from
// GLFW callbacks (which run on the main thread) and game-logic goroutines.
type Manager struct {
	mu sync.RWMutex

	keys    map[int]*keyState
	buttons map[int]*buttonState

	mouseX, mouseY         float64
	lastMouseX, lastMouseY float64
	deltaX, deltaY         float64
	firstMouse             bool

	scrollX, scrollY float64
}

// NewManager returns an initialised Manager ready to receive callbacks.
func NewManager() *Manager {
	return &Manager{
		keys:       make(map[int]*keyState),
		buttons:    make(map[int]*buttonState),
		firstMouse: true,
	}
}

// ---------- GLFW callback handlers ----------

// KeyCallback should be registered as the GLFW key callback.
// Parameters match the GLFW signature: key, scancode, action, mods.
func (m *Manager) KeyCallback(key, scancode, action, mods int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ks, ok := m.keys[key]
	if !ok {
		ks = &keyState{}
		m.keys[key] = ks
	}

	switch action {
	case ActionPress, ActionRepeat:
		ks.current = true
	case ActionRelease:
		ks.current = false
	}
}

// MouseButtonCallback should be registered as the GLFW mouse button callback.
func (m *Manager) MouseButtonCallback(button, action, mods int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bs, ok := m.buttons[button]
	if !ok {
		bs = &buttonState{}
		m.buttons[button] = bs
	}

	switch action {
	case ActionPress:
		bs.current = true
	case ActionRelease:
		bs.current = false
	}
}

// CursorPosCallback should be registered as the GLFW cursor position callback.
func (m *Manager) CursorPosCallback(x, y float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.firstMouse {
		m.lastMouseX = x
		m.lastMouseY = y
		m.firstMouse = false
	}

	m.mouseX = x
	m.mouseY = y
}

// ScrollCallback should be registered as the GLFW scroll callback.
// Scroll offsets accumulate until the next Update call.
func (m *Manager) ScrollCallback(xoff, yoff float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scrollX += xoff
	m.scrollY += yoff
}

// ---------- Per-frame update ----------

// Update advances the input state by one frame. It must be called exactly
// once per frame, after all queries for the current frame have been made.
//
// It promotes "just pressed" keys to "pressed" and "just released" keys to
// "released" by copying current into previous. It also computes the mouse
// delta and resets the scroll accumulator.
func (m *Manager) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Advance key states: previous catches up to current.
	for _, ks := range m.keys {
		ks.previous = ks.current
	}

	// Advance button states.
	for _, bs := range m.buttons {
		bs.previous = bs.current
	}

	// Compute mouse delta from last-known position.
	m.deltaX = m.mouseX - m.lastMouseX
	m.deltaY = m.mouseY - m.lastMouseY
	m.lastMouseX = m.mouseX
	m.lastMouseY = m.mouseY

	// Reset scroll accumulators.
	m.scrollX = 0
	m.scrollY = 0
}

// ---------- Query helpers ----------

// IsKeyDown returns true while the key is held down.
func (m *Manager) IsKeyDown(key int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ks, ok := m.keys[key]; ok {
		return ks.current
	}
	return false
}

// IsKeyJustPressed returns true only during the frame when the key
// transitioned from released to pressed (current=true, previous=false).
func (m *Manager) IsKeyJustPressed(key int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ks, ok := m.keys[key]; ok {
		return ks.current && !ks.previous
	}
	return false
}

// IsKeyJustReleased returns true only during the frame when the key
// transitioned from pressed to released (current=false, previous=true).
func (m *Manager) IsKeyJustReleased(key int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ks, ok := m.keys[key]; ok {
		return !ks.current && ks.previous
	}
	return false
}

// IsMouseDown returns true while the given mouse button is held down.
func (m *Manager) IsMouseDown(button int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if bs, ok := m.buttons[button]; ok {
		return bs.current
	}
	return false
}

// IsMouseJustPressed returns true only during the frame when the button
// transitioned from released to pressed.
func (m *Manager) IsMouseJustPressed(button int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if bs, ok := m.buttons[button]; ok {
		return bs.current && !bs.previous
	}
	return false
}

// MousePos returns the current cursor position.
func (m *Manager) MousePos() (x, y float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.mouseX, m.mouseY
}

// MouseDelta returns the cursor movement since the last Update call.
func (m *Manager) MouseDelta() (dx, dy float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.deltaX, m.deltaY
}

// ScrollDelta returns the accumulated vertical scroll offset since the last
// Update call. Positive values indicate scrolling up.
func (m *Manager) ScrollDelta() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.scrollY
}
