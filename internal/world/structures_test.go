package world

import (
	"math/rand"
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildCaveChunk creates a chunk with stone everywhere except a hollow cave
// zone, suitable for dungeon placement.
func buildCaveChunk() *chunk.Chunk {
	c := &chunk.Chunk{Pos: mcmath.ChunkPos{X: 0, Z: 0}}
	for x := 0; x < mcmath.ChunkSize; x++ {
		for z := 0; z < mcmath.ChunkSize; z++ {
			for y := 0; y < mcmath.ChunkHeight; y++ {
				if y >= dungeonMinY && y <= dungeonMaxY+dungeonHeight {
					c.SetBlock(x, y, z, block.Air)
				} else {
					c.SetBlock(x, y, z, block.Stone)
				}
			}
		}
	}
	return c
}

// countBlock returns how many times blockID appears in the chunk.
func countBlock(c *chunk.Chunk, blockID block.BlockID) int {
	n := 0
	for x := 0; x < mcmath.ChunkSize; x++ {
		for z := 0; z < mcmath.ChunkSize; z++ {
			for y := 0; y < mcmath.ChunkHeight; y++ {
				if c.GetBlock(x, y, z) == blockID {
					n++
				}
			}
		}
	}
	return n
}

// forceDungeonPlacement tries many seeds until a dungeon is placed.
func forceDungeonPlacement(t *testing.T) (*chunk.Chunk, bool) {
	t.Helper()
	dg := DungeonGenerator{}

	for seed := int64(0); seed < 10000; seed++ {
		c := buildCaveChunk()
		rng := rand.New(rand.NewSource(seed))
		if dg.TryPlaceDungeon(c, rng) {
			return c, true
		}
	}
	return nil, false
}

func TestDungeonPlacementWithinBounds(t *testing.T) {
	c, ok := forceDungeonPlacement(t)
	require.True(t, ok, "expected at least one dungeon placement in 10000 attempts")

	for x := 0; x < mcmath.ChunkSize; x++ {
		for z := 0; z < mcmath.ChunkSize; z++ {
			for y := 0; y < mcmath.ChunkHeight; y++ {
				bid := c.GetBlock(x, y, z)
				if bid == block.MossyCobblestone || bid == block.Cobblestone ||
					bid == block.MobSpawner || bid == block.Chest {
					assert.True(t, x >= 0 && x < mcmath.ChunkSize, "x out of bounds: %d", x)
					assert.True(t, z >= 0 && z < mcmath.ChunkSize, "z out of bounds: %d", z)
					assert.True(t, y >= dungeonMinY && y < mcmath.ChunkHeight, "y out of bounds: %d", y)
				}
			}
		}
	}
}

func TestDungeonHasMossyCobblestoneFloor(t *testing.T) {
	c, ok := forceDungeonPlacement(t)
	require.True(t, ok)

	mossyCount := countBlock(c, block.MossyCobblestone)
	assert.True(t, mossyCount == 25 || mossyCount == 49,
		"expected 25 or 49 mossy cobblestone floor tiles, got %d", mossyCount)
}

func TestDungeonHasSpawnerAtCenter(t *testing.T) {
	c, ok := forceDungeonPlacement(t)
	require.True(t, ok)

	assert.Equal(t, 1, countBlock(c, block.MobSpawner), "expected exactly 1 mob spawner")
}

func TestDungeonHasChests(t *testing.T) {
	c, ok := forceDungeonPlacement(t)
	require.True(t, ok)

	chestCount := countBlock(c, block.Chest)
	assert.True(t, chestCount >= 1 && chestCount <= 2,
		"expected 1-2 chests, got %d", chestCount)
}

func TestDungeonSpawnerIsAtRoomCenter(t *testing.T) {
	c, ok := forceDungeonPlacement(t)
	require.True(t, ok)

	var sx, sy, sz int
	found := false
	for x := 0; x < mcmath.ChunkSize; x++ {
		for z := 0; z < mcmath.ChunkSize; z++ {
			for y := 0; y < mcmath.ChunkHeight; y++ {
				if c.GetBlock(x, y, z) == block.MobSpawner {
					sx, sy, sz = x, y, z
					found = true
				}
			}
		}
	}
	require.True(t, found)

	assert.Equal(t, block.MossyCobblestone, c.GetBlock(sx, sy-1, sz),
		"floor below spawner should be mossy cobblestone")
}

func TestLootTableGeneratesValidItems(t *testing.T) {
	lt := NewDungeonLootTable()
	rng := rand.New(rand.NewSource(123))

	validItems := map[item.ItemID]bool{
		item.IronIngot:  true,
		item.GoldIngot:  true,
		item.Bread:      true,
		item.WheatSeeds: true,
		item.StringItem: true,
	}

	for i := 0; i < 100; i++ {
		stacks := lt.GenerateLoot(rng)
		assert.True(t, len(stacks) >= 1 && len(stacks) <= 4,
			"expected 1-4 stacks, got %d", len(stacks))
		for _, s := range stacks {
			assert.True(t, validItems[s.ItemID],
				"unexpected item ID in loot: %d", s.ItemID)
			assert.True(t, s.Count >= 1, "item count should be >= 1")
		}
	}
}

func TestLootTableItemCounts(t *testing.T) {
	lt := NewDungeonLootTable()
	rng := rand.New(rand.NewSource(456))

	for i := 0; i < 200; i++ {
		stacks := lt.GenerateLoot(rng)
		for _, s := range stacks {
			switch s.ItemID {
			case item.IronIngot:
				assert.True(t, s.Count >= 1 && s.Count <= 3)
			case item.GoldIngot:
				assert.True(t, s.Count >= 1 && s.Count <= 3)
			case item.Bread:
				assert.Equal(t, 1, s.Count)
			case item.WheatSeeds:
				assert.True(t, s.Count >= 2 && s.Count <= 4)
			case item.StringItem:
				assert.True(t, s.Count >= 1 && s.Count <= 4)
			}
		}
	}
}

func TestDungeonNotPlacedWithoutCave(t *testing.T) {
	dg := DungeonGenerator{}
	pos := mcmath.ChunkPos{X: 0, Z: 0}
	placed := false
	for seed := int64(0); seed < 10000; seed++ {
		c := &chunk.Chunk{Pos: pos}
		for x := 0; x < mcmath.ChunkSize; x++ {
			for z := 0; z < mcmath.ChunkSize; z++ {
				for y := 0; y < mcmath.ChunkHeight; y++ {
					c.SetBlock(x, y, z, block.Stone)
				}
			}
		}
		rng := rand.New(rand.NewSource(seed))
		if dg.TryPlaceDungeon(c, rng) {
			placed = true
			break
		}
	}
	assert.False(t, placed, "dungeon should not be placed in solid stone")
}

func TestNewBlockProperties(t *testing.T) {
	mossy := block.GetProperties(block.MossyCobblestone)
	assert.Equal(t, "mossy_cobblestone", mossy.Name)
	assert.True(t, mossy.Solid)
	assert.False(t, mossy.Transparent)

	spawner := block.GetProperties(block.MobSpawner)
	assert.Equal(t, "mob_spawner", spawner.Name)
	assert.Equal(t, float32(5.0), spawner.Hardness)
	assert.True(t, spawner.Solid)
	assert.False(t, spawner.Transparent)
}
