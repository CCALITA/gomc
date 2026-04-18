package audio

// Sound event constants identify each procedural sound effect.
const (
	SoundBlockBreak   = "block_break"
	SoundBlockPlace   = "block_place"
	SoundFootstepStone = "footstep_stone"
	SoundFootstepDirt  = "footstep_dirt"
	SoundFootstepWood  = "footstep_wood"
	SoundFootstepGrass = "footstep_grass"
	SoundFootstepSand  = "footstep_sand"
	SoundHurt         = "hurt"
	SoundDeath        = "death"
	SoundEat          = "eat"
	SoundLevelUp      = "level_up"
	SoundExplosion    = "explosion"
	SoundSplash       = "splash"
)

// allSoundEvents is an ordered list of every sound event for iteration.
var allSoundEvents = []string{
	SoundBlockBreak,
	SoundBlockPlace,
	SoundFootstepStone,
	SoundFootstepDirt,
	SoundFootstepWood,
	SoundFootstepGrass,
	SoundFootstepSand,
	SoundHurt,
	SoundDeath,
	SoundEat,
	SoundLevelUp,
	SoundExplosion,
	SoundSplash,
}

// SoundManager wraps an Engine and its SoundBank, providing a convenient API
// for playing named procedural sound effects.
//
// Integration points (do not modify game code — this documents where sounds
// should be wired in):
//   - Block break / place: internal/player/interaction.go — after BreakBlock / PlaceBlock calls
//   - Footsteps: internal/player/controller.go — in the movement update loop
//   - Hurt / Death: internal/entity/systems.go — in DamageSystem when health changes
//   - Eat: internal/entity/hunger.go — in Eat() when food is consumed
//   - Level-up: wherever XP thresholds are crossed
//   - Explosion / Splash: in the relevant entity or world effect handlers
type SoundManager struct {
	engine *Engine
}

// NewSoundManager creates a SoundManager backed by the given Engine.
func NewSoundManager(engine *Engine) *SoundManager {
	return &SoundManager{engine: engine}
}

// Init generates all procedural sounds and loads them into the Engine's SoundBank.
func (sm *SoundManager) Init() {
	sampleRate := sm.engine.SampleRate()

	bank := sm.engine.SoundBank()

	// Block break / place — short noise bursts at different decay rates.
	bank.Store(SoundBlockBreak, GenerateNoiseBurst(0.15, sampleRate, 12.0))
	bank.Store(SoundBlockPlace, GenerateNoiseBurst(0.10, sampleRate, 18.0))

	// Footsteps — very short noise bursts with material-specific decay.
	bank.Store(SoundFootstepStone, GenerateNoiseBurst(0.06, sampleRate, 25.0))
	bank.Store(SoundFootstepDirt, GenerateNoiseBurst(0.07, sampleRate, 20.0))
	bank.Store(SoundFootstepWood, GenerateNoiseBurst(0.05, sampleRate, 30.0))
	bank.Store(SoundFootstepGrass, GenerateNoiseBurst(0.08, sampleRate, 15.0))
	bank.Store(SoundFootstepSand, GenerateNoiseBurst(0.09, sampleRate, 10.0))

	// Character sounds — chirps at varying frequency ranges.
	bank.Store(SoundHurt, GenerateChirp(0.20, sampleRate, 600.0, 200.0))
	bank.Store(SoundDeath, GenerateChirp(0.50, sampleRate, 500.0, 100.0))
	bank.Store(SoundEat, GenerateChirp(0.15, sampleRate, 300.0, 500.0))

	// UI / environment sounds.
	bank.Store(SoundLevelUp, GenerateSineBeep(0.30, sampleRate, 880.0))
	bank.Store(SoundExplosion, GenerateNoiseBurst(0.40, sampleRate, 5.0))
	bank.Store(SoundSplash, GenerateNoiseBurst(0.25, sampleRate, 8.0))
}

// PlaySound plays the named sound event at full volume.
func (sm *SoundManager) PlaySound(event string) error {
	return sm.engine.Play(event)
}

// PlaySoundAt plays the named sound event with 3D positional attenuation.
func (sm *SoundManager) PlaySoundAt(event string, x, y, z, lx, ly, lz float32) error {
	return sm.engine.PlayAt(event, x, y, z, lx, ly, lz)
}

// AllSoundEvents returns a copy of every registered sound event name.
func AllSoundEvents() []string {
	out := make([]string, len(allSoundEvents))
	copy(out, allSoundEvents)
	return out
}
