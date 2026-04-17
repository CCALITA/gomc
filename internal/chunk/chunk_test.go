package chunk

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Section tests
// ---------------------------------------------------------------------------

func TestSection_GetSetBasic(t *testing.T) {
	s := NewSection()

	// Default is air.
	assert.Equal(t, uint16(0), s.GetBlock(0, 0, 0))
	assert.True(t, s.IsEmpty())

	// Place a block.
	s.SetBlock(5, 3, 7, 1)
	assert.Equal(t, uint16(1), s.GetBlock(5, 3, 7))
	assert.False(t, s.IsEmpty())

	// Overwrite with air.
	s.SetBlock(5, 3, 7, 0)
	assert.Equal(t, uint16(0), s.GetBlock(5, 3, 7))
	assert.True(t, s.IsEmpty())
}

func TestSection_MultipleBlockTypes(t *testing.T) {
	s := NewSection()
	s.SetBlock(0, 0, 0, 1) // stone
	s.SetBlock(1, 0, 0, 2) // dirt
	s.SetBlock(2, 0, 0, 3) // grass

	assert.Equal(t, uint16(1), s.GetBlock(0, 0, 0))
	assert.Equal(t, uint16(2), s.GetBlock(1, 0, 0))
	assert.Equal(t, uint16(3), s.GetBlock(2, 0, 0))

	// Palette should have air + 3 types = 4.
	assert.Equal(t, 4, s.PaletteSize())
}

func TestSection_PaletteGrowth(t *testing.T) {
	s := NewSection()

	// Fill with 260 distinct block types to trigger wide storage migration.
	id := uint16(1)
	for y := 0; y < 16; y++ {
		for z := 0; z < 16; z++ {
			s.SetBlock(0, y, z, id)
			id++
			if id > 260 {
				break
			}
		}
		if id > 260 {
			break
		}
	}

	assert.True(t, s.PaletteSize() > 256, "palette should have grown past 256")

	// Verify a few values survive the migration.
	assert.Equal(t, uint16(1), s.GetBlock(0, 0, 0))
	assert.Equal(t, uint16(17), s.GetBlock(0, 1, 0))
}

func TestSection_AllPositions(t *testing.T) {
	s := NewSection()
	// Set all blocks to ID 42.
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				s.SetBlock(x, y, z, 42)
			}
		}
	}
	assert.False(t, s.IsEmpty())

	// Verify.
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				assert.Equal(t, uint16(42), s.GetBlock(x, y, z))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Chunk tests
// ---------------------------------------------------------------------------

func TestChunk_GetSetAcrossSections(t *testing.T) {
	c := &Chunk{Pos: mcmath.ChunkPos{X: 0, Z: 0}}

	// Set blocks in different sections.
	c.SetBlock(0, 0, 0, 1)    // section 0
	c.SetBlock(0, 16, 0, 2)   // section 1
	c.SetBlock(0, 255, 0, 3)  // section 15

	assert.Equal(t, uint16(1), c.GetBlock(0, 0, 0))
	assert.Equal(t, uint16(2), c.GetBlock(0, 16, 0))
	assert.Equal(t, uint16(3), c.GetBlock(0, 255, 0))

	// Uninitialized section returns air.
	assert.Equal(t, uint16(0), c.GetBlock(8, 128, 8))
}

func TestChunk_LazyInit(t *testing.T) {
	c := &Chunk{}

	// All sections should be nil initially.
	for i := 0; i < 16; i++ {
		assert.Nil(t, c.GetSection(i))
	}

	// Setting a block should lazily initialise the section.
	c.SetBlock(0, 33, 0, 5) // section 2
	assert.NotNil(t, c.GetSection(2))
	assert.Nil(t, c.GetSection(0))
}

func TestChunk_HeightMap(t *testing.T) {
	c := &Chunk{}

	// Empty column.
	assert.Equal(t, -1, c.HighestBlock(0, 0))

	c.SetBlock(0, 10, 0, 1)
	assert.Equal(t, 10, c.HighestBlock(0, 0))

	c.SetBlock(0, 50, 0, 2)
	assert.Equal(t, 50, c.HighestBlock(0, 0))

	// Remove the highest block — should scan down.
	c.SetBlock(0, 50, 0, 0)
	assert.Equal(t, 10, c.HighestBlock(0, 0))

	// Remove the last block.
	c.SetBlock(0, 10, 0, 0)
	assert.Equal(t, -1, c.HighestBlock(0, 0))
}

func TestChunk_IsEmpty(t *testing.T) {
	c := &Chunk{}
	assert.True(t, c.IsEmpty())

	c.SetBlock(7, 7, 7, 1)
	assert.False(t, c.IsEmpty())

	c.SetBlock(7, 7, 7, 0)
	assert.True(t, c.IsEmpty())
}

