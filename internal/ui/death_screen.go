package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
)

const (
	deathBtnWidth  = 200.0
	deathBtnHeight = 40.0
)

// DeathScreen is displayed when the player's health reaches zero.
// It shows "You Died!" text and a "Respawn" button. Pressing Enter or
// clicking the button triggers the onRespawn callback.
type DeathScreen struct {
	onRespawn func()

	mouseX, mouseY float64
	hoveredBtn     bool

	screenWidth, screenHeight float32
}

// NewDeathScreen creates a DeathScreen with the given respawn callback.
func NewDeathScreen(onRespawn func()) *DeathScreen {
	return &DeathScreen{
		onRespawn: onRespawn,
	}
}

// Update handles input: Enter key or click on the Respawn button.
func (d *DeathScreen) Update(inp *input.Manager, _ float64) {
	d.mouseX, d.mouseY = inp.MousePos()
	d.hoveredBtn = d.hitTestButton(float32(d.mouseX), float32(d.mouseY))

	if inp.IsKeyJustPressed(input.KeyEnter) {
		d.respawn()
		return
	}

	if inp.IsMouseJustPressed(input.MouseButtonLeft) && d.hoveredBtn {
		d.respawn()
	}
}

// Draw renders the death screen overlay.
func (d *DeathScreen) Draw(r *UIRenderer) {
	// Dark red overlay.
	r.DrawRect(0, 0, r.ScreenWidth, r.ScreenHeight, 0.4, 0, 0, 0.6)

	// "You Died!" title.
	title := "You Died!"
	titleWidth := float32(len(title)) * 8 * 2.0
	r.DrawText(
		(r.ScreenWidth-titleWidth)/2,
		r.ScreenHeight/4,
		title,
		2.0,
		1, 0.2, 0.2,
	)

	// Respawn button.
	btnX := (r.ScreenWidth - deathBtnWidth) / 2
	btnY := r.ScreenHeight/2 + 20

	if d.hoveredBtn {
		r.DrawRect(btnX, btnY, deathBtnWidth, deathBtnHeight, 0.4, 0.4, 0.6, 0.9)
	} else {
		r.DrawRect(btnX, btnY, deathBtnWidth, deathBtnHeight, 0.3, 0.3, 0.3, 0.9)
	}

	label := "Respawn"
	labelWidth := float32(len(label)) * 8
	textX := btnX + (deathBtnWidth-labelWidth)/2
	textY := btnY + (deathBtnHeight-12)/2
	r.DrawText(textX, textY, label, 1.0, 1, 1, 1)
}

// IsOverlay returns false: the death screen replaces gameplay entirely.
func (d *DeathScreen) IsOverlay() bool {
	return false
}

// HandleKey processes a raw key code.
func (d *DeathScreen) HandleKey(key int) {
	if key == input.KeyEnter {
		d.respawn()
	}
}

// SetScreenSize sets the screen dimensions used for hit testing.
func (d *DeathScreen) SetScreenSize(w, h float32) {
	d.screenWidth = w
	d.screenHeight = h
}

// respawn triggers the respawn callback.
func (d *DeathScreen) respawn() {
	if d.onRespawn != nil {
		d.onRespawn()
	}
}

// hitTestButton returns true if the mouse is over the Respawn button.
func (d *DeathScreen) hitTestButton(mx, my float32) bool {
	w := d.screenWidth
	h := d.screenHeight
	if w == 0 {
		w = 800
	}
	if h == 0 {
		h = 600
	}

	btnX := (w - deathBtnWidth) / 2
	btnY := h/2 + 20

	return mx >= btnX && mx <= btnX+deathBtnWidth &&
		my >= btnY && my <= btnY+deathBtnHeight
}
