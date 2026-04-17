package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Manager: key lifecycle
// ---------------------------------------------------------------------------

func TestKeyPressReleaseCycle(t *testing.T) {
	m := NewManager()

	// Before any input, key should not be down.
	assert.False(t, m.IsKeyDown(KeyW))
	assert.False(t, m.IsKeyJustPressed(KeyW))
	assert.False(t, m.IsKeyJustReleased(KeyW))

	// Press key — just-pressed is true immediately after callback.
	m.KeyCallback(KeyW, 0, ActionPress, 0)

	assert.True(t, m.IsKeyDown(KeyW))
	assert.True(t, m.IsKeyJustPressed(KeyW), "should be just pressed right after callback")
	assert.False(t, m.IsKeyJustReleased(KeyW))

	// Update advances previous = current; just-pressed goes away.
	m.Update()
	assert.True(t, m.IsKeyDown(KeyW), "key still held after update")
	assert.False(t, m.IsKeyJustPressed(KeyW), "just pressed should not survive Update")
	assert.False(t, m.IsKeyJustReleased(KeyW))

	// Release key — just-released is true immediately.
	m.KeyCallback(KeyW, 0, ActionRelease, 0)

	assert.False(t, m.IsKeyDown(KeyW))
	assert.False(t, m.IsKeyJustPressed(KeyW))
	assert.True(t, m.IsKeyJustReleased(KeyW), "should be just released right after callback")

	// Update advances state; just-released goes away.
	m.Update()
	assert.False(t, m.IsKeyDown(KeyW))
	assert.False(t, m.IsKeyJustPressed(KeyW))
	assert.False(t, m.IsKeyJustReleased(KeyW), "just released should not survive Update")
}

func TestJustPressedOnlyLastsOneFrame(t *testing.T) {
	m := NewManager()

	m.KeyCallback(KeySpace, 0, ActionPress, 0)
	assert.True(t, m.IsKeyJustPressed(KeySpace))

	// Update ends the frame — just-pressed should be gone.
	m.Update()
	assert.False(t, m.IsKeyJustPressed(KeySpace), "just pressed must not persist beyond one frame")
}

func TestKeyRepeatCountsAsDown(t *testing.T) {
	m := NewManager()

	m.KeyCallback(KeyA, 0, ActionRepeat, 0)
	assert.True(t, m.IsKeyDown(KeyA), "repeat action should count as key down")
}

func TestUnknownKeyReturnsFalse(t *testing.T) {
	m := NewManager()
	assert.False(t, m.IsKeyDown(9999))
	assert.False(t, m.IsKeyJustPressed(9999))
	assert.False(t, m.IsKeyJustReleased(9999))
}

// ---------------------------------------------------------------------------
// Manager: mouse buttons
// ---------------------------------------------------------------------------

func TestMouseButtonPressRelease(t *testing.T) {
	m := NewManager()

	assert.False(t, m.IsMouseDown(MouseButtonLeft))
	assert.False(t, m.IsMouseJustPressed(MouseButtonLeft))

	m.MouseButtonCallback(MouseButtonLeft, ActionPress, 0)

	assert.True(t, m.IsMouseDown(MouseButtonLeft))
	assert.True(t, m.IsMouseJustPressed(MouseButtonLeft))

	// Update ends the frame.
	m.Update()
	assert.True(t, m.IsMouseDown(MouseButtonLeft))
	assert.False(t, m.IsMouseJustPressed(MouseButtonLeft))

	// Release.
	m.MouseButtonCallback(MouseButtonLeft, ActionRelease, 0)
	assert.False(t, m.IsMouseDown(MouseButtonLeft))
}

func TestUnknownMouseButtonReturnsFalse(t *testing.T) {
	m := NewManager()
	assert.False(t, m.IsMouseDown(99))
	assert.False(t, m.IsMouseJustPressed(99))
}

// ---------------------------------------------------------------------------
// Manager: mouse position and delta
// ---------------------------------------------------------------------------

