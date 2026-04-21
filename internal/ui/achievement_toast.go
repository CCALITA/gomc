package ui

import (
	"github.com/fanxiyao/gomc/internal/input"
)

const (
	toastDuration = 5.0 // seconds
	toastWidth    = 280.0
	toastHeight   = 50.0
	toastPadding  = 10.0
)

// AchievementToast displays an "Achievement Get!" notification that
// automatically dismisses after 5 seconds. It implements the Screen
// interface as an overlay.
type AchievementToast struct {
	name    string
	elapsed float64
	done    bool
}

// NewAchievementToast creates a toast notification for the given
// achievement name.
func NewAchievementToast(name string) *AchievementToast {
	return &AchievementToast{
		name: name,
	}
}

// Update advances the toast timer. After toastDuration seconds the toast
// marks itself as done.
func (t *AchievementToast) Update(_ *input.Manager, dt float64) {
	t.elapsed += dt
	if t.elapsed >= toastDuration {
		t.done = true
	}
}

// Draw renders the toast in the upper-right area of the screen.
func (t *AchievementToast) Draw(r *UIRenderer) {
	if t.done {
		return
	}

	x := r.ScreenWidth - toastWidth - toastPadding
	y := float32(toastPadding)

	// Background.
	r.DrawRect(x, y, toastWidth, toastHeight, 0.1, 0.1, 0.1, 0.85)

	// Title line.
	title := "Achievement Get!"
	r.DrawText(x+10, y+8, title, 1.0, 1.0, 0.84, 0.0)

	// Achievement name.
	r.DrawText(x+10, y+28, t.name, 1.0, 1, 1, 1)
}

// IsOverlay returns true; the toast is drawn on top of the game world.
func (t *AchievementToast) IsOverlay() bool {
	return true
}

// HandleKey is a no-op for the toast.
func (t *AchievementToast) HandleKey(_ int) {}

// IsDone reports whether the toast has finished displaying.
func (t *AchievementToast) IsDone() bool {
	return t.done
}
