package noise_test

import (
	"math"
	"testing"

	"github.com/fanxiyao/gomc/internal/noise"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Determinism — same seed must produce identical output
// ---------------------------------------------------------------------------

func TestNoise2D_Deterministic(t *testing.T) {
	g1 := noise.NewNoiseGenerator(42)
	g2 := noise.NewNoiseGenerator(42)

	for _, pt := range [][2]float64{{0, 0}, {1.5, -3.7}, {100, 200}} {
		assert.Equal(t, g1.Noise2D(pt[0], pt[1]), g2.Noise2D(pt[0], pt[1]),
			"Noise2D should be deterministic for seed=42 at (%v,%v)", pt[0], pt[1])
	}
}

func TestNoise3D_Deterministic(t *testing.T) {
	g1 := noise.NewNoiseGenerator(42)
	g2 := noise.NewNoiseGenerator(42)

	for _, pt := range [][3]float64{{0, 0, 0}, {1.5, -3.7, 8.2}, {100, 200, 300}} {
		assert.Equal(t, g1.Noise3D(pt[0], pt[1], pt[2]), g2.Noise3D(pt[0], pt[1], pt[2]),
			"Noise3D should be deterministic for seed=42 at (%v,%v,%v)", pt[0], pt[1], pt[2])
	}
}

// ---------------------------------------------------------------------------
// Range — raw noise must be in [-1, 1]
// ---------------------------------------------------------------------------

func TestNoise2D_Range(t *testing.T) {
	gen := noise.NewNoiseGenerator(123)
	for x := -50.0; x <= 50.0; x += 0.7 {
		for z := -50.0; z <= 50.0; z += 0.7 {
			v := gen.Noise2D(x, z)
			assert.True(t, v >= -1.0 && v <= 1.0,
				"Noise2D(%v,%v) = %v, want in [-1,1]", x, z, v)
		}
	}
}

func TestNoise3D_Range(t *testing.T) {
	gen := noise.NewNoiseGenerator(123)
	for x := -10.0; x <= 10.0; x += 1.3 {
		for y := -10.0; y <= 10.0; y += 1.3 {
			for z := -10.0; z <= 10.0; z += 1.3 {
				v := gen.Noise3D(x, y, z)
				assert.True(t, v >= -1.0 && v <= 1.0,
					"Noise3D(%v,%v,%v) = %v, want in [-1,1]", x, y, z, v)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Seed accessor
// ---------------------------------------------------------------------------

func TestNoiseGenerator_Seed(t *testing.T) {
	gen := noise.NewNoiseGenerator(999)
	assert.Equal(t, int64(999), gen.Seed())
}

// ---------------------------------------------------------------------------
// Octave noise — more octaves produce a different distribution
// ---------------------------------------------------------------------------

func TestOctaveNoise_DiffersFromSingleOctave(t *testing.T) {
	gen := noise.NewNoiseGenerator(7)
	single := noise.NewOctaveNoise(gen, 1, 0.5, 2.0)
	multi := noise.NewOctaveNoise(gen, 6, 0.5, 2.0)

	differs := false
	for x := 0.0; x < 20.0; x += 0.5 {
		for z := 0.0; z < 20.0; z += 0.5 {
			if single.Sample2D(x, z) != multi.Sample2D(x, z) {
				differs = true
				break
			}
		}
		if differs {
			break
		}
	}
	assert.True(t, differs, "6-octave noise should differ from single-octave at some point")
}

func TestOctaveNoise_Sample2D_Range(t *testing.T) {
	gen := noise.NewNoiseGenerator(55)
	oct := noise.NewOctaveNoise(gen, 4, 0.5, 2.0)

	for x := -30.0; x <= 30.0; x += 0.9 {
		for z := -30.0; z <= 30.0; z += 0.9 {
			v := oct.Sample2D(x, z)
			assert.True(t, v >= -1.0 && v <= 1.0,
				"OctaveNoise.Sample2D(%v,%v) = %v, want in [-1,1]", x, z, v)
		}
	}
}

func TestOctaveNoise_Sample3D_Range(t *testing.T) {
	gen := noise.NewNoiseGenerator(55)
	oct := noise.NewOctaveNoise(gen, 4, 0.5, 2.0)

	for x := -10.0; x <= 10.0; x += 2.0 {
		for y := -10.0; y <= 10.0; y += 2.0 {
			for z := -10.0; z <= 10.0; z += 2.0 {
				v := oct.Sample3D(x, y, z)
				assert.True(t, v >= -1.0 && v <= 1.0,
					"OctaveNoise.Sample3D(%v,%v,%v) = %v, want in [-1,1]", x, y, z, v)
			}
		}
	}
}

func TestOctaveNoise_Deterministic(t *testing.T) {
	g1 := noise.NewNoiseGenerator(42)
	g2 := noise.NewNoiseGenerator(42)
	o1 := noise.NewOctaveNoise(g1, 4, 0.5, 2.0)
	o2 := noise.NewOctaveNoise(g2, 4, 0.5, 2.0)

	assert.Equal(t, o1.Sample2D(5.5, 3.3), o2.Sample2D(5.5, 3.3))
	assert.Equal(t, o1.Sample3D(1.0, 2.0, 3.0), o2.Sample3D(1.0, 2.0, 3.0))
}

// ---------------------------------------------------------------------------
// Ridged noise — output must be non-negative
// ---------------------------------------------------------------------------

func TestRidgedNoise_NonNegative2D(t *testing.T) {
	gen := noise.NewNoiseGenerator(77)
	ridged := noise.NewRidgedNoise(gen, 4, 0.5, 2.0)

	for x := -30.0; x <= 30.0; x += 0.9 {
		for z := -30.0; z <= 30.0; z += 0.9 {
			v := ridged.Sample2D(x, z)
			assert.True(t, v >= 0.0,
				"RidgedNoise.Sample2D(%v,%v) = %v, want >= 0", x, z, v)
			assert.True(t, v <= 1.0,
				"RidgedNoise.Sample2D(%v,%v) = %v, want <= 1", x, z, v)
		}
	}
}

func TestRidgedNoise_NonNegative3D(t *testing.T) {
	gen := noise.NewNoiseGenerator(77)
	ridged := noise.NewRidgedNoise(gen, 4, 0.5, 2.0)

	for x := -10.0; x <= 10.0; x += 2.0 {
		for y := -10.0; y <= 10.0; y += 2.0 {
			for z := -10.0; z <= 10.0; z += 2.0 {
				v := ridged.Sample3D(x, y, z)
				assert.True(t, v >= 0.0,
					"RidgedNoise.Sample3D(%v,%v,%v) = %v, want >= 0", x, y, z, v)
				assert.True(t, v <= 1.0,
					"RidgedNoise.Sample3D(%v,%v,%v) = %v, want <= 1", x, y, z, v)
			}
		}
	}
}

func TestRidgedNoise_Deterministic(t *testing.T) {
	g1 := noise.NewNoiseGenerator(42)
	g2 := noise.NewNoiseGenerator(42)
	r1 := noise.NewRidgedNoise(g1, 4, 0.5, 2.0)
	r2 := noise.NewRidgedNoise(g2, 4, 0.5, 2.0)

	assert.Equal(t, r1.Sample2D(5.5, 3.3), r2.Sample2D(5.5, 3.3))
	assert.Equal(t, r1.Sample3D(1.0, 2.0, 3.0), r2.Sample3D(1.0, 2.0, 3.0))
}

// ---------------------------------------------------------------------------
// HeightMap — reasonable values for default terrain parameters
// ---------------------------------------------------------------------------

func TestHeightMap_ReasonableRange(t *testing.T) {
	gen := noise.NewNoiseGenerator(12345)
	oct := noise.NewOctaveNoise(gen, 4, 0.5, 2.0)
	hm := noise.NewHeightMap(oct, 64, 32) // base 64 +/- 32 => 32..96

	for wx := -200; wx <= 200; wx += 3 {
		for wz := -200; wz <= 200; wz += 3 {
			h := hm.HeightAt(wx, wz)
			assert.True(t, h >= 32 && h <= 96,
				"HeightAt(%d,%d) = %d, want in [32,96]", wx, wz, h)
		}
	}
}

func TestHeightMap_Deterministic(t *testing.T) {
	makeHM := func() *noise.HeightMap {
		gen := noise.NewNoiseGenerator(42)
		oct := noise.NewOctaveNoise(gen, 4, 0.5, 2.0)
		return noise.NewHeightMap(oct, 64, 32)
	}
	hm1 := makeHM()
	hm2 := makeHM()

	for wx := -10; wx <= 10; wx++ {
		for wz := -10; wz <= 10; wz++ {
			assert.Equal(t, hm1.HeightAt(wx, wz), hm2.HeightAt(wx, wz),
				"HeightMap should be deterministic at (%d,%d)", wx, wz)
		}
	}
}

func TestHeightMap_NotFlat(t *testing.T) {
	gen := noise.NewNoiseGenerator(7)
	oct := noise.NewOctaveNoise(gen, 4, 0.5, 2.0)
	hm := noise.NewHeightMap(oct, 64, 32)

	heights := make(map[int]bool)
	for wx := 0; wx < 100; wx++ {
		heights[hm.HeightAt(wx, 0)] = true
	}
	assert.True(t, len(heights) > 1, "HeightMap should produce varying heights, got %d unique", len(heights))
}

// ---------------------------------------------------------------------------
// Different seeds produce different noise
// ---------------------------------------------------------------------------

func TestDifferentSeeds_ProduceDifferentNoise(t *testing.T) {
	g1 := noise.NewNoiseGenerator(1)
	g2 := noise.NewNoiseGenerator(2)

	differs := false
	for x := 0.0; x < 10.0; x += 0.3 {
		if math.Abs(g1.Noise2D(x, x)-g2.Noise2D(x, x)) > 1e-10 {
			differs = true
			break
		}
	}
	assert.True(t, differs, "Different seeds should produce different noise")
}
