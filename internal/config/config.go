package config

// WindowConfig holds display and window settings.
type WindowConfig struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Fullscreen bool   `json:"fullscreen"`
	VSync      bool   `json:"vsync"`
	Title      string `json:"title"`
}

// RenderConfig holds rendering settings.
type RenderConfig struct {
	ViewDistance int     `json:"view_distance"`
	FOV         float32 `json:"fov"`
	MaxFPS      int     `json:"max_fps"`
}

// AudioConfig holds audio volume settings.
type AudioConfig struct {
	MasterVolume float64 `json:"master_volume"`
	MusicVolume  float64 `json:"music_volume"`
	SFXVolume    float64 `json:"sfx_volume"`
}

// ControlsConfig holds input settings.
type ControlsConfig struct {
	MouseSensitivity float64 `json:"mouse_sensitivity"`
	InvertY          bool    `json:"invert_y"`
}

// ServerConfig holds network settings.
type ServerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// DebugConfig holds debug/development settings.
type DebugConfig struct {
	ShowFPS          bool `json:"show_fps"`
	ShowChunkBorders bool `json:"show_chunk_borders"`
	Wireframe        bool `json:"wireframe"`
}

// Config is the top-level application configuration.
type Config struct {
	Window   WindowConfig   `json:"window"`
	Render   RenderConfig   `json:"render"`
	Audio    AudioConfig    `json:"audio"`
	Controls ControlsConfig `json:"controls"`
	Server   ServerConfig   `json:"server"`
	Debug    DebugConfig    `json:"debug"`
}

// Default returns a Config populated with all default values.
func Default() *Config {
	return &Config{
		Window: WindowConfig{
			Width:      1280,
			Height:     720,
			Fullscreen: false,
			VSync:      true,
			Title:      "GoMC",
		},
		Render: RenderConfig{
			ViewDistance: 8,
			FOV:         70.0,
			MaxFPS:      0,
		},
		Audio: AudioConfig{
			MasterVolume: 0.8,
			MusicVolume:  0.5,
			SFXVolume:    1.0,
		},
		Controls: ControlsConfig{
			MouseSensitivity: 0.15,
			InvertY:          false,
		},
		Server: ServerConfig{
			Address: "localhost",
			Port:    25565,
		},
		Debug: DebugConfig{
			ShowFPS:          false,
			ShowChunkBorders: false,
			Wireframe:        false,
		},
	}
}
