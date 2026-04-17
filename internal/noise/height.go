package noise

import "math"

// HeightMap converts continuous octave noise into discrete block heights
// suitable for terrain generation.
type HeightMap struct {
	noise      *OctaveNoise
	baseHeight float64 // centre height (e.g. 64)
	amplitude  float64 // half-range around base (e.g. 32 => heights 32-96)
}

// NewHeightMap creates a HeightMap that maps noise values to integer heights.
//   - baseHeight is the centre of the height range.
//   - amplitude is the maximum deviation from baseHeight.
func NewHeightMap(noise *OctaveNoise, baseHeight, amplitude float64) *HeightMap {
	return &HeightMap{
		noise:      noise,
		baseHeight: baseHeight,
		amplitude:  amplitude,
	}
}

// HeightAt returns the terrain height at the given world-space block column.
func (h *HeightMap) HeightAt(worldX, worldZ int) int {
	// Scale world coordinates so that features have a reasonable size.
	const scale = 0.01
	n := h.noise.Sample2D(float64(worldX)*scale, float64(worldZ)*scale)
	return int(math.Round(h.baseHeight + n*h.amplitude))
}
