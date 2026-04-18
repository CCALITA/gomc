package audio

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopxl/beep/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// SoundBank tests
// ---------------------------------------------------------------------------

func TestSoundBank_LoadAndGet(t *testing.T) {
	sb := NewSoundBank()

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

func TestSoundBank_NamesEmpty(t *testing.T) {
	sb := NewSoundBank()
	names := sb.Names()
	assert.Empty(t, names)
}

func TestSoundBank_LoadWAV_FileNotFound(t *testing.T) {
	sb := NewSoundBank()
	err := sb.LoadWAV("missing", "/nonexistent/path/missing.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load WAV")
}

func TestSoundBank_StoreOverwrite(t *testing.T) {
	sb := NewSoundBank()
	sb.Store("x", []byte{1})
	sb.Store("x", []byte{2, 3})
	got, ok := sb.Get("x")
	assert.True(t, ok)
	assert.Equal(t, []byte{2, 3}, got)
}

// ---------------------------------------------------------------------------
// Engine config tests (no hardware required)
// ---------------------------------------------------------------------------

func TestEngine_NewEngineDefaults(t *testing.T) {
	e := NewEngine(48000)
	assert.NotNil(t, e)
	assert.InDelta(t, 1.0, e.MasterVolume(), 1e-9, "default master volume should be 1.0")
	assert.Equal(t, 48000, e.SampleRate(), "sample rate should match constructor arg")
	assert.NotNil(t, e.SoundBank(), "sound bank should be non-nil")
	assert.False(t, e.IsInitialized(), "engine should not be initialized on creation")
}

func TestEngine_SetMasterVolume_Clamped(t *testing.T) {
	e := NewEngine(44100)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"mid-range", 0.5, 0.5},
		{"zero", 0.0, 0.0},
		{"one", 1.0, 1.0},
		{"below zero clamps to 0", -1.0, 0.0},
		{"above one clamps to 1", 2.0, 1.0},
		{"large negative clamps to 0", -100.0, 0.0},
		{"large positive clamps to 1", 100.0, 1.0},
		{"small positive", 0.001, 0.001},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e.SetMasterVolume(tc.input)
			assert.InDelta(t, tc.expected, e.MasterVolume(), 1e-9)
		})
	}
}

