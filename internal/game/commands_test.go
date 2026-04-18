//go:build !ci

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandRegistry_HelpListsAll(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/help")

	names := reg.CommandNames()
	for _, name := range names {
		assert.Contains(t, result, "/"+name,
			"help output should list /%s", name)
	}
}

func TestCommandRegistry_UnknownCommandError(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/notacommand")
	assert.Contains(t, result, "Unknown command")
	assert.Contains(t, result, "notacommand")
}

func TestCommandRegistry_GameModeParseSurvival(t *testing.T) {
	tests := []struct {
		name  string
		input string
		mode  GameMode
	}{
		{"survival by name", "/gamemode survival", ModeSurvival},
		{"creative by name", "/gamemode creative", ModeCreative},
		{"spectator by name", "/gamemode spectator", ModeSpectator},
		{"survival by number", "/gamemode 0", ModeSurvival},
		{"creative by number", "/gamemode 1", ModeCreative},
		{"spectator by number", "/gamemode 2", ModeSpectator},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMode GameMode
			called := false
			cb := &CommandCallbacks{
				SetGameMode: func(mode GameMode) {
					gotMode = mode
					called = true
				},
			}
			reg := NewCommandRegistry(cb)
			result := reg.Execute(tt.input)

			assert.True(t, called, "SetGameMode should be called")
			assert.Equal(t, tt.mode, gotMode)
			assert.Contains(t, result, "Game mode set to")
		})
	}
}

func TestCommandRegistry_GameModeUnknown(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/gamemode adventure")
	assert.Contains(t, result, "Unknown game mode")
}

func TestCommandRegistry_GameModeNoArgs(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/gamemode")
	assert.Contains(t, result, "Usage:")
}

func TestCommandRegistry_TimeSetDay(t *testing.T) {
	var gotTicks int64
	cb := &CommandCallbacks{
		SetTime: func(ticks int64) { gotTicks = ticks },
	}
	reg := NewCommandRegistry(cb)
	result := reg.Execute("/time set day")

	assert.Equal(t, int64(6000), gotTicks)
	assert.Contains(t, result, "6000")
}

func TestCommandRegistry_TimeSetNumeric(t *testing.T) {
	var gotTicks int64
	cb := &CommandCallbacks{
		SetTime: func(ticks int64) { gotTicks = ticks },
	}
	reg := NewCommandRegistry(cb)
	result := reg.Execute("/time set 12345")

	assert.Equal(t, int64(12345), gotTicks)
	assert.Contains(t, result, "12345")
}

func TestCommandRegistry_TimeSetNight(t *testing.T) {
	var gotTicks int64
	cb := &CommandCallbacks{
		SetTime: func(ticks int64) { gotTicks = ticks },
	}
	reg := NewCommandRegistry(cb)
	reg.Execute("/time set night")
	assert.Equal(t, int64(18000), gotTicks)
}

func TestCommandRegistry_TimeInvalidValue(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/time set abc")
	assert.Contains(t, result, "Invalid time value")
}

func TestCommandRegistry_TimeNoArgs(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/time")
	assert.Contains(t, result, "Usage:")
}

func TestCommandRegistry_Kill(t *testing.T) {
	called := false
	cb := &CommandCallbacks{
		KillPlayer: func() { called = true },
	}
	reg := NewCommandRegistry(cb)
	result := reg.Execute("/kill")

	assert.True(t, called)
	assert.Contains(t, result, "Killed player")
}

func TestCommandRegistry_KillNilCallback(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/kill")
	assert.Contains(t, result, "Killed player")
}

func TestCommandRegistry_TeleportCoords(t *testing.T) {
	var gotX, gotY, gotZ float64
	cb := &CommandCallbacks{
		Teleport: func(x, y, z float64) {
			gotX, gotY, gotZ = x, y, z
		},
	}
	reg := NewCommandRegistry(cb)
	result := reg.Execute("/tp 100.5 64 -200.3")

	assert.InDelta(t, 100.5, gotX, 0.01)
	assert.InDelta(t, 64.0, gotY, 0.01)
	assert.InDelta(t, -200.3, gotZ, 0.01)
	assert.Contains(t, result, "Teleported to")
}

func TestCommandRegistry_TeleportInvalidCoords(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/tp abc 64 100")
	assert.Contains(t, result, "Invalid x coordinate")
}

func TestCommandRegistry_TeleportTooFewArgs(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/tp 100 64")
	assert.Contains(t, result, "Usage:")
}

