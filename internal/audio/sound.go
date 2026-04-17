package audio

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// SoundBank stores raw audio bytes keyed by name.
type SoundBank struct {
	mu     sync.RWMutex
	sounds map[string][]byte
}

// NewSoundBank creates an empty SoundBank.
func NewSoundBank() *SoundBank {
	return &SoundBank{
		sounds: make(map[string][]byte),
	}
}

// LoadWAV loads a WAV file from disk and stores its raw bytes under the given name.
func (sb *SoundBank) LoadWAV(name, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("audio: failed to load WAV %q from %q: %w", name, path, err)
	}

	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.sounds[name] = data
	return nil
}

// Get returns the raw audio bytes for the named sound and whether it exists.
func (sb *SoundBank) Get(name string) ([]byte, bool) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	data, ok := sb.sounds[name]
	return data, ok
}

// Has reports whether the named sound exists in the bank.
func (sb *SoundBank) Has(name string) bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	_, ok := sb.sounds[name]
	return ok
}

// Names returns a sorted list of all sound names in the bank.
func (sb *SoundBank) Names() []string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	names := make([]string, 0, len(sb.sounds))
	for name := range sb.sounds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
