package block

import (
	"sync"

	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// JukeboxState holds the current state of a single jukebox block.
type JukeboxState struct {
	DiscID  uint16
	Playing bool
}

// JukeboxManager tracks jukebox states across the world.
type JukeboxManager struct {
	mu     sync.Mutex
	states map[mcmath.BlockPos]JukeboxState
}

// NewJukeboxManager creates a new JukeboxManager.
func NewJukeboxManager() *JukeboxManager {
	return &JukeboxManager{
		states: make(map[mcmath.BlockPos]JukeboxState),
	}
}

// InsertDisc sets the jukebox at pos to playing with the given disc ID.
// Returns false if a disc is already inserted.
func (m *JukeboxManager) InsertDisc(pos mcmath.BlockPos, discID uint16) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.states[pos]; ok && s.Playing {
		return false
	}

	m.states[pos] = JukeboxState{DiscID: discID, Playing: true}
	return true
}

// EjectDisc removes and returns the disc ID from the jukebox at pos.
// Returns 0 if no disc is inserted.
func (m *JukeboxManager) EjectDisc(pos mcmath.BlockPos) uint16 {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.states[pos]
	if !ok || !s.Playing {
		return 0
	}

	delete(m.states, pos)
	return s.DiscID
}

// IsPlaying reports whether the jukebox at pos is currently playing.
func (m *JukeboxManager) IsPlaying(pos mcmath.BlockPos) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.states[pos]
	return ok && s.Playing
}

// discTrackNames maps disc item IDs to their audio track names.
var discTrackNames = map[uint16]string{
	item.Disc13:     "music/13",
	item.DiscCat:    "music/cat",
	item.DiscBlocks: "music/blocks",
	item.DiscChirp:  "music/chirp",
	item.DiscFar:    "music/far",
}

// GetDiscTrackName returns the audio track name for the given disc item ID.
// Returns an empty string if the disc ID is not recognized.
func GetDiscTrackName(discID uint16) string {
	return discTrackNames[discID]
}

// IsDisc reports whether the given item ID is a music disc.
func IsDisc(itemID uint16) bool {
	_, ok := discTrackNames[itemID]
	return ok
}
