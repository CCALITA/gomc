package particle

import (
	"math"
	"math/rand"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// BlockBreakParticles returns 8 particles radiating outward from the center
// of the given block position with the block's color.
func BlockBreakParticles(pos mcmath.BlockPos, blockColor [4]float32) []Particle {
	center := pos.ToVec3().Add(mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	particles := make([]Particle, 8)
	for i := range particles {
		dir := randomDirection()
		particles[i] = Particle{
			Position: center,
			Velocity: dir.Scale(2.0),
			Color:    blockColor,
			Size:     0.1,
			Life:     0.5,
			MaxLife:  0.5,
			Gravity:  4.0,
		}
	}
	return particles
}

// TorchFlameParticle returns a small orange/yellow particle floating upward
// from the given torch position.
func TorchFlameParticle(pos mcmath.Vec3) Particle {
	return Particle{
		Position: pos,
		Velocity: mcmath.Vec3{
			X: float32(rand.Float64()*0.2 - 0.1),
			Y: 0.5,
			Z: float32(rand.Float64()*0.2 - 0.1),
		},
		Color:   [4]float32{1.0, 0.6, 0.1, 1.0}, // orange/yellow
		Size:    0.05,
		Life:    1.0,
		MaxLife: 1.0,
		Gravity: 0.0,
	}
}

// SplashParticles returns blue particles for a water splash, radiating
// outward and upward from the given position.
func SplashParticles(pos mcmath.Vec3) []Particle {
	const count = 12
	particles := make([]Particle, count)
	blue := [4]float32{0.2, 0.4, 1.0, 0.8}

	for i := range particles {
		angle := float64(i) * (2 * math.Pi / count)
		particles[i] = Particle{
			Position: pos,
			Velocity: mcmath.Vec3{
				X: float32(math.Cos(angle) * 1.5),
				Y: float32(1.0 + rand.Float64()*1.0),
				Z: float32(math.Sin(angle) * 1.5),
			},
			Color:   blue,
			Size:    0.08,
			Life:    0.8,
			MaxLife: 0.8,
			Gravity: 6.0,
		}
	}
	return particles
}
