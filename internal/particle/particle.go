package particle

import (
	"math"
	"math/rand"
	"sync"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// Particle represents a single visual particle with physics properties.
type Particle struct {
	Position mcmath.Vec3
	Velocity mcmath.Vec3
	Color    [4]float32 // RGBA
	Size     float32
	Life     float32 // remaining seconds
	MaxLife  float32
	Gravity  float32
}

// alive reports whether the particle still has remaining life.
func (p Particle) alive() bool {
	return p.Life > 0
}

// ParticleSystem manages a pool of particles with a fixed capacity.
// All exported methods are safe for concurrent use. Emit may be called from
// game-event goroutines while Update runs on the main tick goroutine.
type ParticleSystem struct {
	mu        sync.Mutex
	particles []Particle
	maxP      int
}

// NewParticleSystem creates a ParticleSystem with the given maximum capacity.
// If maxParticles is <= 0 the default of 1000 is used.
func NewParticleSystem(maxParticles int) *ParticleSystem {
	if maxParticles <= 0 {
		maxParticles = 1000
	}
	return &ParticleSystem{
		particles: make([]Particle, 0, maxParticles),
		maxP:      maxParticles,
	}
}

// Emit adds a single particle to the system.
// If the system is at capacity the oldest particle (index 0) is dropped.
func (ps *ParticleSystem) Emit(p Particle) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.emitLocked(p)
}

// emitLocked is the lock-free core of Emit. Callers must hold ps.mu.
func (ps *ParticleSystem) emitLocked(p Particle) {
	if len(ps.particles) >= ps.maxP {
		// Drop the oldest particle.
		ps.particles = ps.particles[1:]
	}
	ps.particles = append(ps.particles, p)
}

// EmitBurst emits count particles radiating outward from center with random
// directions. Each particle receives the supplied color, size, and life values.
// The entire burst is emitted under a single lock acquisition.
func (ps *ParticleSystem) EmitBurst(center mcmath.Vec3, count int, color [4]float32, speed, size, life float32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for i := 0; i < count; i++ {
		dir := randomDirection()
		p := Particle{
			Position: center,
			Velocity: dir.Scale(speed),
			Color:    color,
			Size:     size,
			Life:     life,
			MaxLife:  life,
			Gravity:  9.8,
		}
		ps.emitLocked(p)
	}
}

// Update advances the particle simulation by dt seconds.
// For each particle it applies gravity to velocity, integrates position,
// decreases life, and removes dead particles.
func (ps *ParticleSystem) Update(dt float32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	n := 0
	for i := range ps.particles {
		p := ps.particles[i]
		// Apply gravity (negative Y).
		p.Velocity = mcmath.Vec3{
			X: p.Velocity.X,
			Y: p.Velocity.Y - p.Gravity*dt,
			Z: p.Velocity.Z,
		}
		// Integrate position.
		p.Position = p.Position.Add(p.Velocity.Scale(dt))
		// Decrease life.
		p.Life -= dt

		if p.alive() {
			ps.particles[n] = p
			n++
		}
	}
	ps.particles = ps.particles[:n]
}

// Particles returns a copy of the active particle slice for rendering.
func (ps *ParticleSystem) Particles() []Particle {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	out := make([]Particle, len(ps.particles))
	copy(out, ps.particles)
	return out
}

// Count returns the number of active particles.
func (ps *ParticleSystem) Count() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return len(ps.particles)
}

// Clear removes all particles from the system.
func (ps *ParticleSystem) Clear() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.particles = ps.particles[:0]
}

// randomDirection returns a uniformly distributed unit vector on the sphere.
func randomDirection() mcmath.Vec3 {
	// Use spherical coordinates with uniform sampling.
	theta := rand.Float64() * 2 * math.Pi
	z := rand.Float64()*2 - 1 // uniform in [-1, 1]
	r := math.Sqrt(1 - z*z)
	return mcmath.Vec3{
		X: float32(r * math.Cos(theta)),
		Y: float32(z),
		Z: float32(r * math.Sin(theta)),
	}
}
