package world

import (
	"sync"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// maxSignLineLength is the maximum number of characters allowed per sign line.
const maxSignLineLength = 15

// SignData holds the four lines of text displayed on a sign.
type SignData struct {
	Lines [4]string `json:"lines"`
}

// SignManager provides thread-safe storage for sign text data indexed by
// block position.
type SignManager struct {
	mu    sync.RWMutex
	signs map[mcmath.BlockPos]SignData
}

// NewSignManager creates an empty SignManager.
func NewSignManager() *SignManager {
	return &SignManager{
		signs: make(map[mcmath.BlockPos]SignData),
	}
}

// truncateLine returns s truncated to maxSignLineLength characters.
func truncateLine(s string) string {
	r := []rune(s)
	if len(r) > maxSignLineLength {
		return string(r[:maxSignLineLength])
	}
	return s
}

// SetSignText stores sign text at the given position. Each line is
// truncated to maxSignLineLength characters.
func (m *SignManager) SetSignText(pos mcmath.BlockPos, lines [4]string) {
	var truncated [4]string
	for i, line := range lines {
		truncated[i] = truncateLine(line)
	}
	m.mu.Lock()
	m.signs[pos] = SignData{Lines: truncated}
	m.mu.Unlock()
}

// GetSignText returns the sign data at pos and whether a sign exists there.
func (m *SignManager) GetSignText(pos mcmath.BlockPos) (SignData, bool) {
	m.mu.RLock()
	data, ok := m.signs[pos]
	m.mu.RUnlock()
	return data, ok
}

// RemoveSign deletes the sign data at the given position.
func (m *SignManager) RemoveSign(pos mcmath.BlockPos) {
	m.mu.Lock()
	delete(m.signs, pos)
	m.mu.Unlock()
}

// AllSigns returns a snapshot of all sign positions and their data.
func (m *SignManager) AllSigns() map[mcmath.BlockPos]SignData {
	m.mu.RLock()
	snapshot := make(map[mcmath.BlockPos]SignData, len(m.signs))
	for k, v := range m.signs {
		snapshot[k] = v
	}
	m.mu.RUnlock()
	return snapshot
}

// LoadSigns replaces the current sign data with the provided map.
func (m *SignManager) LoadSigns(data map[mcmath.BlockPos]SignData) {
	m.mu.Lock()
	m.signs = make(map[mcmath.BlockPos]SignData, len(data))
	for k, v := range data {
		m.signs[k] = v
	}
	m.mu.Unlock()
}