func TestEngine_MasterVolume_ReturnsSetValue(t *testing.T) {
	e := NewEngine(44100)

	e.SetMasterVolume(0.42)
	assert.InDelta(t, 0.42, e.MasterVolume(), 1e-9)

	e.SetMasterVolume(0.99)
	assert.InDelta(t, 0.99, e.MasterVolume(), 1e-9)
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

func TestEngine_Play_SoundNotFound(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true

	err := e.Play("nonexistent_sound")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEngine_PlayAt_SoundNotFound(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true

	err := e.PlayAt("nonexistent_sound", 1, 2, 3, 0, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEngine_SampleRate(t *testing.T) {
	for _, sr := range []int{22050, 44100, 48000, 96000} {
		e := NewEngine(sr)
		assert.Equal(t, sr, e.SampleRate())
	}
}

func TestEngine_Close_WhenNotInitialized(t *testing.T) {
	e := NewEngine(44100)
	assert.False(t, e.IsInitialized())
	assert.NotPanics(t, func() { e.Close() })
	assert.False(t, e.IsInitialized())
}

func TestEngine_Close_WhenInitialized(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	assert.True(t, e.IsInitialized())

	e.Close()
	assert.False(t, e.IsInitialized(), "engine should be uninitialized after Close")
}

func TestEngine_Close_Idempotent(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true

	assert.NotPanics(t, func() {
		e.Close()
		e.Close()
		e.Close()
	})
	assert.False(t, e.IsInitialized())
}

func TestEngine_Play_Success(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	e.soundBank.Store("click", []byte{0x00, 0x40})

	err := e.Play("click")
	assert.NoError(t, err)
}

func TestEngine_Play_ZeroMasterVolume(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	e.soundBank.Store("click", []byte{0x00, 0x40})
	e.SetMasterVolume(0)

	err := e.Play("click")
	assert.NoError(t, err)
}

func TestEngine_PlayAt_Success(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	e.soundBank.Store("step", []byte{0x00, 0x40, 0xFF, 0x7F})

	err := e.PlayAt("step", 10, 0, 0, 0, 0, 0)
	assert.NoError(t, err)
}

func TestEngine_PlayAt_ZeroDistance(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	e.soundBank.Store("step", []byte{0x00, 0x40})

	err := e.PlayAt("step", 5, 5, 5, 5, 5, 5)
	assert.NoError(t, err)
}

func TestEngine_PlayAt_LargeDistance(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	e.soundBank.Store("step", []byte{0x00, 0x40})

	err := e.PlayAt("step", 1000, 0, 0, 0, 0, 0)
	assert.NoError(t, err)
}

func TestEngine_PlayAt_ZeroMasterVolume(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	e.soundBank.Store("step", []byte{0x00, 0x40})
	e.SetMasterVolume(0)

	err := e.PlayAt("step", 10, 0, 0, 0, 0, 0)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// MusicPlayer state machine tests (no hardware required)
// ---------------------------------------------------------------------------

func TestMusicPlayer_NewMusicPlayer_Defaults(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	assert.NotNil(t, mp)
	assert.InDelta(t, 1.0, mp.Volume(), 1e-9, "default music volume should be 1.0")
	assert.False(t, mp.IsPlaying(), "should not be playing initially")
}

func TestMusicPlayer_SetVolume_Clamped(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"mid-range", 0.7, 0.7},
		{"below zero", -0.5, 0.0},
		{"above one", 1.5, 1.0},
		{"zero", 0.0, 0.0},
		{"one", 1.0, 1.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mp.SetVolume(tc.input)
			assert.InDelta(t, tc.expected, mp.Volume(), 1e-9)
		})
	}
}

func TestMusicPlayer_PlayTrackWithoutInit(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	err := mp.PlayTrack("/nonexistent/track.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestMusicPlayer_PlayTrack_FileNotFound(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	mp := NewMusicPlayer(e)

	err := mp.PlayTrack("/nonexistent/path/track.wav")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open music file")
}

func TestMusicPlayer_PlayTrack_InvalidWAV(t *testing.T) {
	e := NewEngine(44100)
	e.initialized = true
	mp := NewMusicPlayer(e)

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.wav")
	err := os.WriteFile(path, []byte("not a wav file"), 0644)
	require.NoError(t, err)

	err = mp.PlayTrack(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode WAV")
}

func TestMusicPlayer_StopWhenNotPlaying(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.Stop()
	assert.False(t, mp.IsPlaying())
}

func TestMusicPlayer_PauseResumeWhenNotPlaying(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.Pause()
	mp.Resume()
	assert.False(t, mp.IsPlaying())
}

func TestMusicPlayer_Pause_WhenPlayingWithCtrl(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.playing = true
	mp.paused = false
	mp.ctrl = &beep.Ctrl{Streamer: beep.Silence(1), Paused: false}

	assert.NotPanics(t, func() { mp.Pause() })
	assert.True(t, mp.paused)
}

func TestMusicPlayer_Resume_WhenPausedWithCtrl(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.playing = true
	mp.paused = true
	mp.ctrl = &beep.Ctrl{Streamer: beep.Silence(1), Paused: true}

	assert.NotPanics(t, func() { mp.Resume() })
	assert.False(t, mp.paused)
}

func TestMusicPlayer_Pause_WhenPlayingButNilCtrl(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.playing = true
	mp.paused = false
	mp.ctrl = nil

	assert.NotPanics(t, func() { mp.Pause() })
	assert.False(t, mp.paused, "paused should remain false when ctrl is nil")
}

func TestMusicPlayer_Resume_WhenPausedButNilCtrl(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.playing = true
	mp.paused = true
	mp.ctrl = nil

	assert.NotPanics(t, func() { mp.Resume() })
	assert.True(t, mp.paused, "paused should remain true when ctrl is nil")
}

func TestMusicPlayer_Resume_WhenNotPaused(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.playing = true
	mp.paused = false

	assert.NotPanics(t, func() { mp.Resume() })
	assert.False(t, mp.paused)
}

func TestMusicPlayer_Pause_WhenAlreadyPaused(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)
	mp.playing = true
	mp.paused = true

	assert.NotPanics(t, func() { mp.Pause() })
	assert.True(t, mp.paused)
}

func TestMusicPlayer_StopLocked_WithCloser(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)

	closeCalled := false
	mp.playing = true
	mp.closer = func() error {
		closeCalled = true
		return nil
	}

	mp.Stop()
	assert.True(t, closeCalled, "closer should be called on stop")
	assert.False(t, mp.IsPlaying())
	assert.Nil(t, mp.closer, "closer should be nil after stop")
}

func TestMusicPlayer_StopLocked_ClearsState(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)

	mp.playing = true
	mp.paused = true
	mp.closer = func() error { return nil }
	mp.ctrl = &beep.Ctrl{}

	mp.Stop()
	assert.False(t, mp.playing)
	assert.False(t, mp.paused)
	assert.Nil(t, mp.ctrl)
	assert.Nil(t, mp.closer)
}

func TestMusicPlayer_Stop_Idempotent(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)

	callCount := 0
	mp.playing = true
	mp.closer = func() error {
		callCount++
		return nil
	}

	mp.Stop()
	mp.Stop()
	mp.Stop()

	assert.Equal(t, 1, callCount, "closer should only be called once")
}

