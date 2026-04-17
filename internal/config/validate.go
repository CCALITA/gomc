package config

import "fmt"

// Validate checks cfg for invalid field values. It returns the first error found.
func Validate(cfg *Config) error {
	if cfg.Window.Width <= 0 {
		return fmt.Errorf("window.width must be > 0, got %d", cfg.Window.Width)
	}
	if cfg.Window.Height <= 0 {
		return fmt.Errorf("window.height must be > 0, got %d", cfg.Window.Height)
	}

	if cfg.Render.ViewDistance < 2 || cfg.Render.ViewDistance > 32 {
		return fmt.Errorf("render.view_distance must be 2-32, got %d", cfg.Render.ViewDistance)
	}
	if cfg.Render.FOV < 30 || cfg.Render.FOV > 120 {
		return fmt.Errorf("render.fov must be 30-120, got %v", cfg.Render.FOV)
	}

	if cfg.Audio.MasterVolume < 0.0 || cfg.Audio.MasterVolume > 1.0 {
		return fmt.Errorf("audio.master_volume must be 0.0-1.0, got %v", cfg.Audio.MasterVolume)
	}
	if cfg.Audio.MusicVolume < 0.0 || cfg.Audio.MusicVolume > 1.0 {
		return fmt.Errorf("audio.music_volume must be 0.0-1.0, got %v", cfg.Audio.MusicVolume)
	}
	if cfg.Audio.SFXVolume < 0.0 || cfg.Audio.SFXVolume > 1.0 {
		return fmt.Errorf("audio.sfx_volume must be 0.0-1.0, got %v", cfg.Audio.SFXVolume)
	}

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be 1-65535, got %d", cfg.Server.Port)
	}

	if cfg.Controls.MouseSensitivity < 0.01 || cfg.Controls.MouseSensitivity > 10.0 {
		return fmt.Errorf("controls.mouse_sensitivity must be 0.01-10.0, got %v", cfg.Controls.MouseSensitivity)
	}

	return nil
}
