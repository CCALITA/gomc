package particle

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitAddsParticle(t *testing.T) {
	ps := NewParticleSystem(10)
	assert.Equal(t, 0, ps.Count())

	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0})
	assert.Equal(t, 1, ps.Count())
}

func TestUpdateMovesParticles(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.Emit(Particle{
		Position: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Velocity: mcmath.Vec3{X: 1, Y: 0, Z: 0},
		Life:     5.0,
		MaxLife:  5.0,
		Gravity:  0,
	})

	ps.Update(1.0)

	particles := ps.Particles()
	require.Len(t, particles, 1)
	assert.InDelta(t, 1.0, float64(particles[0].Position.X), 0.001)
}

func TestDeadParticlesRemoved(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.Emit(Particle{Life: 0.5, MaxLife: 0.5})

	ps.Update(1.0) // dt > life, so particle should die
	assert.Equal(t, 0, ps.Count())
}

func TestMaxParticlesEnforcedDropsOldest(t *testing.T) {
	ps := NewParticleSystem(3)

	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0, Size: 1})
	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0, Size: 2})
	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0, Size: 3})
	assert.Equal(t, 3, ps.Count())

	// Emitting a 4th should drop the oldest (Size=1).
	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0, Size: 4})
	assert.Equal(t, 3, ps.Count())

	particles := ps.Particles()
	assert.InDelta(t, 2.0, float64(particles[0].Size), 0.001)
	assert.InDelta(t, 3.0, float64(particles[1].Size), 0.001)
	assert.InDelta(t, 4.0, float64(particles[2].Size), 0.001)
}

func TestBurstCreatesCorrectCount(t *testing.T) {
	ps := NewParticleSystem(100)
	color := [4]float32{1, 0, 0, 1}
	ps.EmitBurst(mcmath.Vec3{}, 25, color, 1.0, 0.1, 1.0)
	assert.Equal(t, 25, ps.Count())
}

func TestGravityAffectsVelocity(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.Emit(Particle{
		Velocity: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Life:     5.0,
		MaxLife:  5.0,
		Gravity:  10.0,
	})

	ps.Update(1.0)

	particles := ps.Particles()
	require.Len(t, particles, 1)
	// After 1s with gravity=10, Y velocity should be -10.
	assert.InDelta(t, -10.0, float64(particles[0].Velocity.Y), 0.001)
}

func TestBlockBreakPresetCreates8Particles(t *testing.T) {
	pos := mcmath.BlockPos{X: 5, Y: 10, Z: 3}
	color := [4]float32{0.5, 0.3, 0.1, 1.0}
	particles := BlockBreakParticles(pos, color)
	assert.Len(t, particles, 8)

	for _, p := range particles {
		assert.Equal(t, color, p.Color)
		assert.InDelta(t, 0.1, float64(p.Size), 0.001)
		assert.InDelta(t, 0.5, float64(p.Life), 0.001)
		// Position should be at block center.
		assert.InDelta(t, 5.5, float64(p.Position.X), 0.001)
		assert.InDelta(t, 10.5, float64(p.Position.Y), 0.001)
		assert.InDelta(t, 3.5, float64(p.Position.Z), 0.001)
	}
}

func TestTorchCreatesUpwardMovingParticle(t *testing.T) {
	pos := mcmath.Vec3{X: 1, Y: 2, Z: 3}
	p := TorchFlameParticle(pos)

	assert.Equal(t, pos, p.Position)
	assert.Greater(t, p.Velocity.Y, float32(0), "torch particle should move upward")
	assert.InDelta(t, 0.05, float64(p.Size), 0.001)
	assert.InDelta(t, 1.0, float64(p.Life), 0.001)
}

func TestSplashParticlesCount(t *testing.T) {
	pos := mcmath.Vec3{X: 0, Y: 0, Z: 0}
	particles := SplashParticles(pos)
	assert.Len(t, particles, 12)

	for _, p := range particles {
		assert.Greater(t, p.Velocity.Y, float32(0), "splash particles should move upward")
		assert.InDelta(t, 0.8, float64(p.Life), 0.001)
	}
}

func TestClearRemovesAll(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0})
	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0})
	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0})
	assert.Equal(t, 3, ps.Count())

	ps.Clear()
	assert.Equal(t, 0, ps.Count())
}

func TestNewParticleSystemDefaultMax(t *testing.T) {
	ps := NewParticleSystem(0)
	assert.NotNil(t, ps)
	// Should use default of 1000; fill to capacity.
	for i := 0; i < 1001; i++ {
		ps.Emit(Particle{Life: 1.0, MaxLife: 1.0})
	}
	assert.Equal(t, 1000, ps.Count())
}

func TestParticlesReturnsCopy(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.Emit(Particle{Life: 1.0, MaxLife: 1.0, Size: 5})

	snap := ps.Particles()
	snap[0].Size = 999

	// Original should be unaffected.
	assert.InDelta(t, 5.0, float64(ps.Particles()[0].Size), 0.001)
}

func TestUpdateMultipleParticlesMixedLife(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.Emit(Particle{Life: 0.1, MaxLife: 0.5}) // will die
	ps.Emit(Particle{Life: 2.0, MaxLife: 2.0}) // will survive

	ps.Update(0.5)
	assert.Equal(t, 1, ps.Count())
}
