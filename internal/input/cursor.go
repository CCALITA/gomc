package input

import "sync"

// CursorMode describes how the cursor behaves.
type CursorMode int

const (
	// CursorNormal is the default desktop cursor mode.
	CursorNormal CursorMode = iota
	// CursorCaptured locks and hides the cursor (first-person camera control).
	CursorCaptured
	// CursorHidden hides the cursor but does not lock it.
	CursorHidden
)

// CursorState tracks the logical cursor mode. The actual GLFW
// glfwSetInputMode call is made externally; this struct is the
// source of truth for the desired mode.
type CursorState struct {
	mu   sync.RWMutex
	mode CursorMode
}

// NewCursorState returns a CursorState initialised to CursorNormal.
func NewCursorState() *CursorState {
	return &CursorState{
		mode: CursorNormal,
	}
}

// SetCursorMode changes the logical cursor mode.
func (cs *CursorState) SetCursorMode(mode CursorMode) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.mode = mode
}

// Mode returns the current cursor mode.
func (cs *CursorState) Mode() CursorMode {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return cs.mode
}

// IsCaptured returns true when the cursor is in captured (locked) mode.
func (cs *CursorState) IsCaptured() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return cs.mode == CursorCaptured
}
