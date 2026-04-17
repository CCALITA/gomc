package noise

// OctaveNoise layers multiple noise samples at increasing frequency and
// decreasing amplitude to produce fractal-like detail.
type OctaveNoise struct {
	gen         *NoiseGenerator
	octaves     int
	persistence float64 // amplitude multiplier per octave (typically 0.5)
	lacunarity  float64 // frequency multiplier per octave (typically 2.0)
}

// NewOctaveNoise creates an OctaveNoise that sums the given number of octaves.
//   - persistence controls how quickly amplitude decreases (0 < p < 1).
//   - lacunarity  controls how quickly frequency increases (> 1).
func NewOctaveNoise(gen *NoiseGenerator, octaves int, persistence, lacunarity float64) *OctaveNoise {
	return &OctaveNoise{
		gen:         gen,
		octaves:     octaves,
		persistence: persistence,
		lacunarity:  lacunarity,
	}
}

// Sample2D returns fractal noise in 2D by summing multiple octaves.
func (o *OctaveNoise) Sample2D(x, z float64) float64 {
	var total float64
	amplitude := 1.0
	frequency := 1.0
	maxAmplitude := 0.0

	for i := 0; i < o.octaves; i++ {
		total += o.gen.Noise2D(x*frequency, z*frequency) * amplitude
		maxAmplitude += amplitude
		amplitude *= o.persistence
		frequency *= o.lacunarity
	}

	return total / maxAmplitude
}

// Sample3D returns fractal noise in 3D by summing multiple octaves.
func (o *OctaveNoise) Sample3D(x, y, z float64) float64 {
	var total float64
	amplitude := 1.0
	frequency := 1.0
	maxAmplitude := 0.0

	for i := 0; i < o.octaves; i++ {
		total += o.gen.Noise3D(x*frequency, y*frequency, z*frequency) * amplitude
		maxAmplitude += amplitude
		amplitude *= o.persistence
		frequency *= o.lacunarity
	}

	return total / maxAmplitude
}
