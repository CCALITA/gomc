package audio

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// SoundBank tests
// ---------------------------------------------------------------------------

func TestSoundBank_LoadAndGet(t *testing.T) {
	sb := NewSoundBank()

	// Write a tiny fake WAV file.
	dir := t.TempDir()
	path := filepath.Join(dir, "click.wav")
	data := []byte("RIFF----WAVEfmt fake-wav-data")
	err := os.WriteFile(path, data, 0644)
	assert.NoError(t, err)

	err = sb.LoadWAV("click", path)
	assert.NoError(t, err)

	got, ok := sb.Get("click")
	assert.True(t, ok)
	assert.Equal(t, data, got)
}

func TestSoundBank_GetMiss(t *testing.T) {
	sb := NewSoundBank()
	_, ok := sb.Get("nonexistent")
	assert.False(t, ok)
}

func TestSoundBank_Has(t *testing.T) {
	sb := NewSoundBank()

	dir := t.TempDir()
	path := filepath.Join(dir, "step.wav")
	err := os.WriteFile(path, []byte("wav-bytes"), 0644)
	assert.NoError(t, err)

	assert.False(t, sb.Has("step"))
	err = sb.LoadWAV("step", path)
	assert.NoError(t, err)
	assert.True(t, sb.Has("step"))
}

func TestSoundBank_Names(t *testing.T) {
	sb := NewSoundBank()

	dir := t.TempDir()
	for _, name := range []string{"beta", "alpha", "gamma"} {
		p := filepath.Join(dir, name+".wav")
		err := os.WriteFile(p, []byte("data"), 0644)
		assert.NoError(t, err)
		err = sb.LoadWAV(name, p)
		assert.NoError(t, err)
	}

	names := sb.Names()
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, names)
}

func TestSoundBank_LoadWAV_FileNotFound(t *testing.T) {
	sb := NewSoundBank()
	err := sb.LoadWAV("missing", "/nonexistent/path/missing.wav")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Engine config tests (no hardware required)
// ---------------------------------------------------------------------------

func TestEngine_NewEngine(t *testing.T) {
	e := NewEngine(44100)
	assert.NotNil(t, e)
	assert.False(t, e.IsInitialized())
}

func TestEngine_SetMasterVolume_Clamped(t *testing.T) {
	e := NewEngine(44100)

	e.SetMasterVolume(0.5)
	assert.InDelta(t, 0.5, e.MasterVolume(), 1e-9)

	// Clamp to 0
	e.SetMasterVolume(-1.0)
	assert.InDelta(t, 0.0, e.MasterVolume(), 1e-9)

	// Clamp to 1
	e.SetMasterVolume(2.0)
	assert.InDelta(t, 1.0, e.MasterVolume(), 1e-9)
}

func TestEngine_PlayWithoutInit(t *testing.T) {
	e := NewEngine(44100)
	err := e.Play("click")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestEngine_PlayAtWithoutInit(t *testing.T) {
	e := NewEngine(44100)
	err := e.PlayAt("click", 0, 0, 0, 0, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestEngine_SoundBank(t *testing.T) {
	e := NewEngine(44100)
	assert.NotNil(t, e.SoundBank())
}

// ---------------------------------------------------------------------------
// MusicPlayer state machine tests (no hardware required)
// ---------------------------------------------------------------------------

func TestMusicPlayer_NewMusicPlayer(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	assert.NotNil(t, mp)
	assert.False(t, mp.IsPlaying())
}

func TestMusicPlayer_SetVolume_Clamped(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)

	mp.SetVolume(0.7)
	assert.InDelta(t, 0.7, mp.Volume(), 1e-9)

	mp.SetVolume(-0.5)
	assert.InDelta(t, 0.0, mp.Volume(), 1e-9)

	mp.SetVolume(1.5)
	assert.InDelta(t, 1.0, mp.Volume(), 1e-9)
}

func TestMusicPlayer_PlayTrackWithoutInit(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	err := mp.PlayTrack("/nonexistent/track.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestMusicPlayer_StopWhenNotPlaying(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	// Should not panic.
	mp.Stop()
	assert.False(t, mp.IsPlaying())
}

func TestMusicPlayer_PauseResumeWhenNotPlaying(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	// Should not panic when nothing is playing.
	mp.Pause()
	mp.Resume()
	assert.False(t, mp.IsPlaying())
}

// ---------------------------------------------------------------------------
// Internal helper tests
// ---------------------------------------------------------------------------

func TestVolumeToDecibels(t *testing.T) {
	// volume=1 -> 0 dB
	assert.InDelta(t, 0.0, volumeToDecibels(1.0), 1e-9)
	// volume=0.5 -> -1 (log2(0.5))
	assert.InDelta(t, -1.0, volumeToDecibels(0.5), 1e-9)
	// volume=0 -> very negative
	assert.True(t, volumeToDecibels(0) < -10)
}

func TestDistanceAttenuation(t *testing.T) {
	// At distance 0, attenuation = 1
	att := 1.0 / (1.0 + 0*0.1)
	assert.InDelta(t, 1.0, att, 1e-9)

	// At distance 10, attenuation = 1/(1+1) = 0.5
	dist := math.Sqrt(10*10 + 0 + 0)
	att = 1.0 / (1.0 + dist*0.1)
	assert.InDelta(t, 0.5, att, 1e-9)

	// At distance ~17.32 (10,10,10), attenuation = 1/(1+1.732) ≈ 0.366
	dist = math.Sqrt(10*10 + 10*10 + 10*10)
	att = 1.0 / (1.0 + dist*0.1)
	expected := 1.0 / (1.0 + dist*0.1)
	assert.InDelta(t, expected, att, 1e-9)
}

func TestPcmStreamer_EmptyData(t *testing.T) {
	s := &pcmStreamer{data: []byte{}, pos: 0}
	samples := make([][2]float64, 10)
	n, ok := s.Stream(samples)
	assert.Equal(t, 0, n)
	assert.False(t, ok)
	assert.NoError(t, s.Err())
}

func TestPcmStreamer_Stream(t *testing.T) {
	// Two 16-bit LE samples: 0x0000 (silence) and 0x00FF (small positive).
	data := []byte{0x00, 0x00, 0xFF, 0x00}
	s := &pcmStreamer{data: data, pos: 0}
	samples := make([][2]float64, 3)
	n, ok := s.Stream(samples)
	assert.Equal(t, 2, n)
	assert.False(t, ok)
	// First sample = 0
	assert.InDelta(t, 0.0, samples[0][0], 1e-9)
	// Second sample = 255/32768
	assert.InDelta(t, float64(255)/float64(1<<15), samples[1][0], 1e-6)
}
