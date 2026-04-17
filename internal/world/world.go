package world

import (
	"math"
	"sync"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// World manages loaded chunks and provides block-level access to the voxel world.
type World struct {
	mu        sync.RWMutex
	chunks    map[[2]int32]*chunk.Chunk
	generator *TerrainGenerator
}

// NewWorld creates a new World with the given seed for terrain generation.
func NewWorld(seed int64) *World {
	return &World{
		chunks:    make(map[[2]int32]*chunk.Chunk),
		generator: NewTerrainGenerator(seed),
	}
}

// GetChunk returns the loaded chunk at pos, or nil if it is not loaded.
func (w *World) GetChunk(pos mcmath.ChunkPos) *chunk.Chunk {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.chunks[pos.Key()]
}

// LoadChunk returns the chunk at pos, generating it if it is not already loaded.
func (w *World) LoadChunk(pos mcmath.ChunkPos) *chunk.Chunk {
	key := pos.Key()

	w.mu.RLock()
	c, ok := w.chunks[key]
	w.mu.RUnlock()
	if ok {
		return c
	}

	c = w.generator.GenerateChunk(pos)

	w.mu.Lock()
	// Double-check after acquiring write lock.
	if existing, ok := w.chunks[key]; ok {
		w.mu.Unlock()
		return existing
	}
	w.chunks[key] = c
	w.mu.Unlock()
	return c
}

// UnloadChunk removes the chunk at pos from the world.
func (w *World) UnloadChunk(pos mcmath.ChunkPos) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.chunks, pos.Key())
}

// GetBlock returns the block ID at the given world position.
// Returns block.Air if the containing chunk is not loaded or the position is out of range.
func (w *World) GetBlock(pos mcmath.BlockPos) uint16 {
	if pos.Y < 0 || pos.Y >= mcmath.ChunkHeight {
		return block.Air
	}
	cp := pos.ToChunkPos()
	c := w.GetChunk(cp)
	if c == nil {
		return block.Air
	}
	local := pos.LocalPos()
	return c.GetBlock(int(local.X), int(local.Y), int(local.Z))
}

// SetBlock sets the block ID at the given world position.
// The operation is a no-op if the containing chunk is not loaded.
func (w *World) SetBlock(pos mcmath.BlockPos, id uint16) {
	if pos.Y < 0 || pos.Y >= mcmath.ChunkHeight {
		return
	}
	cp := pos.ToChunkPos()
	w.mu.RLock()
	c := w.chunks[cp.Key()]
	w.mu.RUnlock()
	if c == nil {
		return
	}
	local := pos.LocalPos()
	c.SetBlock(int(local.X), int(local.Y), int(local.Z), id)
}

// GetBlockAABBs returns the AABBs of all solid blocks that intersect the given region.
// This is used for physics collision detection.
func (w *World) GetBlockAABBs(region mcmath.AABB) []mcmath.AABB {
	minX := int32(math.Floor(float64(region.Min.X)))
	minY := int32(math.Floor(float64(region.Min.Y)))
	minZ := int32(math.Floor(float64(region.Min.Z)))
	maxX := int32(math.Ceil(float64(region.Max.X)))
	maxY := int32(math.Ceil(float64(region.Max.Y)))
	maxZ := int32(math.Ceil(float64(region.Max.Z)))

	if minY < 0 {
		minY = 0
	}
	if maxY > mcmath.ChunkHeight {
		maxY = mcmath.ChunkHeight
	}

	var result []mcmath.AABB
	for x := minX; x < maxX; x++ {
		for z := minZ; z < maxZ; z++ {
			for y := minY; y < maxY; y++ {
				bp := mcmath.BlockPos{X: x, Y: y, Z: z}
				id := w.GetBlock(bp)
				if block.IsSolid(id) {
					result = append(result, mcmath.BlockAABB(bp))
				}
			}
		}
	}
	return result
}

// LoadedChunkCount returns the number of currently loaded chunks.
func (w *World) LoadedChunkCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.chunks)
}
