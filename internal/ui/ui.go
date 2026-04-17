// Package ui provides the game UI system including HUD, inventory screen,
// pause menu, and main menu. It manages a stack of screens where the HUD
// is always at the bottom and overlay screens (inventory, pause) are pushed
// on top.
package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
)

// Screen is the interface that all UI screens implement.
type Screen interface {
	// Update handles per-frame logic such as input processing.
	Update(inp *input.Manager, dt float64)

	// Draw emits draw commands into the UIRenderer.
	Draw(r *UIRenderer)

	// IsOverlay reports whether this screen is drawn on top of another
	// screen (true) or replaces the screen below it (false).
	IsOverlay() bool

	// HandleKey processes a raw key code. This is used for screens that
	// need to respond to specific key events beyond what Update provides.
	HandleKey(key int)
}

// UIManager manages a stack of active UI screens. The bottom screen is
// typically the HUD, and overlay screens such as the inventory or pause
// menu are pushed on top.
type UIManager struct {
	screens []Screen
}

// NewUIManager creates a UIManager with no screens.
func NewUIManager() *UIManager {
	return &UIManager{
		screens: make([]Screen, 0, 4),
	}
}

// PushScreen adds a screen to the top of the stack.
func (m *UIManager) PushScreen(s Screen) {
	if s == nil {
		return
	}
	m.screens = append(m.screens, s)
}

// PopScreen removes the top screen from the stack and returns it.
// Returns nil if the stack is empty.
func (m *UIManager) PopScreen() Screen {
	n := len(m.screens)
	if n == 0 {
		return nil
	}
	top := m.screens[n-1]
	m.screens = m.screens[:n-1]
	return top
}

// CurrentScreen returns the topmost screen, or nil if the stack is empty.
func (m *UIManager) CurrentScreen() Screen {
	if len(m.screens) == 0 {
		return nil
	}
	return m.screens[len(m.screens)-1]
}

// ScreenCount returns the number of screens on the stack.
func (m *UIManager) ScreenCount() int {
	return len(m.screens)
}

// Update delegates the update call to the topmost screen.
func (m *UIManager) Update(inp *input.Manager, dt float64) {
	if s := m.CurrentScreen(); s != nil {
		s.Update(inp, dt)
	}
}

// Draw renders all visible screens. Non-overlay screens hide everything
// below them; overlay screens are drawn on top of the screens beneath.
func (m *UIManager) Draw(r *UIRenderer) {
	// Find the lowest visible screen: walk backwards until we find a
	// non-overlay screen (or hit the bottom).
	start := 0
	for i := len(m.screens) - 1; i >= 0; i-- {
		start = i
		if !m.screens[i].IsOverlay() {
			break
		}
	}
	for i := start; i < len(m.screens); i++ {
		m.screens[i].Draw(r)
	}
}

// IsBlockingInput reports whether the current top screen blocks game-world
// input (e.g. an inventory or pause menu is open). The HUD does not block
// input; menus and overlays do.
func (m *UIManager) IsBlockingInput() bool {
	s := m.CurrentScreen()
	if s == nil {
		return false
	}
	// HUD is never blocking; everything else is.
	_, isHUD := s.(*HUD)
	return !isHUD
}