func TestChunk_OutOfBounds(t *testing.T) {
	c := &Chunk{}

	// Out-of-bounds Y returns air and does not panic.
	assert.Equal(t, uint16(0), c.GetBlock(0, -1, 0))
	assert.Equal(t, uint16(0), c.GetBlock(0, 256, 0))

	// SetBlock with out-of-bounds Y is a no-op.
	c.SetBlock(0, -1, 0, 1)
	c.SetBlock(0, 300, 0, 1)
	assert.True(t, c.IsEmpty())
}

func TestChunk_GetSection_OutOfBounds(t *testing.T) {
	c := &Chunk{}
	assert.Nil(t, c.GetSection(-1))
	assert.Nil(t, c.GetSection(16))
}

// ---------------------------------------------------------------------------
// Mesh tests
// ---------------------------------------------------------------------------

func defaultSolid(id uint16) bool      { return id != 0 }
func defaultTransparent(id uint16) bool { return id == 0 }

func TestMesh_SingleBlock(t *testing.T) {
	c := &Chunk{}
	c.SetBlock(8, 8, 8, 1) // single block in the middle

	mesh := MeshChunk(c, [4]*Chunk{}, defaultSolid, defaultTransparent)

	// A single isolated block should produce 6 faces × 4 vertices = 24 vertices
	// and 6 faces × 6 indices = 36 indices.
	assert.Equal(t, 24*Stride, len(mesh.Vertices), "expected 24 vertices (6 faces)")
	assert.Equal(t, 36, len(mesh.Indices), "expected 36 indices (6 faces)")
}

func TestMesh_FullLayer(t *testing.T) {
	c := &Chunk{}
	// Fill an entire Y=0 layer.
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			c.SetBlock(x, 0, z, 1)
		}
	}

	mesh := MeshChunk(c, [4]*Chunk{}, defaultSolid, defaultTransparent)

	// Should produce some faces (at least top/bottom which can be greedy-merged).
	assert.Greater(t, len(mesh.Vertices), 0, "mesh should have vertices")
	assert.Greater(t, len(mesh.Indices), 0, "mesh should have indices")

	// Count quads by indices: each quad is 6 indices.
	numQuads := len(mesh.Indices) / 6

	// Greedy meshing merges each face direction into a single large quad:
	// Top (1) + Bottom (1) + North (1) + South (1) + East (1) + West (1) = 6
	assert.Equal(t, 6, numQuads, "expected 6 quads for a full layer (greedy merged)")
}

func TestMesh_CheckeredPattern(t *testing.T) {
	c := &Chunk{}
	// Checkerboard of two block types — no merging should happen.
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			if (x+z)%2 == 0 {
				c.SetBlock(x, 0, z, 1)
			} else {
				c.SetBlock(x, 0, z, 2)
			}
		}
	}

	mesh := MeshChunk(c, [4]*Chunk{}, defaultSolid, defaultTransparent)
	assert.Greater(t, len(mesh.Vertices), 0)
	assert.Greater(t, len(mesh.Indices), 0)
}

func TestMesh_EmptyChunk(t *testing.T) {
	c := &Chunk{}
	mesh := MeshChunk(c, [4]*Chunk{}, defaultSolid, defaultTransparent)
	assert.Equal(t, 0, len(mesh.Vertices))
	assert.Equal(t, 0, len(mesh.Indices))
}

func TestMesh_AdjacentBlocks_InternalFacesCulled(t *testing.T) {
	c := &Chunk{}
	// Two adjacent blocks along X — the shared face should be culled.
	c.SetBlock(0, 0, 0, 1)
	c.SetBlock(1, 0, 0, 1)

	mesh := MeshChunk(c, [4]*Chunk{}, defaultSolid, defaultTransparent)

	// 2 adjacent blocks share 2 internal faces (culled).
	// Greedy meshing merges the remaining exposed faces:
	// Top: 1 merged quad (2x1), Bottom: 1, North: 1, South: 1, West: 1, East: 1 = 6.
	numQuads := len(mesh.Indices) / 6
	assert.Equal(t, 6, numQuads, "expected 6 quads (shared faces culled, remaining merged)")
}

func TestMesh_VertexFormat(t *testing.T) {
	c := &Chunk{}
	c.SetBlock(0, 0, 0, 1)

	mesh := MeshChunk(c, [4]*Chunk{}, defaultSolid, defaultTransparent)

	// Verify vertex count is a multiple of Stride.
	assert.Equal(t, 0, len(mesh.Vertices)%Stride)

	// Check AO values are in valid range [0, 1].
	for i := 8; i < len(mesh.Vertices); i += Stride {
		ao := mesh.Vertices[i]
		assert.GreaterOrEqual(t, ao, float32(0))
		assert.LessOrEqual(t, ao, float32(1))
	}
}

