package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Window defaults
	assert.Equal(t, 1280, cfg.Window.Width)
	assert.Equal(t, 720, cfg.Window.Height)
	assert.False(t, cfg.Window.Fullscreen)
	assert.True(t, cfg.Window.VSync)
	assert.Equal(t, "GoMC", cfg.Window.Title)

	// Render defaults
	assert.Equal(t, 8, cfg.Render.ViewDistance)
	assert.InDelta(t, float32(70.0), cfg.Render.FOV, 0.001)
	assert.Equal(t, 0, cfg.Render.MaxFPS)

	// Audio defaults
	assert.InDelta(t, 0.8, cfg.Audio.MasterVolume, 0.001)
	assert.InDelta(t, 0.5, cfg.Audio.MusicVolume, 0.001)
	assert.InDelta(t, 1.0, cfg.Audio.SFXVolume, 0.001)

	// Controls defaults
	assert.InDelta(t, 0.15, cfg.Controls.MouseSensitivity, 0.001)
	assert.False(t, cfg.Controls.InvertY)

	// Server defaults
	assert.Equal(t, "localhost", cfg.Server.Address)
	assert.Equal(t, 25565, cfg.Server.Port)

	// Debug defaults
	assert.False(t, cfg.Debug.ShowFPS)
	assert.False(t, cfg.Debug.ShowChunkBorders)
	assert.False(t, cfg.Debug.Wireframe)
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/tmp/gomc_test_nonexistent_config.json")
	assert.NoError(t, err)
	assert.Equal(t, Default(), cfg)
}

func TestSaveThenLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	original := Default()
	original.Window.Width = 1920
	original.Window.Height = 1080
	original.Audio.MasterVolume = 0.5

	err := Save(path, original)
	assert.NoError(t, err)

	loaded, err := Load(path)
	assert.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestPartialJSONMerge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.json")

	// Write partial JSON that only sets a few fields.
	partial := []byte(`{"window":{"width":1920},"audio":{"sfx_volume":0.3}}`)
	err := os.WriteFile(path, partial, 0o644)
	assert.NoError(t, err)

	cfg, err := Load(path)
	assert.NoError(t, err)

	// Overridden fields
	assert.Equal(t, 1920, cfg.Window.Width)
	assert.InDelta(t, 0.3, cfg.Audio.SFXVolume, 0.001)

	// Fields not in partial JSON keep defaults
	assert.Equal(t, 720, cfg.Window.Height)
	assert.True(t, cfg.Window.VSync)
	assert.Equal(t, "GoMC", cfg.Window.Title)
	assert.Equal(t, 8, cfg.Render.ViewDistance)
	assert.InDelta(t, 0.8, cfg.Audio.MasterVolume, 0.001)
	assert.Equal(t, "localhost", cfg.Server.Address)
	assert.Equal(t, 25565, cfg.Server.Port)
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	err := os.WriteFile(path, []byte(`{not valid json}`), 0o644)
	assert.NoError(t, err)

	_, err = Load(path)
	assert.Error(t, err)
}

func TestSaveCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "config.json")

	err := Save(path, Default())
	assert.NoError(t, err)

	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestValidatePassesGoodConfig(t *testing.T) {
	err := Validate(Default())
	assert.NoError(t, err)
}

func TestValidateCatchesBadValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(cfg *Config)
	}{
		{"zero width", func(cfg *Config) { cfg.Window.Width = 0 }},
		{"negative height", func(cfg *Config) { cfg.Window.Height = -1 }},
		{"view distance too low", func(cfg *Config) { cfg.Render.ViewDistance = 1 }},
		{"view distance too high", func(cfg *Config) { cfg.Render.ViewDistance = 33 }},
		{"fov too low", func(cfg *Config) { cfg.Render.FOV = 29 }},
		{"fov too high", func(cfg *Config) { cfg.Render.FOV = 121 }},
		{"master volume negative", func(cfg *Config) { cfg.Audio.MasterVolume = -0.1 }},
		{"master volume too high", func(cfg *Config) { cfg.Audio.MasterVolume = 1.1 }},
		{"music volume negative", func(cfg *Config) { cfg.Audio.MusicVolume = -0.1 }},
		{"music volume too high", func(cfg *Config) { cfg.Audio.MusicVolume = 1.1 }},
		{"sfx volume negative", func(cfg *Config) { cfg.Audio.SFXVolume = -0.1 }},
		{"sfx volume too high", func(cfg *Config) { cfg.Audio.SFXVolume = 1.1 }},
		{"port zero", func(cfg *Config) { cfg.Server.Port = 0 }},
		{"port too high", func(cfg *Config) { cfg.Server.Port = 65536 }},
		{"sensitivity too low", func(cfg *Config) { cfg.Controls.MouseSensitivity = 0.001 }},
		{"sensitivity too high", func(cfg *Config) { cfg.Controls.MouseSensitivity = 11.0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.mutate(cfg)
			err := Validate(cfg)
			assert.Error(t, err)
		})
	}
}

func TestValidateBoundaryValues(t *testing.T) {
	// Boundary values that should pass
	cfg := Default()
	cfg.Render.ViewDistance = 2
	cfg.Render.FOV = 30
	cfg.Audio.MasterVolume = 0.0
	cfg.Audio.MusicVolume = 1.0
	cfg.Audio.SFXVolume = 0.0
	cfg.Server.Port = 1
	cfg.Controls.MouseSensitivity = 0.01
	assert.NoError(t, Validate(cfg))

	cfg2 := Default()
	cfg2.Render.ViewDistance = 32
	cfg2.Render.FOV = 120
	cfg2.Server.Port = 65535
	cfg2.Controls.MouseSensitivity = 10.0
	assert.NoError(t, Validate(cfg2))
}
