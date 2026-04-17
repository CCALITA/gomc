package mcmath

import "fmt"

const (
	// ChunkSize is the horizontal size (X and Z) of a chunk in blocks.
	ChunkSize = 16
	// ChunkHeight is the total vertical height of a chunk in blocks.
	ChunkHeight = 256
	// SectionHeight is the vertical height of a chunk section in blocks.
	SectionHeight = 16
)

// BlockPos represents a discrete block position with integer coordinates.
type BlockPos struct {
	X, Y, Z int32
}

// ToVec3 converts the block position to a Vec3 at the block's origin corner.
func (p BlockPos) ToVec3() Vec3 {
	return Vec3{float32(p.X), float32(p.Y), float32(p.Z)}
}

// ToChunkPos returns the chunk position containing this block.
func (p BlockPos) ToChunkPos() ChunkPos {
	return ChunkPos{
		X: floorDiv(p.X, ChunkSize),
		Z: floorDiv(p.Z, ChunkSize),
	}
}

// LocalPos returns the position of this block within its chunk (0-15 for X and Z).
func (p BlockPos) LocalPos() BlockPos {
	return BlockPos{
		X: floorMod(p.X, ChunkSize),
		Y: p.Y,
		Z: floorMod(p.Z, ChunkSize),
	}
}

// Neighbors returns the six blocks adjacent to this position.
func (p BlockPos) Neighbors() [6]BlockPos {
	return [6]BlockPos{
		{p.X, p.Y, p.Z - 1}, // North
		{p.X, p.Y, p.Z + 1}, // South
		{p.X + 1, p.Y, p.Z}, // East
		{p.X - 1, p.Y, p.Z}, // West
		{p.X, p.Y + 1, p.Z}, // Up
		{p.X, p.Y - 1, p.Z}, // Down
	}
}

// Above returns the block position one step above.
func (p BlockPos) Above() BlockPos {
	return BlockPos{p.X, p.Y + 1, p.Z}
}

// Below returns the block position one step below.
func (p BlockPos) Below() BlockPos {
	return BlockPos{p.X, p.Y - 1, p.Z}
}

// Offset returns a new block position offset by (dx, dy, dz).
func (p BlockPos) Offset(dx, dy, dz int32) BlockPos {
	return BlockPos{p.X + dx, p.Y + dy, p.Z + dz}
}

// String returns a string representation of the block position.
func (p BlockPos) String() string {
	return fmt.Sprintf("BlockPos(%d, %d, %d)", p.X, p.Y, p.Z)
}

// floorDiv performs floored integer division (rounds towards negative infinity).
func floorDiv(a, b int32) int32 {
	d := a / b
	// If the signs differ and there's a remainder, adjust downward.
	if (a^b) < 0 && d*b != a {
		d--
	}
	return d
}

// floorMod returns the floored modulus (always non-negative for positive b).
func floorMod(a, b int32) int32 {
	return a - floorDiv(a, b)*b
}