func TestMesh_WithNeighborChunk(t *testing.T) {
	c := &Chunk{}
	// Place block at edge.
	c.SetBlock(0, 0, 0, 1)

	// Create a west neighbor with a block adjacent to our edge.
	west := &Chunk{}
	west.SetBlock(15, 0, 0, 1)

	neighbors := [4]*Chunk{nil, nil, nil, west}
	mesh := MeshChunk(c, neighbors, defaultSolid, defaultTransparent)

	// The west face of our block should be culled because the neighbor has a solid block.
	numQuads := len(mesh.Indices) / 6
	assert.Equal(t, 5, numQuads, "expected 5 quads (west face culled by neighbor)")
}

// ---------------------------------------------------------------------------
// Serialization tests
// ---------------------------------------------------------------------------

func TestSerialize_RoundTrip(t *testing.T) {
	c := &Chunk{Pos: mcmath.ChunkPos{X: 3, Z: -7}}

	// Place some blocks.
	c.SetBlock(0, 0, 0, 1)
	c.SetBlock(15, 255, 15, 42)
	c.SetBlock(8, 128, 8, 100)

	data, err := SerializeChunk(c)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	c2, err := DeserializeChunk(data)
	assert.NoError(t, err)

	// Verify position.
	assert.Equal(t, c.Pos, c2.Pos)

	// Verify blocks.
	assert.Equal(t, uint16(1), c2.GetBlock(0, 0, 0))
	assert.Equal(t, uint16(42), c2.GetBlock(15, 255, 15))
	assert.Equal(t, uint16(100), c2.GetBlock(8, 128, 8))
	assert.Equal(t, uint16(0), c2.GetBlock(5, 5, 5))

	// Verify height map.
	assert.Equal(t, c.HighestBlock(0, 0), c2.HighestBlock(0, 0))
	assert.Equal(t, c.HighestBlock(15, 15), c2.HighestBlock(15, 15))
	assert.Equal(t, c.HighestBlock(8, 8), c2.HighestBlock(8, 8))
}

func TestSerialize_EmptyChunk(t *testing.T) {
	c := &Chunk{Pos: mcmath.ChunkPos{X: 0, Z: 0}}

	data, err := SerializeChunk(c)
	assert.NoError(t, err)

	c2, err := DeserializeChunk(data)
	assert.NoError(t, err)
	assert.True(t, c2.IsEmpty())
	assert.Equal(t, c.Pos, c2.Pos)
}

func TestSerialize_FullSection(t *testing.T) {
	c := &Chunk{Pos: mcmath.ChunkPos{X: 1, Z: 2}}

	// Fill an entire section.
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				c.SetBlock(x, y, z, 7)
			}
		}
	}

	data, err := SerializeChunk(c)
	assert.NoError(t, err)

	c2, err := DeserializeChunk(data)
	assert.NoError(t, err)

	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				assert.Equal(t, uint16(7), c2.GetBlock(x, y, z))
			}
		}
	}
}

func TestSerialize_MultipleBlockTypes(t *testing.T) {
	c := &Chunk{Pos: mcmath.ChunkPos{X: 5, Z: 5}}

	c.SetBlock(0, 0, 0, 1)
	c.SetBlock(1, 0, 0, 2)
	c.SetBlock(2, 0, 0, 3)
	c.SetBlock(0, 16, 0, 10)
	c.SetBlock(0, 32, 0, 20)

	data, err := SerializeChunk(c)
	assert.NoError(t, err)

	c2, err := DeserializeChunk(data)
	assert.NoError(t, err)

	assert.Equal(t, uint16(1), c2.GetBlock(0, 0, 0))
	assert.Equal(t, uint16(2), c2.GetBlock(1, 0, 0))
	assert.Equal(t, uint16(3), c2.GetBlock(2, 0, 0))
	assert.Equal(t, uint16(10), c2.GetBlock(0, 16, 0))
	assert.Equal(t, uint16(20), c2.GetBlock(0, 32, 0))
}

// ---------------------------------------------------------------------------
// blockIndex test
// ---------------------------------------------------------------------------

func TestBlockIndex(t *testing.T) {
	// Verify blockIndex produces unique values for all positions.
	seen := make(map[int]bool)
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				idx := blockIndex(x, y, z)
				assert.False(t, seen[idx], "duplicate index for (%d,%d,%d)", x, y, z)
				assert.GreaterOrEqual(t, idx, 0)
				assert.Less(t, idx, 4096)
				seen[idx] = true
			}
		}
	}
	assert.Equal(t, 4096, len(seen))
}
