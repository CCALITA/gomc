package audio

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Procedural generation tests
// ---------------------------------------------------------------------------

func TestGenerateNoiseBurst_ValidWAV(t *testing.T) {
	data := GenerateNoiseBurst(0.1, 44100, 10.0)
	assertValidWAV(t, data, 44100)
	assertNonZeroSamples(t, data)
}

func TestGenerateSineBeep_ValidWAV(t *testing.T) {
	data := GenerateSineBeep(0.1, 44100, 440.0)
	assertValidWAV(t, data, 44100)
	assertNonZeroSamples(t, data)
}

func TestGenerateChirp_ValidWAV(t *testing.T) {
	data := GenerateChirp(0.1, 44100, 200.0, 800.0)
	assertValidWAV(t, data, 44100)
	assertNonZeroSamples(t, data)
}

func TestGenerateNoiseBurst_DifferentSampleRates(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate int
	}{
		{"22050Hz", 22050},
		{"44100Hz", 44100},
		{"48000Hz", 48000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := GenerateNoiseBurst(0.05, tc.sampleRate, 15.0)
			assertValidWAV(t, data, tc.sampleRate)
		})
	}
}

func TestGenerateSineBeep_ZeroDuration(t *testing.T) {
	data := GenerateSineBeep(0, 44100, 440.0)
	assertValidWAV(t, data, 44100)
	// Zero duration should produce a valid WAV with zero data bytes.
	dataSize := binary.LittleEndian.Uint32(data[40:44])
	assert.Equal(t, uint32(0), dataSize)
}

func TestGenerateChirp_SameStartEnd(t *testing.T) {
	// When start == end the chirp degenerates to a pure tone — still valid.
	data := GenerateChirp(0.05, 44100, 440.0, 440.0)
	assertValidWAV(t, data, 44100)
	assertNonZeroSamples(t, data)
}

// ---------------------------------------------------------------------------
// SoundManager tests
// ---------------------------------------------------------------------------

func TestSoundManager_Init_LoadsAllSounds(t *testing.T) {
	engine := NewEngine(44100)
	sm := NewSoundManager(engine)
	sm.Init()

	bank := engine.SoundBank()
	for _, event := range AllSoundEvents() {
		assert.True(t, bank.Has(event), "sound %q should be loaded after Init", event)
	}
}

func TestSoundManager_Init_ProducesNonEmptyData(t *testing.T) {
	engine := NewEngine(44100)
	sm := NewSoundManager(engine)
	sm.Init()

	bank := engine.SoundBank()
	for _, event := range AllSoundEvents() {
		data, ok := bank.Get(event)
		require.True(t, ok)
		assert.Greater(t, len(data), 44, "sound %q should have PCM data beyond the WAV header", event)
	}
}

func TestSoundManager_PlaySound_UnknownDoesNotPanic(t *testing.T) {
	engine := NewEngine(44100)
	sm := NewSoundManager(engine)
	sm.Init()

	// Engine is not initialized (no speaker), so Play returns an error, but must not panic.
	assert.NotPanics(t, func() {
		_ = sm.PlaySound("totally_unknown_sound")
	})
}

func TestSoundManager_PlaySoundAt_UnknownDoesNotPanic(t *testing.T) {
	engine := NewEngine(44100)
	sm := NewSoundManager(engine)
	sm.Init()

	assert.NotPanics(t, func() {
		_ = sm.PlaySoundAt("totally_unknown_sound", 0, 0, 0, 0, 0, 0)
	})
}

func TestSoundManager_PlaySound_ReturnsErrorWithoutInit(t *testing.T) {
	engine := NewEngine(44100)
	sm := NewSoundManager(engine)
	// Don't call sm.Init() — no sounds loaded; engine also not initialized.
	err := sm.PlaySound(SoundBlockBreak)
	assert.Error(t, err)
}

func TestAllSoundEvents_ReturnsExpectedCount(t *testing.T) {
	events := AllSoundEvents()
	assert.Len(t, events, 13)
}

func TestAllSoundEvents_ReturnsNewSlice(t *testing.T) {
	a := AllSoundEvents()
	b := AllSoundEvents()
	a[0] = "mutated"
	assert.NotEqual(t, a[0], b[0], "AllSoundEvents should return independent copies")
}

// ---------------------------------------------------------------------------
// SoundBank.Store tests
// ---------------------------------------------------------------------------

func TestSoundBank_Store(t *testing.T) {
	sb := NewSoundBank()
	sb.Store("test", []byte{1, 2, 3})
	got, ok := sb.Get("test")
	assert.True(t, ok)
	assert.Equal(t, []byte{1, 2, 3}, got)
}

// ---------------------------------------------------------------------------
// WAV validation helpers
// ---------------------------------------------------------------------------

func assertValidWAV(t *testing.T, data []byte, expectedSampleRate int) {
	t.Helper()
	require.GreaterOrEqual(t, len(data), 44, "WAV data must be at least 44 bytes (header)")

	// RIFF header
	assert.Equal(t, "RIFF", string(data[0:4]))
	assert.Equal(t, "WAVE", string(data[8:12]))

	// fmt sub-chunk
	assert.Equal(t, "fmt ", string(data[12:16]))
	fmtSize := binary.LittleEndian.Uint32(data[16:20])
	assert.Equal(t, uint32(16), fmtSize)
	audioFormat := binary.LittleEndian.Uint16(data[20:22])
	assert.Equal(t, uint16(1), audioFormat, "should be PCM")
	channels := binary.LittleEndian.Uint16(data[22:24])
	assert.Equal(t, uint16(1), channels, "should be mono")
	sampleRate := binary.LittleEndian.Uint32(data[24:28])
	assert.Equal(t, uint32(expectedSampleRate), sampleRate)
	bitsPerSample := binary.LittleEndian.Uint16(data[34:36])
	assert.Equal(t, uint16(16), bitsPerSample)

	// data sub-chunk
	assert.Equal(t, "data", string(data[36:40]))

	// RIFF size consistency check.
	riffSize := binary.LittleEndian.Uint32(data[4:8])
	assert.Equal(t, uint32(len(data)-8), riffSize)
}

func assertNonZeroSamples(t *testing.T, data []byte) {
	t.Helper()
	require.Greater(t, len(data), 44)

	hasNonZero := false
	for i := 44; i+1 < len(data); i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		if sample != 0 {
			hasNonZero = true
			break
		}
	}
	assert.True(t, hasNonZero, "WAV should contain at least one non-zero sample")
}
