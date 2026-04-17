package audio

import (
	"fmt"
	"os"
	"sync"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

// MusicPlayer handles background music playback.
type MusicPlayer struct {
	mu      sync.RWMutex
	engine  *Engine
	volume  float64
	playing bool
	paused  bool
	ctrl    *beep.Ctrl
	closer  func() error
}

// NewMusicPlayer creates a new MusicPlayer tied to the given Engine.
func NewMusicPlayer(engine *Engine) *MusicPlayer {
	return &MusicPlayer{
		engine: engine,
		volume: 1.0,
	}
}

// PlayTrack starts playing a WAV music file from the given path.
func (mp *MusicPlayer) PlayTrack(path string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !mp.engine.IsInitialized() {
		return fmt.Errorf("audio: engine not initialized")
	}

	// Stop any currently playing track.
	mp.stopLocked()

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("audio: failed to open music file %q: %w", path, err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		f.Close()
		return fmt.Errorf("audio: failed to decode WAV %q: %w", path, err)
	}

	// Resample if the track sample rate differs from the engine sample rate.
	var src beep.Streamer = streamer
	if format.SampleRate != mp.engine.sampleRate {
		src = beep.Resample(4, format.SampleRate, mp.engine.sampleRate, streamer)
	}

	ctrl := &beep.Ctrl{Streamer: src, Paused: false}
	vol := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   volumeToDecibels(mp.volume),
		Silent:   mp.volume == 0,
	}

	mp.ctrl = ctrl
	mp.closer = func() error {
		return streamer.Close()
	}
	mp.playing = true
	mp.paused = false

	speaker.Play(beep.Seq(vol, beep.Callback(func() {
		mp.mu.Lock()
		defer mp.mu.Unlock()
		mp.playing = false
		mp.paused = false
	})))

	return nil
}

// Stop stops the currently playing track.
func (mp *MusicPlayer) Stop() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.stopLocked()
}

func (mp *MusicPlayer) stopLocked() {
	if mp.ctrl != nil {
		speaker.Lock()
		mp.ctrl.Paused = true
		speaker.Unlock()
	}
	if mp.closer != nil {
		mp.closer()
		mp.closer = nil
	}
	mp.ctrl = nil
	mp.playing = false
	mp.paused = false
}

// SetVolume sets the music volume. The value is clamped to [0, 1].
func (mp *MusicPlayer) SetVolume(v float64) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	mp.volume = v
}

// Volume returns the current music volume.
func (mp *MusicPlayer) Volume() float64 {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.volume
}

// IsPlaying reports whether a track is currently playing (and not paused).
func (mp *MusicPlayer) IsPlaying() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.playing && !mp.paused
}

// Pause pauses the currently playing track.
func (mp *MusicPlayer) Pause() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if mp.playing && !mp.paused && mp.ctrl != nil {
		speaker.Lock()
		mp.ctrl.Paused = true
		speaker.Unlock()
		mp.paused = true
	}
}

// Resume resumes a paused track.
func (mp *MusicPlayer) Resume() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if mp.playing && mp.paused && mp.ctrl != nil {
		speaker.Lock()
		mp.ctrl.Paused = false
		speaker.Unlock()
		mp.paused = false
	}
}
