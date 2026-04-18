package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
)

const (
	mainBtnWidth   = 240.0
	mainBtnHeight  = 44.0
	mainBtnSpacing = 14.0
)

// mainMenuAction identifies what a main menu button does when clicked.
type mainMenuAction int

const (
	// mainMenuSingleplayer starts a singleplayer world.
	mainMenuSingleplayer mainMenuAction = iota
	// mainMenuMultiplayer opens the multiplayer browser (placeholder).
	mainMenuMultiplayer
	// mainMenuOptions opens the options screen (placeholder).
	mainMenuOptions
	// mainMenuQuit exits the game.
	mainMenuQuit
)

// MainMenuButton represents a clickable button on the main menu.
type mainMenuButton struct {
	Label  string
	Action mainMenuAction
}

// MainMenu is the first screen shown when the game starts. It provides
// Singleplayer, Multiplayer, Options, and Quit buttons.
type MainMenu struct {
	Buttons []mainMenuButton

	// Title text shown at the top.
	Title string

	// mouseX, mouseY track the cursor for hover highlighting.
	mouseX, mouseY float64

	// hoveredIndex is the index of the button under the cursor, or -1.
	hoveredIndex int

	// onAction is called when a button is clicked with its action.
	onAction func(mainMenuAction)

	// screenWidth, screenHeight for hit testing.
	screenWidth, screenHeight float32
}

// NewMainMenu creates the main menu with standard buttons.
func NewMainMenu(onAction func(mainMenuAction)) *MainMenu {
	return &MainMenu{
		Title: "GoMC",
		Buttons: []mainMenuButton{
			{Label: "Singleplayer", Action: mainMenuSingleplayer},
			{Label: "Multiplayer", Action: mainMenuMultiplayer},
			{Label: "Options", Action: mainMenuOptions},
			{Label: "Quit", Action: mainMenuQuit},
		},
		onAction:     onAction,
		hoveredIndex: -1,
	}
}

// Update handles mouse hover and click input.
func (m *MainMenu) Update(inp *input.Manager, _ float64) {
	m.mouseX, m.mouseY = inp.MousePos()
	m.hoveredIndex = m.hitTestButton(float32(m.mouseX), float32(m.mouseY))

	if inp.IsMouseJustPressed(input.MouseButtonLeft) && m.hoveredIndex >= 0 {
		m.executeButton(m.hoveredIndex)
	}
}

// Draw renders the main menu: background, title, and buttons.
func (m *MainMenu) Draw(r *UIRenderer) {
	// Full-screen background.
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0.1, 0.1, 0.15, 1.0)

	// Title at top.
	titleWidth := float32(len(m.Title)) * 8 * 2.5
	r.DrawText((r.ScreenWidth-titleWidth)/2, r.ScreenHeight*0.15, m.Title, 2.5, 1, 1, 1)

	// Subtitle.
	subtitle := "A Minecraft Clone in Go"
	subWidth := float32(len(subtitle)) * 8 * 0.8
	r.DrawText((r.ScreenWidth-subWidth)/2, r.ScreenHeight*0.15+50, subtitle, 0.8, 0.7, 0.7, 0.7)

	// Buttons centered vertically in the lower half.
	totalHeight := float32(len(m.Buttons))*(mainBtnHeight+mainBtnSpacing) - mainBtnSpacing
	startY := r.ScreenHeight*0.4 + (r.ScreenHeight*0.5-totalHeight)/2

	for i, btn := range m.Buttons {
		x := (r.ScreenWidth - mainBtnWidth) / 2
		y := startY + float32(i)*(mainBtnHeight+mainBtnSpacing)

		if i == m.hoveredIndex {
			r.DrawRect(x, y, mainBtnWidth, mainBtnHeight, 0.35, 0.55, 0.35, 0.95)
		} else {
			r.DrawRect(x, y, mainBtnWidth, mainBtnHeight, 0.25, 0.25, 0.25, 0.9)
		}

		labelWidth := float32(len(btn.Label)) * 8
		textX := x + (mainBtnWidth-labelWidth)/2
		textY := y + (mainBtnHeight-12)/2
		r.DrawText(textX, textY, btn.Label, 1.0, 1, 1, 1)
	}
}

// IsOverlay returns false: the main menu replaces whatever is below it.
func (m *MainMenu) IsOverlay() bool {
	return false
}

// HandleKey is a no-op for the main menu.
func (m *MainMenu) HandleKey(_ int) {}

// HoveredIndex returns the index of the currently hovered button, or -1.
func (m *MainMenu) HoveredIndex() int {
	return m.hoveredIndex
}

// SetScreenSize sets the screen dimensions used for hit testing.
func (m *MainMenu) SetScreenSize(w, h float32) {
	m.screenWidth = w
	m.screenHeight = h
}

// executeButton performs the action associated with the given button index.
func (m *MainMenu) executeButton(index int) {
	if index < 0 || index >= len(m.Buttons) {
		return
	}
	if m.onAction != nil {
		m.onAction(m.Buttons[index].Action)
	}
}

// hitTestButton returns the index of the button under (mx, my), or -1.
func (m *MainMenu) hitTestButton(mx, my float32) int {
	w := m.screenWidth
	h := m.screenHeight
	if w == 0 {
		w = 800
	}
	if h == 0 {
		h = 600
	}

	totalHeight := float32(len(m.Buttons))*(mainBtnHeight+mainBtnSpacing) - mainBtnSpacing
	startY := h*0.4 + (h*0.5-totalHeight)/2

	for i := range m.Buttons {
		x := (w - mainBtnWidth) / 2
		y := startY + float32(i)*(mainBtnHeight+mainBtnSpacing)
		if mx >= x && mx <= x+mainBtnWidth && my >= y && my <= y+mainBtnHeight {
			return i
		}
	}
	return -1
}
