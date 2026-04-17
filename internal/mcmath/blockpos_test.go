package mcmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockPos_ToVec3(t *testing.T) {
	p := BlockPos{1, 2, 3}
	v := p.ToVec3()
	assert.Equal(t, Vec3{1, 2, 3}, v)
}

func TestBlockPos_ToChunkPos(t *testing.T) {
	tests := []struct {
		name string
		pos  BlockPos
		want ChunkPos
	}{
		{"positive", BlockPos{17, 0, 33}, ChunkPos{1, 2}},
		{"origin", BlockPos{0, 0, 0}, ChunkPos{0, 0}},
		{"negative", BlockPos{-1, 0, -1}, ChunkPos{-1, -1}},
		{"negative_exact", BlockPos{-16, 0, -32}, ChunkPos{-1, -2}},
		{"boundary", BlockPos{15, 0, 15}, ChunkPos{0, 0}},
		{"boundary16", BlockPos{16, 0, 16}, ChunkPos{1, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.pos.ToChunkPos())
		})
	}
}

func TestBlockPos_LocalPos(t *testing.T) {
	tests := []struct {
		name string
		pos  BlockPos
		want BlockPos
	}{
		{"positive", BlockPos{17, 5, 33}, BlockPos{1, 5, 1}},
		{"origin", BlockPos{0, 0, 0}, BlockPos{0, 0, 0}},
		{"negative", BlockPos{-1, 0, -1}, BlockPos{15, 0, 15}},
		{"boundary", BlockPos{15, 0, 15}, BlockPos{15, 0, 15}},
		{"exact_negative", BlockPos{-16, 0, -16}, BlockPos{0, 0, 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			local := tc.pos.LocalPos()
			assert.Equal(t, tc.want, local)
			// Local X and Z must be in [0, 15]
			assert.True(t, local.X >= 0 && local.X < ChunkSize)
			assert.True(t, local.Z >= 0 && local.Z < ChunkSize)
		})
	}
}

func TestBlockPos_Neighbors(t *testing.T) {
	p := BlockPos{0, 0, 0}
	n := p.Neighbors()
	assert.Len(t, n, 6)
	expected := [6]BlockPos{
		{0, 0, -1},  // North
		{0, 0, 1},   // South
		{1, 0, 0},   // East
		{-1, 0, 0},  // West
		{0, 1, 0},   // Up
		{0, -1, 0},  // Down
	}
	assert.Equal(t, expected, n)
}

func TestBlockPos_Above(t *testing.T) {
	p := BlockPos{5, 10, 3}
	assert.Equal(t, BlockPos{5, 11, 3}, p.Above())
}

func TestBlockPos_Below(t *testing.T) {
	p := BlockPos{5, 10, 3}
	assert.Equal(t, BlockPos{5, 9, 3}, p.Below())
}

func TestBlockPos_Offset(t *testing.T) {
	p := BlockPos{1, 2, 3}
	result := p.Offset(4, 5, 6)
	assert.Equal(t, BlockPos{5, 7, 9}, result)
}

func TestBlockPos_String(t *testing.T) {
	p := BlockPos{1, 2, 3}
	s := p.String()
	assert.Contains(t, s, "BlockPos")
	assert.Contains(t, s, "1")
	assert.Contains(t, s, "2")
	assert.Contains(t, s, "3")
}

func TestBlockPos_Above_Below_Inverse(t *testing.T) {
	p := BlockPos{5, 64, 10}
	assert.Equal(t, p, p.Above().Below())
	assert.Equal(t, p, p.Below().Above())
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 16, ChunkSize)
	assert.Equal(t, 256, ChunkHeight)
	assert.Equal(t, 16, SectionHeight)
}
