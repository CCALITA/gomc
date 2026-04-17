package mcmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkPos_WorldBlockX(t *testing.T) {
	tests := []struct {
		name string
		c    ChunkPos
		want int32
	}{
		{"positive", ChunkPos{2, 0}, 32},
		{"zero", ChunkPos{0, 0}, 0},
		{"negative", ChunkPos{-1, 0}, -16},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.c.WorldBlockX())
		})
	}
}

func TestChunkPos_WorldBlockZ(t *testing.T) {
	tests := []struct {
		name string
		c    ChunkPos
		want int32
	}{
		{"positive", ChunkPos{0, 3}, 48},
		{"zero", ChunkPos{0, 0}, 0},
		{"negative", ChunkPos{0, -2}, -32},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.c.WorldBlockZ())
		})
	}
}

func TestChunkPos_BlockPosAt(t *testing.T) {
	c := ChunkPos{1, 2}
	bp := c.BlockPosAt(5, 64, 10)
	assert.Equal(t, BlockPos{21, 64, 42}, bp)
}

func TestChunkPos_BlockPosAt_Negative(t *testing.T) {
	c := ChunkPos{-1, -1}
	bp := c.BlockPosAt(0, 0, 0)
	assert.Equal(t, BlockPos{-16, 0, -16}, bp)
}

func TestChunkPos_Distance(t *testing.T) {
	a := ChunkPos{0, 0}
	b := ChunkPos{3, 4}
	assert.InDelta(t, 5.0, a.Distance(b), 1e-10)
}

func TestChunkPos_Distance_Same(t *testing.T) {
	a := ChunkPos{5, 5}
	assert.InDelta(t, 0.0, a.Distance(a), 1e-10)
}

func TestChunkPos_Key(t *testing.T) {
	c := ChunkPos{10, 20}
	key := c.Key()
	assert.Equal(t, [2]int32{10, 20}, key)

	// Verify usable as map key
	m := map[[2]int32]bool{}
	m[key] = true
	assert.True(t, m[c.Key()])
}

func TestChunkPos_String(t *testing.T) {
	c := ChunkPos{1, 2}
	s := c.String()
	assert.Contains(t, s, "ChunkPos")
	assert.Contains(t, s, "1")
	assert.Contains(t, s, "2")
}

func TestChunkPos_RoundTrip(t *testing.T) {
	// BlockPos -> ChunkPos -> BlockPosAt with local coords should give back original
	bp := BlockPos{21, 64, 42}
	cp := bp.ToChunkPos()
	local := bp.LocalPos()
	recovered := cp.BlockPosAt(local.X, bp.Y, local.Z)
	assert.Equal(t, bp, recovered)
}

func TestChunkPos_RoundTrip_Negative(t *testing.T) {
	bp := BlockPos{-5, 10, -20}
	cp := bp.ToChunkPos()
	local := bp.LocalPos()
	recovered := cp.BlockPosAt(local.X, bp.Y, local.Z)
	assert.Equal(t, bp, recovered)
}