func TestMouseDeltaCalculation(t *testing.T) {
	m := NewManager()

	// First cursor callback initialises position without producing delta.
	m.CursorPosCallback(100, 200)
	m.Update()

	dx, dy := m.MouseDelta()
	assert.Equal(t, 0.0, dx, "first callback should produce zero delta X")
	assert.Equal(t, 0.0, dy, "first callback should produce zero delta Y")

	x, y := m.MousePos()
	assert.Equal(t, 100.0, x)
	assert.Equal(t, 200.0, y)

	// Move the mouse.
	m.CursorPosCallback(150, 220)
	m.Update()

	dx, dy = m.MouseDelta()
	assert.Equal(t, 50.0, dx)
	assert.Equal(t, 20.0, dy)

	x, y = m.MousePos()
	assert.Equal(t, 150.0, x)
	assert.Equal(t, 220.0, y)

	// No movement: delta should be zero.
	m.Update()
	dx, dy = m.MouseDelta()
	assert.Equal(t, 0.0, dx)
	assert.Equal(t, 0.0, dy)
}

// ---------------------------------------------------------------------------
// Manager: scroll
// ---------------------------------------------------------------------------

func TestScrollAccumulation(t *testing.T) {
	m := NewManager()

	// Multiple scroll events between frames should accumulate.
	m.ScrollCallback(0, 1.0)
	m.ScrollCallback(0, 0.5)
	m.ScrollCallback(0, -0.25)

	// ScrollDelta reads the accumulated value before Update resets it.
	assert.Equal(t, 1.25, m.ScrollDelta())

	// Update resets the accumulator.
	m.Update()
	assert.Equal(t, 0.0, m.ScrollDelta(), "scroll delta should reset after Update")
}

func TestScrollDeltaDefaultsToZero(t *testing.T) {
	m := NewManager()
	assert.Equal(t, 0.0, m.ScrollDelta())
}

// ---------------------------------------------------------------------------
// KeyMap: defaults and set/get
// ---------------------------------------------------------------------------

func TestKeybindDefaults(t *testing.T) {
	km := NewKeyMap()

	assert.Equal(t, KeyW, km.GetKey(MoveForward))
	assert.Equal(t, KeyS, km.GetKey(MoveBackward))
	assert.Equal(t, KeyA, km.GetKey(MoveLeft))
	assert.Equal(t, KeyD, km.GetKey(MoveRight))
	assert.Equal(t, KeySpace, km.GetKey(Jump))
	assert.Equal(t, KeyLeftShift, km.GetKey(Sneak))
	assert.Equal(t, KeyLeftControl, km.GetKey(Sprint))
	assert.Equal(t, KeyQ, km.GetKey(DropItem))
	assert.Equal(t, KeyE, km.GetKey(OpenInventory))
	assert.Equal(t, KeyEscape, km.GetKey(Pause))
	assert.Equal(t, KeyF3, km.GetKey(ToggleDebug))
	assert.Equal(t, Key1, km.GetKey(Hotbar1))
	assert.Equal(t, Key9, km.GetKey(Hotbar9))
}

func TestKeybindSetGet(t *testing.T) {
	km := NewKeyMap()

	// Rebind forward to up-arrow (GLFW key 265).
	km.SetKey(MoveForward, 265)
	assert.Equal(t, 265, km.GetKey(MoveForward))

	// Other bindings should be unaffected.
	assert.Equal(t, KeyS, km.GetKey(MoveBackward))
}

func TestGetKeyUnboundAction(t *testing.T) {
	km := NewKeyMap()
	assert.Equal(t, -1, km.GetKey(Action(999)), "unbound action should return -1")
}

func TestActionName(t *testing.T) {
	assert.Equal(t, "Move Forward", ActionName(MoveForward))
	assert.Equal(t, "Jump", ActionName(Jump))
	assert.Equal(t, "Hotbar 1", ActionName(Hotbar1))
	assert.Equal(t, "Unknown", ActionName(Action(999)))
}

// ---------------------------------------------------------------------------
// CursorState
// ---------------------------------------------------------------------------

func TestCursorModeState(t *testing.T) {
	cs := NewCursorState()

	assert.Equal(t, CursorNormal, cs.Mode())
	assert.False(t, cs.IsCaptured())

	cs.SetCursorMode(CursorCaptured)
	assert.Equal(t, CursorCaptured, cs.Mode())
	assert.True(t, cs.IsCaptured())

	cs.SetCursorMode(CursorHidden)
	assert.Equal(t, CursorHidden, cs.Mode())
	assert.False(t, cs.IsCaptured())

	cs.SetCursorMode(CursorNormal)
	assert.Equal(t, CursorNormal, cs.Mode())
	assert.False(t, cs.IsCaptured())
}
