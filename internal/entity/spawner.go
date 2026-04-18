package entity

import (
	"math"
	"math/rand"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

const (
	// spawnCycleInterval is the minimum time in seconds between spawn attempts.
	spawnCycleInterval float64 = 10.0

	// spawnRadius is how far from the player mobs can spawn (in blocks).
	spawnRadius float32 = 24.0

	// spawnMinDistance is the minimum distance from the player for a spawn.
	spawnMinDistance float32 = 8.0

	// maxPassiveMobs caps the number of passive mobs allowed in the world.
	maxPassiveMobs int = 10

	// maxHostileMobs caps the number of hostile mobs allowed in the world.
	maxHostileMobs int = 8
)

// mobSpawnEntry pairs an entity type with its factory function.
type mobSpawnEntry struct {
	Spawn func(*ecs.World, mcmath.Vec3) ecs.Entity
}

// passiveMobTypes lists the factory functions for each passive mob type.
var passiveMobTypes = []mobSpawnEntry{
	{SpawnCow},
	{SpawnPig},
	{SpawnSheep},
	{SpawnChicken},
}

// hostileMobTypes lists the factory functions for each hostile mob type.
var hostileMobTypes = []mobSpawnEntry{
	{SpawnZombie},
	{SpawnSkeleton},
}

// Spawner controls periodic mob spawning around the player.
//
// Thread safety: Spawner is NOT safe for concurrent use. SpawnCycle must only
// be called from the main game tick goroutine, which is single-threaded with
// respect to ECS world access. This is enforced by the game loop in Game.tick.
type Spawner struct {
	Rand            *rand.Rand
	timer           float64
	HostileSpawnCap int
	PassiveSpawnCap int
	SpawnRadius     float32
}

// NewSpawner creates a Spawner with the given random source.
// If rng is nil, the global math/rand source is used.
func NewSpawner(rng *rand.Rand) *Spawner {
	return &Spawner{Rand: rng}
}

// SpawnCycle should be called every tick. It accumulates dt and, once
// the interval elapses, attempts to spawn passive and hostile mobs
// around the given player position.
func (s *Spawner) SpawnCycle(w *ecs.World, dt float64, playerPos mcmath.Vec3) {
	s.timer += dt
	if s.timer < spawnCycleInterval {
		return
	}
	s.timer -= spawnCycleInterval

	passiveCount, hostileCount := s.countMobs(w)

	if passiveCount < s.passiveSpawnCap() {
		s.spawnRandom(w, playerPos, passiveMobTypes)
	}
	if hostileCount < s.hostileSpawnCap() {
		s.spawnRandom(w, playerPos, hostileMobTypes)
	}
}

// countMobs returns the current number of passive and hostile mobs in the world.
func (s *Spawner) countMobs(w *ecs.World) (passive, hostile int) {
	etStore := ecs.GetStore[EntityTypeComp](w)
	etStore.Each(func(_ ecs.Entity, et *EntityTypeComp) {
		switch et.Type {
		case TypeCow, TypePig, TypeSheep, TypeChicken:
			passive++
		case TypeZombie, TypeSkeleton, TypeCreeper:
			hostile++
		}
	})
	return
}

// spawnRandom picks a random mob type from the given list and spawns it
// at a random position within the spawn ring around the player.
func (s *Spawner) spawnRandom(w *ecs.World, playerPos mcmath.Vec3, types []mobSpawnEntry) {
	if len(types) == 0 {
		return
	}

	idx := s.intn(len(types))
	pos := s.randomSpawnPos(playerPos)
	types[idx].Spawn(w, pos)
}

// randomSpawnPos generates a random position in the spawn ring
// (between spawnMinDistance and spawnRadius) around the player.
func (s *Spawner) randomSpawnPos(center mcmath.Vec3) mcmath.Vec3 {
	angle := s.float64() * 2 * math.Pi
	radius := s.effectiveSpawnRadius()
	dist := spawnMinDistance + float32(s.float64())*float32(radius-spawnMinDistance)

	return mcmath.Vec3{
		X: center.X + dist*float32(math.Cos(angle)),
		Y: center.Y,
		Z: center.Z + dist*float32(math.Sin(angle)),
	}
}

// hostileSpawnCap returns the effective hostile mob cap.
func (s *Spawner) hostileSpawnCap() int {
	if s.HostileSpawnCap > 0 {
		return s.HostileSpawnCap
	}
	return maxHostileMobs
}

// passiveSpawnCap returns the effective passive mob cap.
func (s *Spawner) passiveSpawnCap() int {
	if s.PassiveSpawnCap > 0 {
		return s.PassiveSpawnCap
	}
	return maxPassiveMobs
}

// effectiveSpawnRadius returns the effective spawn radius.
func (s *Spawner) effectiveSpawnRadius() float32 {
	if s.SpawnRadius > 0 {
		return s.SpawnRadius
	}
	return spawnRadius
}

func (s *Spawner) intn(n int) int {
	if s.Rand != nil {
		return s.Rand.Intn(n)
	}
	return rand.Intn(n)
}

func (s *Spawner) float64() float64 {
	if s.Rand != nil {
		return s.Rand.Float64()
	}
	return rand.Float64()
}
