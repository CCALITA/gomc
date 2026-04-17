// Package noise provides coherent noise generators for terrain generation.
// It wraps opensimplex noise with octave and ridged variants, plus a
// height-map helper that converts continuous noise into discrete block heights.
package noise

import opensimplex "github.com/ojrac/opensimplex-go"

// NoiseGenerator wraps an opensimplex noise instance seeded once at creation.
type NoiseGenerator struct {
	noise opensimplex.Noise
	seed  int64
}

// NewNoiseGenerator creates a NoiseGenerator for the given seed.
func NewNoiseGenerator(seed int64) *NoiseGenerator {
	return &NoiseGenerator{
		noise: opensimplex.New(seed),
		seed:  seed,
	}
}

// Noise2D returns coherent noise in [-1, 1] for the given x, z coordinates.
func (g *NoiseGenerator) Noise2D(x, z float64) float64 {
	return g.noise.Eval2(x, z)
}

// Noise3D returns coherent noise in [-1, 1] for the given x, y, z coordinates.
func (g *NoiseGenerator) Noise3D(x, y, z float64) float64 {
	return g.noise.Eval3(x, y, z)
}

// Seed returns the seed used to create this generator.
func (g *NoiseGenerator) Seed() int64 {
	return g.seed
}
