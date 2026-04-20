package ui

import (
	"fmt"
	"math"

	"github.com/fanxiyao/gomc/internal/input"
)

const (
	debugLineHeight = 14.0
	debugPadX       = 8.0
	debugPadY       = 8.0
	debugPanelWidth = 260.0
)

// DebugOverlay renders an F3-style debug information panel in the top-left
// corner of the screen. It is toggled with the F3 key and drawn on top of
// the game world as a permanent overlay.
type DebugOverlay struct {
	Visible                   bool
	FPS                       float64
	PlayerX, PlayerY, PlayerZ float32
	ChunkX, ChunkZ            int32
	BiomeName                 string
	Facing                    string
	PlayerYaw                 float32
	PlayerPitch               float32
	LoadedChunks              int
	EntityCount               int
	MemoryMB                  float64
}

// NewDebugOverlay creates a DebugOverlay that starts hidden.
func NewDebugOverlay() *DebugOverlay {
	return &DebugOverlay{}
}

// Update toggles visibility when the F3 key is pressed.
func (d *DebugOverlay) Update(inp *input.Manager, _ float64) {
	if inp.IsKeyJustPressed(input.KeyF3) {
		d.Visible = !d.Visible
	}
}

// Draw renders the debug info panel when visible.
func (d *DebugOverlay) Draw(r *UIRenderer) {
	if !d.Visible {
		return
	}

	lines := d.textLines()

	panelHeight := debugPadY*2 + float32(len(lines))*debugLineHeight

	// Semi-transparent dark background panel.
	r.DrawRect(0, 0, debugPanelWidth, panelHeight, 0, 0, 0, 0.5)

	// Draw each line of debug text.
	for i, line := range lines {
		if line == "" {
			continue
		}
		y := debugPadY + float32(i)*debugLineHeight
		r.DrawText(debugPadX, y, line, 1.0, 1, 1, 1)
	}
}

// IsOverlay returns true: the debug overlay renders on top of the game.
func (d *DebugOverlay) IsOverlay() bool {
	return true
}

// HandleKey is a no-op; input is handled in Update.
func (d *DebugOverlay) HandleKey(_ int) {}

// SetPlayerRotation updates yaw/pitch and recomputes the facing direction.
func (d *DebugOverlay) SetPlayerRotation(yaw, pitch float32) {
	d.PlayerYaw = yaw
	d.PlayerPitch = pitch
	d.Facing = FacingDirectionFromYaw(yaw)
}

// textLines returns the formatted debug info lines.
func (d *DebugOverlay) textLines() []string {
	return []string{
		fmt.Sprintf("GoMC 0.12 (FPS: %.0f)", d.FPS),
		fmt.Sprintf("XYZ: %.1f / %.1f / %.1f", d.PlayerX, d.PlayerY, d.PlayerZ),
		fmt.Sprintf("Chunk: %d %d Facing: %s", d.ChunkX, d.ChunkZ, d.Facing),
		fmt.Sprintf("Biome: %s", d.BiomeName),
		fmt.Sprintf("Chunks: %d Entities: %d", d.LoadedChunks, d.EntityCount),
		fmt.Sprintf("Memory: %.1f MB", d.MemoryMB),
	}
}

// FacingDirectionFromYaw returns a cardinal direction string for the given
// yaw angle in degrees. The mapping follows Minecraft conventions:
//
//	-45 to 45   -> South  (toward +Z)
//	 45 to 135  -> West   (toward -X)
//	135 to 180 or -180 to -135 -> North (toward -Z)
//	-135 to -45 -> East   (toward +X)
func FacingDirectionFromYaw(yaw float32) string {
	y := normalizeYawAngle(yaw)

	switch {
	case y >= -45 && y < 45:
		return "South"
	case y >= 45 && y < 135:
		return "West"
	case y >= -135 && y < -45:
		return "East"
	default:
		return "North"
	}
}

// normalizeYawAngle reduces a yaw angle to the range [-180, 180).
func normalizeYawAngle(yaw float32) float32 {
	return float32(math.Remainder(float64(yaw), 360))
}
