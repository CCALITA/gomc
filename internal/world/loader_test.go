package world

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// waitForPending polls until the loader's pending queue drains to zero.
func waitForPending(t *testing.T, cl *ChunkLoader) {
	t.Helper()
	require.Eventually(t, func() bool {
		return cl.PendingLoads() == 0
	}, 10*time.Second, 50*time.Millisecond, "all pending loads should complete")
}

func TestNewChunkLoaderDefaults(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)
	defer cl.Stop()

	assert.Equal(t, 8, cl.LoadRadius, "default LoadRadius should be 8")
	assert.Equal(t, 0, cl.PendingLoads(), "new loader should have zero pending loads")
	assert.Equal(t, 0, cl.LoadedCount(), "new loader should have zero loaded count")
}

func TestUpdateLoadsChunksWithinRadius(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)
	cl.LoadRadius = 2
	defer cl.Stop()

	center := mcmath.ChunkPos{X: 0, Z: 0}
	cl.Update(center)
	waitForPending(t, cl)

	radius := float64(cl.LoadRadius)
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			cp := mcmath.ChunkPos{X: center.X + dx, Z: center.Z + dz}
			if center.Distance(cp) <= radius {
				assert.NotNilf(t, w.GetChunk(cp), "chunk %v (distance %.2f) should be loaded", cp, center.Distance(cp))
			}
		}
	}
}

func TestUpdateUnloadsChunksOutsideRadiusPlusPadding(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)
	cl.LoadRadius = 2
	defer cl.Stop()

	origin := mcmath.ChunkPos{X: 0, Z: 0}
	cl.Update(origin)
	waitForPending(t, cl)

	assert.Greater(t, w.LoadedChunkCount(), 0, "some chunks should be loaded around origin")

	// Move the center far away so old chunks are outside radius+unloadPadding.
	farCenter := mcmath.ChunkPos{X: 50, Z: 50}
	cl.Update(farCenter)
	waitForPending(t, cl)

	unloadRadius := float64(cl.LoadRadius + unloadPadding)
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			cp := mcmath.ChunkPos{X: origin.X + dx, Z: origin.Z + dz}
			dist := farCenter.Distance(cp)
			if dist > unloadRadius {
				assert.Nilf(t, w.GetChunk(cp), "chunk %v (distance %.2f from new center) should be unloaded", cp, dist)
			}
		}
	}
}

func TestPendingLoadsReturnsQueueLength(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)
	cl.LoadRadius = 2
	defer cl.Stop()

	assert.Equal(t, 0, cl.PendingLoads())

	cl.Update(mcmath.ChunkPos{X: 0, Z: 0})
	waitForPending(t, cl)

	assert.Equal(t, 0, cl.PendingLoads(), "no pending loads after all chunks are generated")
}

func TestLoadedCountMatchesWorldLoadedChunkCount(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)
	cl.LoadRadius = 2
	defer cl.Stop()

	cl.Update(mcmath.ChunkPos{X: 0, Z: 0})
	waitForPending(t, cl)

	// Since this is a fresh world, every loaded chunk was loaded by the loader.
	assert.Equal(t, w.LoadedChunkCount(), cl.LoadedCount(),
		"loader LoadedCount should match world LoadedChunkCount for a fresh world")
}

func TestStopIsIdempotent(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)

	cl.Stop()
	cl.Stop()
}

func TestStopPreventsLoading(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)
	cl.LoadRadius = 2
	cl.Stop()

	countBefore := w.LoadedChunkCount()
	cl.Update(mcmath.ChunkPos{X: 0, Z: 0})

	// Give workers a moment even though they should be stopped.
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, countBefore, w.LoadedChunkCount(),
		"no new chunks should load after Stop")
}

func TestSmallRadiusLoadPattern(t *testing.T) {
	w := NewWorld(testSeed)
	cl := NewChunkLoader(w)
	cl.LoadRadius = 2
	defer cl.Stop()

	center := mcmath.ChunkPos{X: 5, Z: -3}
	cl.Update(center)
	waitForPending(t, cl)

	// Count expected chunks: those within Euclidean distance 2 of center.
	expectedCount := 0
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			cp := mcmath.ChunkPos{X: center.X + dx, Z: center.Z + dz}
			if center.Distance(cp) <= float64(cl.LoadRadius) {
				expectedCount++
			}
		}
	}

	assert.Equal(t, expectedCount, w.LoadedChunkCount(),
		"loaded chunk count should match expected count for radius 2")
	assert.Equal(t, expectedCount, cl.LoadedCount(),
		"loader LoadedCount should match expected count for radius 2")
}
