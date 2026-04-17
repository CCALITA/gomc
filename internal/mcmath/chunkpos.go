package mcmath

import (
	"fmt"
	"math"
)

// ChunkPos represents the position of a chunk in chunk coordinates.
type ChunkPos struct {
	X, Z int32
}

// WorldBlockX returns the minimum world X block coordinate for this chunk.
func (c ChunkPos) WorldBlockX() int32 {
	return c.X * ChunkSize
}

// WorldBlockZ returns the minimum world Z block coordinate for this chunk.
func (c ChunkPos) WorldBlockZ() int32 {
	return c.Z * ChunkSize
}

// BlockPosAt returns the world block position for a local block within this chunk.
// localX and localZ should be in [0, ChunkSize); y is the world Y coordinate.
func (c ChunkPos) BlockPosAt(localX, y, localZ int32) BlockPos {
	return BlockPos{
		X: c.X*ChunkSize + localX,
		Y: y,
		Z: c.Z*ChunkSize + localZ,
	}
}

// Distance returns the Euclidean distance between two chunk positions.
func (c ChunkPos) Distance(other ChunkPos) float64 {
	dx := float64(c.X - other.X)
	dz := float64(c.Z - other.Z)
	return math.Sqrt(dx*dx + dz*dz)
}

// Key returns a [2]int32 array suitable for use as a map key.
func (c ChunkPos) Key() [2]int32 {
	return [2]int32{c.X, c.Z}
}

// String returns a string representation of the chunk position.
func (c ChunkPos) String() string {
	return fmt.Sprintf("ChunkPos(%d, %d)", c.X, c.Z)
}
