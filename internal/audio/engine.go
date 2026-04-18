package audio

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
)

// Engine manages audio playback and master volume.
type Engine struct {
	mu           sync.RWMutex
	sampleRate   beep.SampleRate
	masterVolume float64
	initialized  bool
	soundBank    *SoundBank
}

// NewEngine creates a new audio engine with the given sample rate.
func NewEngine(sampleRate int) *Engine {
	return &Engine{
		sampleRate:   beep.SampleRate(sampleRate),
		masterVolume: 1.0,
		soundBank:    NewSoundBank(),
	}
}

// Init initializes the speaker for audio output.
func (e *Engine) Init() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		return nil
	}

	err := speaker.Init(e.sampleRate, e.sampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("audio: failed to initialize speaker: %w", err)
	}

	e.initialized = true
	return nil
}

// Close shuts down the audio engine.
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		speaker.Close()
		e.initialized = false
	}
}

// SetMasterVolume sets the master volume. The value is clamped to [0, 1].
func (e *Engine) SetMasterVolume(v float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	e.masterVolume = v
}

// MasterVolume returns the current master volume.
func (e *Engine) MasterVolume() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.masterVolume
}

// SoundBank returns the engine's sound bank.
func (e *Engine) SoundBank() *SoundBank {
	return e.soundBank
}

// SampleRate returns the engine's sample rate as an integer.
func (e *Engine) SampleRate() int {
	return int(e.sampleRate)
}

// IsInitialized reports whether the audio engine has been initialized.
func (e *Engine) IsInitialized() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.initialized
}

// Play plays the named sound at full volume.
func (e *Engine) Play(name string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return fmt.Errorf("audio: engine not initialized")
	}

	data, ok := e.soundBank.Get(name)
	if !ok {
		return fmt.Errorf("audio: sound %q not found", name)
	}

	streamer := rawStreamer(data, e.sampleRate)
	vol := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   volumeToDecibels(e.masterVolume),
		Silent:   e.masterVolume == 0,
	}

	speaker.Play(vol)
	return nil
}

// PlayAt plays the named sound with 3D positional attenuation.
// Volume is attenuated by distance: volume = 1 / (1 + dist * 0.1).
func (e *Engine) PlayAt(name string, x, y, z, listenerX, listenerY, listenerZ float32) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return fmt.Errorf("audio: engine not initialized")
	}

	data, ok := e.soundBank.Get(name)
	if !ok {
		return fmt.Errorf("audio: sound %q not found", name)
	}

	dx := float64(x - listenerX)
	dy := float64(y - listenerY)
	dz := float64(z - listenerZ)
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	attenuation := 1.0 / (1.0 + dist*0.1)

	finalVolume := e.masterVolume * attenuation

	streamer := rawStreamer(data, e.sampleRate)
	vol := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   volumeToDecibels(finalVolume),
		Silent:   finalVolume == 0,
	}

	speaker.Play(vol)
	return nil
}

// volumeToDecibels converts a linear volume [0,1] to a decibel value for beep's Volume effect.
func volumeToDecibels(v float64) float64 {
	if v <= 0 {
		return -100
	}
	return math.Log2(v)
}

// rawStreamer wraps raw PCM bytes as a beep.Streamer.
func rawStreamer(data []byte, sr beep.SampleRate) beep.Streamer {
	return &pcmStreamer{
		data:       data,
		pos:        0,
		sampleRate: sr,
	}
}

// pcmStreamer streams raw 16-bit signed little-endian mono PCM data.
type pcmStreamer struct {
	data       []byte
	pos        int
	sampleRate beep.SampleRate
}

func (s *pcmStreamer) Stream(samples [][2]float64) (int, bool) {
	bytesPerSample := 2 // 16-bit mono
	n := 0
	for i := range samples {
		if s.pos+bytesPerSample > len(s.data) {
			for j := i; j < len(samples); j++ {
				samples[j] = [2]float64{}
			}
			return n, false
		}
		lo := s.data[s.pos]
		hi := s.data[s.pos+1]
		sample := float64(int16(uint16(lo)|uint16(hi)<<8)) / (1 << 15)
		samples[i] = [2]float64{sample, sample}
		s.pos += bytesPerSample
		n++
	}
	return n, true
}

func (s *pcmStreamer) Err() error {
	return nil
}
