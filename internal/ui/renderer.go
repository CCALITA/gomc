package ui

import (
	"fmt"

	"github.com/fanxiyao/gomc/internal/item"
)

// DrawCommand represents a single 2D draw operation collected by the
// UIRenderer for batch rendering.
type drawCommand struct {
	Type drawCommandType

	// Position and size in screen coordinates.
	X, Y, W, H float32

	// Color (used by Rect and Text commands).
	R, G, B, A float32

	// Texture coordinates (used by TexturedRect).
	U, V, UW, VH float32

	// Text content and scale (used by Text commands).
	Text  string
	Scale float32

	// Item data (used by ItemSlot commands).
	Stack item.ItemStack
}

// drawCommandType identifies the kind of draw command.
type drawCommandType int

const (
	// drawCmdRect is a solid-colour rectangle.
	drawCmdRect drawCommandType = iota
	// drawCmdTexturedRect is a textured rectangle from the atlas.
	drawCmdTexturedRect
	// drawCmdText is a bitmap-font text string.
	drawCmdText
	// drawCmdItemSlot is an item icon with an optional count label.
	drawCmdItemSlot
)

// UIRenderer is a 2D drawing abstraction that collects draw commands into
// a buffer for batch rendering. The actual GPU submission is handled by the
// render package; this struct only accumulates the vertex/command data.
type UIRenderer struct {
	ScreenWidth  float32
	ScreenHeight float32

	commands []drawCommand
}

// NewUIRenderer creates a UIRenderer targeting the given screen dimensions.
func NewUIRenderer(width, height float32) *UIRenderer {
	return &UIRenderer{
		ScreenWidth:  width,
		ScreenHeight: height,
		commands:     make([]drawCommand, 0, 256),
	}
}

// DrawRect adds a solid-colour rectangle command.
func (r *UIRenderer) DrawRect(x, y, w, h float32, cr, cg, cb, ca float32) {
	r.commands = append(r.commands, drawCommand{
		Type: drawCmdRect,
		X:    x, Y: y, W: w, H: h,
		R: cr, G: cg, B: cb, A: ca,
	})
}

// DrawTexturedRect adds a textured rectangle command. The texture coordinates
// (u, v, uw, vh) select a region from the texture atlas.
func (r *UIRenderer) DrawTexturedRect(x, y, w, h float32, u, v, uw, vh float32) {
	r.commands = append(r.commands, drawCommand{
		Type: drawCmdTexturedRect,
		X:    x, Y: y, W: w, H: h,
		U: u, V: v, UW: uw, VH: vh,
	})
}

// DrawText adds a text rendering command using a bitmap font.
func (r *UIRenderer) DrawText(x, y float32, text string, scale float32, cr, cg, cb float32) {
	r.commands = append(r.commands, drawCommand{
		Type:  drawCmdText,
		X:     x, Y: y,
		R:     cr, G: cg, B: cb, A: 1.0,
		Text:  text,
		Scale: scale,
	})
}

// DrawItemSlot adds a draw command for an item slot: the item icon and an
// optional count overlay.
func (r *UIRenderer) DrawItemSlot(x, y float32, stack item.ItemStack) {
	r.commands = append(r.commands, drawCommand{
		Type:  drawCmdItemSlot,
		X:     x, Y: y,
		Stack: stack,
	})
}

// Commands returns all collected draw commands. The slice is valid until
// the next call to Clear.
func (r *UIRenderer) Commands() []drawCommand {
	return r.commands
}

// CommandCount returns the number of queued draw commands.
func (r *UIRenderer) CommandCount() int {
	return len(r.commands)
}

// Clear resets the command buffer for the next frame.
func (r *UIRenderer) Clear() {
	r.commands = r.commands[:0]
}

// String returns a human-readable summary (useful for debugging).
func (r *UIRenderer) String() string {
	return fmt.Sprintf("UIRenderer(%.0fx%.0f, %d cmds)",
		r.ScreenWidth, r.ScreenHeight, len(r.commands))
}