func TestCommandRegistry_Give(t *testing.T) {
	var gotItem string
	var gotCount int
	cb := &CommandCallbacks{
		GiveItem: func(name string, count int) {
			gotItem = name
			gotCount = count
		},
	}
	reg := NewCommandRegistry(cb)
	result := reg.Execute("/give stone 32")

	assert.Equal(t, "stone", gotItem)
	assert.Equal(t, 32, gotCount)
	assert.Contains(t, result, "Gave 32 stone")
}

func TestCommandRegistry_GiveDefaultCount(t *testing.T) {
	var gotCount int
	cb := &CommandCallbacks{
		GiveItem: func(_ string, count int) { gotCount = count },
	}
	reg := NewCommandRegistry(cb)
	reg.Execute("/give diamond")
	assert.Equal(t, 1, gotCount)
}

func TestCommandRegistry_GiveNoArgs(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/give")
	assert.Contains(t, result, "Usage:")
}

func TestCommandRegistry_Seed(t *testing.T) {
	cb := &CommandCallbacks{
		GetSeed: func() int64 { return 42 },
	}
	reg := NewCommandRegistry(cb)
	result := reg.Execute("/seed")
	assert.Contains(t, result, "42")
}

func TestCommandRegistry_SeedNilCallback(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/seed")
	assert.Contains(t, result, "unknown")
}

func TestCommandRegistry_Weather(t *testing.T) {
	var gotState WeatherState
	cb := &CommandCallbacks{
		SetWeather: func(state WeatherState) { gotState = state },
	}
	reg := NewCommandRegistry(cb)

	reg.Execute("/weather rain")
	assert.Equal(t, Rain, gotState)

	reg.Execute("/weather clear")
	assert.Equal(t, Clear, gotState)

	reg.Execute("/weather thunder")
	assert.Equal(t, Thunder, gotState)
}

func TestCommandRegistry_WeatherUnknown(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/weather snow")
	assert.Contains(t, result, "Unknown weather")
}

func TestCommandRegistry_EmptyInput(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("")
	assert.Equal(t, "", result)
}

func TestCommandRegistry_NonCommandInput(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("hello world")
	assert.Equal(t, "", result)
}

func TestCommandRegistry_CaseInsensitive(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/HELP")
	assert.Contains(t, result, "Available commands")
}

func TestCommandRegistry_CustomCommand(t *testing.T) {
	reg := NewCommandRegistry(nil)
	reg.Register("ping", func(_ []string) string {
		return "pong"
	})
	result := reg.Execute("/ping")
	assert.Equal(t, "pong", result)
}

func TestCommandRegistry_RegisterNilHandler(t *testing.T) {
	reg := NewCommandRegistry(nil)
	reg.Register("bad", nil)
	result := reg.Execute("/bad")
	assert.Contains(t, result, "Unknown command")
}

func TestCommandRegistry_CommandNames(t *testing.T) {
	reg := NewCommandRegistry(nil)
	names := reg.CommandNames()

	assert.Contains(t, names, "help")
	assert.Contains(t, names, "gamemode")
	assert.Contains(t, names, "time")
	assert.Contains(t, names, "weather")
	assert.Contains(t, names, "kill")
	assert.Contains(t, names, "tp")
	assert.Contains(t, names, "give")
	assert.Contains(t, names, "seed")

	// Verify sorted.
	for i := 1; i < len(names); i++ {
		assert.True(t, names[i-1] <= names[i], "names should be sorted")
	}
}

func TestCommandRegistry_GiveInvalidCount(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/give stone abc")
	assert.Contains(t, result, "Invalid count")
}

func TestCommandRegistry_GiveZeroCount(t *testing.T) {
	reg := NewCommandRegistry(nil)
	result := reg.Execute("/give stone 0")
	assert.Contains(t, result, "Invalid count")
}

func TestCommandRegistry_TimeSetNoon(t *testing.T) {
	var gotTicks int64
	cb := &CommandCallbacks{
		SetTime: func(ticks int64) { gotTicks = ticks },
	}
	reg := NewCommandRegistry(cb)
	reg.Execute("/time set noon")
	assert.Equal(t, int64(12000), gotTicks)
}

func TestCommandRegistry_TimeSetMidnight(t *testing.T) {
	var gotTicks int64
	cb := &CommandCallbacks{
		SetTime: func(ticks int64) { gotTicks = ticks },
	}
	reg := NewCommandRegistry(cb)
	reg.Execute("/time set midnight")
	assert.Equal(t, int64(0), gotTicks)
}