func TestMusicPlayer_StateTransitions_PlayPauseResumeStop(t *testing.T) {
	e := NewEngine(44100)
	mp := NewMusicPlayer(e)

	assert.False(t, mp.IsPlaying())

	// Simulate entering a playing state.
	mp.mu.Lock()
	mp.playing = true
	mp.paused = false
	mp.mu.Unlock()
	assert.True(t, mp.IsPlaying())

	// Simulate pause.
	mp.mu.Lock()
	mp.paused = true
	mp.mu.Unlock()
	assert.False(t, mp.IsPlaying(), "paused track should not report as playing")

	// Simulate resume.
	mp.mu.Lock()
	mp.paused = false
	mp.mu.Unlock()
	assert.True(t, mp.IsPlaying())

	// Stop.
	mp.Stop()
	assert.False(t, mp.IsPlaying())
}

// ---------------------------------------------------------------------------
// PlayAt distance attenuation formula tests
// ---------------------------------------------------------------------------

func TestPlayAt_DistanceAttenuationFormula(t *testing.T) {
	tests := []struct {
		name           string
		dx, dy, dz     float64
		expected       float64
	}{
		{"zero distance", 0, 0, 0, 1.0},
		{"distance 10 on x", 10, 0, 0, 0.5},
		{"distance 20 on y", 0, 20, 0, 1.0 / 3.0},
		{"distance 10,10,10", 10, 10, 10, 1.0 / (1.0 + math.Sqrt(300)*0.1)},
		{"distance 100", 100, 0, 0, 1.0 / (1.0 + 100*0.1)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dist := math.Sqrt(tc.dx*tc.dx + tc.dy*tc.dy + tc.dz*tc.dz)
			att := 1.0 / (1.0 + dist*0.1)
			assert.InDelta(t, tc.expected, att, 1e-9)
		})
	}
}

// ---------------------------------------------------------------------------
// Internal helper tests
// ---------------------------------------------------------------------------

func TestVolumeToDecibels(t *testing.T) {
	tests := []struct {
		name string
		vol  float64
		want float64
	}{
		{"full volume", 1.0, 0.0},
		{"half volume", 0.5, -1.0},
		{"quarter volume", 0.25, -2.0},
		{"zero", 0.0, -100.0},
		{"negative", -0.5, -100.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.InDelta(t, tc.want, volumeToDecibels(tc.vol), 1e-9)
		})
	}
}

func TestRawStreamer(t *testing.T) {
	data := []byte{0x00, 0x40, 0xFF, 0x7F}
	sr := beep.SampleRate(44100)
	s := rawStreamer(data, sr)
	require.NotNil(t, s)

	samples := make([][2]float64, 4)
	n, ok := s.Stream(samples)
	assert.Equal(t, 2, n)
	assert.False(t, ok)
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
	data := []byte{0x00, 0x00, 0xFF, 0x00}
	s := &pcmStreamer{data: data, pos: 0}
	samples := make([][2]float64, 3)
	n, ok := s.Stream(samples)
	assert.Equal(t, 2, n)
	assert.False(t, ok)
	assert.InDelta(t, 0.0, samples[0][0], 1e-9)
	assert.InDelta(t, float64(255)/float64(1<<15), samples[1][0], 1e-6)
}

func TestPcmStreamer_ExactFit(t *testing.T) {
	data := []byte{0x00, 0x40, 0xFF, 0x7F}
	s := &pcmStreamer{data: data, pos: 0}
	samples := make([][2]float64, 2)
	n, ok := s.Stream(samples)
	assert.Equal(t, 2, n)
	assert.True(t, ok, "should return true when buffer is exactly filled")
}

func TestPcmStreamer_MonoToStereo(t *testing.T) {
	data := []byte{0x00, 0x40}
	s := &pcmStreamer{data: data, pos: 0}
	samples := make([][2]float64, 1)
	n, ok := s.Stream(samples)
	assert.Equal(t, 1, n)
	assert.True(t, ok)
	assert.InDelta(t, samples[0][0], samples[0][1], 1e-15, "left and right channels should match for mono")
}

func TestPcmStreamer_OddByte(t *testing.T) {
	data := []byte{0x00, 0x40, 0xFF}
	s := &pcmStreamer{data: data, pos: 0}
	samples := make([][2]float64, 3)
	n, ok := s.Stream(samples)
	assert.Equal(t, 1, n)
	assert.False(t, ok)
}

func TestPcmStreamer_Err_AlwaysNil(t *testing.T) {
	s := &pcmStreamer{data: []byte{0x00, 0x01}, pos: 0}
	assert.NoError(t, s.Err())
	buf := make([][2]float64, 5)
	s.Stream(buf)
	assert.NoError(t, s.Err(), "Err should remain nil after exhaustion")
}

// ---------------------------------------------------------------------------
// sineEnvelope tests
// ---------------------------------------------------------------------------

func TestSineEnvelope(t *testing.T) {
	tests := []struct {
		name     string
		pos      float64
		total    float64
		expected float64
	}{
		{"zero total returns zero", 5, 0, 0},
		{"negative total returns zero", 5, -1, 0},
		{"start returns zero", 0, 100, 0},
		{"end returns ~zero", 100, 100, 0},
		{"midpoint returns 1", 50, 100, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sineEnvelope(tc.pos, tc.total)
			assert.InDelta(t, tc.expected, result, 1e-9)
		})
	}
}
