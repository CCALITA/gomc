package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// modeNames maps each GameMode to its display name for command output.
var modeNames = map[GameMode]string{
	ModeSurvival:  "Survival",
	ModeCreative:  "Creative",
	ModeSpectator: "Spectator",
}

// weatherNames maps each WeatherState to its display name for command output.
var weatherNames = map[WeatherState]string{
	Clear:   "Clear",
	Rain:    "Rain",
	Thunder: "Thunder",
}

// CommandHandler processes a slash-command's arguments and returns a
// response message shown in the chat.
type CommandHandler func(args []string) string

// CommandCallbacks holds function fields that commands call to produce
// side effects on the game state. Each field is optional; if nil, the
// corresponding command prints an informational message instead.
type CommandCallbacks struct {
	SetGameMode func(mode GameMode)
	SetTime     func(ticks int64)
	SetWeather  func(state WeatherState)
	KillPlayer  func()
	Teleport    func(x, y, z float64)
	GiveItem    func(itemName string, count int)
	GetSeed     func() int64
}

// CommandRegistry stores named slash-command handlers and dispatches
// user input to the correct handler.
type CommandRegistry struct {
	handlers  map[string]CommandHandler
	callbacks *CommandCallbacks
}

// NewCommandRegistry returns a CommandRegistry pre-loaded with the
// default Minecraft-style commands.
func NewCommandRegistry(cb *CommandCallbacks) *CommandRegistry {
	if cb == nil {
		cb = &CommandCallbacks{}
	}
	r := &CommandRegistry{
		handlers:  make(map[string]CommandHandler),
		callbacks: cb,
	}
	r.registerDefaults()
	return r
}

// Register adds (or replaces) a named command handler.
func (r *CommandRegistry) Register(name string, handler CommandHandler) {
	if handler == nil {
		return
	}
	r.handlers[strings.ToLower(name)] = handler
}

// Execute parses a raw chat input string (expected to start with "/")
// and dispatches it to the matching handler. Returns the response text.
func (r *CommandRegistry) Execute(input string) string {
	input = strings.TrimSpace(input)
	if input == "" || input[0] != '/' {
		return ""
	}

	parts := strings.Fields(input)
	name := strings.ToLower(parts[0][1:]) // strip leading "/"
	args := parts[1:]

	handler, ok := r.handlers[name]
	if !ok {
		return fmt.Sprintf("Unknown command: /%s. Type /help for a list.", name)
	}
	return handler(args)
}

// CommandNames returns a sorted list of all registered command names.
func (r *CommandRegistry) CommandNames() []string {
	names := make([]string, 0, len(r.handlers))
	for n := range r.handlers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// registerDefaults wires up the standard set of slash commands.
func (r *CommandRegistry) registerDefaults() {
	r.Register("help", r.cmdHelp)
	r.Register("gamemode", r.cmdGameMode)
	r.Register("time", r.cmdTime)
	r.Register("weather", r.cmdWeather)
	r.Register("kill", r.cmdKill)
	r.Register("tp", r.cmdTeleport)
	r.Register("give", r.cmdGive)
	r.Register("seed", r.cmdSeed)
}

func (r *CommandRegistry) cmdHelp(_ []string) string {
	names := r.CommandNames()
	return "Available commands: /" + strings.Join(names, ", /")
}

func (r *CommandRegistry) cmdGameMode(args []string) string {
	if len(args) < 1 {
		return "Usage: /gamemode <survival|creative|spectator>"
	}

	var mode GameMode
	switch strings.ToLower(args[0]) {
	case "survival", "0":
		mode = ModeSurvival
	case "creative", "1":
		mode = ModeCreative
	case "spectator", "2":
		mode = ModeSpectator
	default:
		return fmt.Sprintf("Unknown game mode: %s", args[0])
	}

	if r.callbacks.SetGameMode != nil {
		r.callbacks.SetGameMode(mode)
	}

	return fmt.Sprintf("Game mode set to %s", modeNames[mode])
}

func (r *CommandRegistry) cmdTime(args []string) string {
	if len(args) < 2 {
		return "Usage: /time set <value|day|night>"
	}

	if strings.ToLower(args[0]) != "set" {
		return "Usage: /time set <value|day|night>"
	}

	var ticks int64
	switch strings.ToLower(args[1]) {
	case "day":
		ticks = dayStart
	case "night":
		ticks = dayEnd
	case "noon":
		ticks = DayLength / 2
	case "midnight":
		ticks = 0
	default:
		v, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Sprintf("Invalid time value: %s", args[1])
		}
		ticks = v
	}

	if r.callbacks.SetTime != nil {
		r.callbacks.SetTime(ticks)
	}
	return fmt.Sprintf("Time set to %d", ticks)
}

func (r *CommandRegistry) cmdWeather(args []string) string {
	if len(args) < 1 {
		return "Usage: /weather <clear|rain|thunder>"
	}

	var state WeatherState
	switch strings.ToLower(args[0]) {
	case "clear":
		state = Clear
	case "rain":
		state = Rain
	case "thunder":
		state = Thunder
	default:
		return fmt.Sprintf("Unknown weather: %s", args[0])
	}

	if r.callbacks.SetWeather != nil {
		r.callbacks.SetWeather(state)
	}

	return fmt.Sprintf("Weather set to %s", weatherNames[state])
}

func (r *CommandRegistry) cmdKill(_ []string) string {
	if r.callbacks.KillPlayer != nil {
		r.callbacks.KillPlayer()
	}
	return "Killed player"
}

func (r *CommandRegistry) cmdTeleport(args []string) string {
	if len(args) < 3 {
		return "Usage: /tp <x> <y> <z>"
	}

	x, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Sprintf("Invalid x coordinate: %s", args[0])
	}
	y, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Sprintf("Invalid y coordinate: %s", args[1])
	}
	z, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Sprintf("Invalid z coordinate: %s", args[2])
	}

	if r.callbacks.Teleport != nil {
		r.callbacks.Teleport(x, y, z)
	}
	return fmt.Sprintf("Teleported to %.1f, %.1f, %.1f", x, y, z)
}

func (r *CommandRegistry) cmdGive(args []string) string {
	if len(args) < 1 {
		return "Usage: /give <item> [count]"
	}

	itemName := args[0]
	count := 1
	if len(args) >= 2 {
		c, err := strconv.Atoi(args[1])
		if err != nil || c < 1 {
			return fmt.Sprintf("Invalid count: %s", args[1])
		}
		count = c
	}

	if r.callbacks.GiveItem != nil {
		r.callbacks.GiveItem(itemName, count)
	}
	return fmt.Sprintf("Gave %d %s", count, itemName)
}

func (r *CommandRegistry) cmdSeed(_ []string) string {
	if r.callbacks.GetSeed != nil {
		seed := r.callbacks.GetSeed()
		return fmt.Sprintf("Seed: %d", seed)
	}
	return "Seed: unknown"
}
