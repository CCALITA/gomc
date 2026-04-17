package noise

import "math"

// RidgedNoise produces ridge-like terrain features by using 1 - |noise|
// instead of the raw noise value, then layering multiple octaves.
type RidgedNoise struct {
	gen         *NoiseGenerator
	octaves     int
	persistence float64
	lacunarity  float64
}

// NewRidgedNoise creates a RidgedNoise generator.
func NewRidgedNoise(gen *NoiseGenerator, octaves int, persistence, lacunarity float64) *RidgedNoise {
	return &RidgedNoise{
		gen:         gen,
		octaves:     octaves,
		persistence: persistence,
		lacunarity:  lacunarity,
	}
}

// Sample2D returns ridged fractal noise in 2D.
// The result is in [0, 1] because each octave contribution is 1 - |noise|.
func (r *RidgedNoise) Sample2D(x, z float64) float64 {
	var total float64
	amplitude := 1.0
	frequency := 1.0
	maxAmplitude := 0.0

	for i := 0; i < r.octaves; i++ {
		raw := r.gen.Noise2D(x*frequency, z*frequency)
		total += (1.0 - math.Abs(raw)) * amplitude
		maxAmplitude += amplitude
		amplitude *= r.persistence
		frequency *= r.lacunarity
	}

	return total / maxAmplitude
}

// Sample3D returns ridged fractal noise in 3D.
// The result is in [0, 1].
func (r *RidgedNoise) Sample3D(x, y, z float64) float64 {
	var total float64
	amplitude := 1.0
	frequency := 1.0
	maxAmplitude := 0.0

	for i := 0; i < r.octaves; i++ {
		raw := r.gen.Noise3D(x*frequency, y*frequency, z*frequency)
		total += (1.0 - math.Abs(raw)) * amplitude
		maxAmplitude += amplitude
		amplitude *= r.persistence
		frequency *= r.lacunarity
	}

	return total / maxAmplitude
}
