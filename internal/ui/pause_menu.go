package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
)

const (
	pauseBtnWidth   = 200.0
	pauseBtnHeight  = 40.0
	pauseBtnSpacing = 12.0
)

// ButtonAction identifies what a pause menu button does when clicked.
type ButtonAction int

const (
	// ButtonResume closes the pause menu and returns to gameplay.
	ButtonResume ButtonAction = iota
	// ButtonOptions opens the options screen (placeholder).
	ButtonOptions
	// ButtonQuit exits the game.
	ButtonQuit
)

// PauseButton represents a clickable button in the pause menu.
type PauseButton struct {
	Label  string
	Action ButtonAction
}

// PauseMenu is displayed when the player presses Escape during gameplay.
// It provides Resume, Options, and Quit buttons.
type PauseMenu struct {
	Buttons []PauseButton

	// mouseX, mouseY track the cursor for hover highlighting.
	mouseX, mouseY float64

	// closed indicates the menu should be removed.
	closed bool

	// onClose is called when the menu is dismissed.
	onClose func()

	// onQuit is called when the Quit button is clicked.
	onQuit func()

	// hoveredIndex is the index of the button under the cursor, or -1.
	hoveredIndex int

	// screenWidth, screenHeight for hit testing.
	screenWidth, screenHeight float32
}

// NewPauseMenu creates a pause menu with the standard buttons.
func NewPauseMenu(onClose func(), onQuit func()) *PauseMenu {
	return &PauseMenu{
		Buttons: []PauseButton{
			{Label: "Resume", Action: ButtonResume},
			{Label: "Options", Action: ButtonOptions},
			{Label: "Quit", Action: ButtonQuit},
		},
		onClose:      onClose,
		onQuit:       onQuit,
		hoveredIndex: -1,
	}
}

// Update handles input: Escape to close, mouse hover, and click.
func (m *PauseMenu) Update(inp *input.Manager, _ float64) {
	m.mouseX, m.mouseY = inp.MousePos()
	m.hoveredIndex = m.hitTestButton(float32(m.mouseX), float32(m.mouseY))

	// Close on Escape.
	if inp.IsKeyJustPressed(input.KeyEscape) {
		m.Close()
		return
	}

	// Handle click.
	if inp.IsMouseJustPressed(input.MouseButtonLeft) && m.hoveredIndex >= 0 {
		m.executeButton(m.hoveredIndex)
	}
}

// Draw renders the pause menu overlay with buttons.
func (m *PauseMenu) Draw(r *UIRenderer) {
	// Dark overlay.
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0, 0, 0, 0.5)

	// Title.
	titleText := "Game Paused"
	// Approximate text centering (8 pixels per character at scale 1.5).
	titleWidth := float32(len(titleText)) * 8 * 1.5
	r.DrawText((r.ScreenWidth-titleWidth)/2, r.ScreenHeight/4, titleText, 1.5, 1, 1, 1)

	// Buttons.
	totalHeight := float32(len(m.Buttons))*(pauseBtnHeight+pauseBtnSpacing) - pauseBtnSpacing
	startY := (r.ScreenHeight - totalHeight) / 2

	for i, btn := range m.Buttons {
		x := (r.ScreenWidth - pauseBtnWidth) / 2
		y := startY + float32(i)*(pauseBtnHeight+pauseBtnSpacing)

		// Button background with hover highlight.
		if i == m.hoveredIndex {
			r.DrawRect(x, y, pauseBtnWidth, pauseBtnHeight, 0.4, 0.4, 0.6, 0.9)
		} else {
			r.DrawRect(x, y, pauseBtnWidth, pauseBtnHeight, 0.3, 0.3, 0.3, 0.9)
		}

		// Button label (centered).
		labelWidth := float32(len(btn.Label)) * 8
		textX := x + (pauseBtnWidth-labelWidth)/2
		textY := y + (pauseBtnHeight-12)/2
		r.DrawText(textX, textY, btn.Label, 1.0, 1, 1, 1)
	}
}

// IsOverlay returns true: the pause menu is drawn over the HUD.
func (m *PauseMenu) IsOverlay() bool {
	return true
}

// HandleKey processes a raw key code.
func (m *PauseMenu) HandleKey(key int) {
	if key == input.KeyEscape {
		m.Close()
	}
}

// IsClosed reports whether the pause menu has been dismissed.
func (m *PauseMenu) IsClosed() bool {
	return m.closed
}

// Close dismisses the pause menu.
func (m *PauseMenu) Close() {
	if m.closed {
		return
	}
	m.closed = true
	if m.onClose != nil {
		m.onClose()
	}
}

// SetScreenSize sets the screen dimensions used for hit testing.
func (m *PauseMenu) SetScreenSize(w, h float32) {
	m.screenWidth = w
	m.screenHeight = h
}

// HoveredIndex returns the index of the currently hovered button, or -1.
func (m *PauseMenu) HoveredIndex() int {
	return m.hoveredIndex
}

// executeButton performs the action associated with a button index.
func (m *PauseMenu) executeButton(index int) {
	if index < 0 || index >= len(m.Buttons) {
		return
	}
	switch m.Buttons[index].Action {
	case ButtonResume:
		m.Close()
	case ButtonOptions:
		// Options screen is a placeholder for now.
	case ButtonQuit:
		if m.onQuit != nil {
			m.onQuit()
		}
	}
}

// hitTestButton returns the index of the button under (mx, my), or -1.
func (m *PauseMenu) hitTestButton(mx, my float32) int {
	w := m.screenWidth
	h := m.screenHeight
	if w == 0 {
		w = 800
	}
	if h == 0 {
		h = 600
	}

	totalHeight := float32(len(m.Buttons))*(pauseBtnHeight+pauseBtnSpacing) - pauseBtnSpacing
	startY := (h - totalHeight) / 2

	for i := range m.Buttons {
		x := (w - pauseBtnWidth) / 2
		y := startY + float32(i)*(pauseBtnHeight+pauseBtnSpacing)
		if mx >= x && mx <= x+pauseBtnWidth && my >= y && my <= y+pauseBtnHeight {
			return i
		}
	}
	return -1
}
