package world

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

func TestNewStorageCreatesDirectoryStructure(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "myworld")

	s, err := NewStorage(savePath)
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// Verify chunks subdirectory was created.
	info, err := os.Stat(filepath.Join(savePath, "chunks"))
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestSaveLoadChunkRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	// Create a chunk with some known blocks.
	pos := mcmath.ChunkPos{X: 3, Z: -7}
	c := &chunk.Chunk{Pos: pos}
	c.SetBlock(0, 0, 0, block.Bedrock)
	c.SetBlock(5, 64, 5, block.Stone)
	c.SetBlock(15, 255, 15, block.Dirt)

	// Save and reload.
	err = s.SaveChunk(pos, c)
	assert.NoError(t, err)

	loaded, err := s.LoadChunk(pos)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)

	// Verify blocks match.
	assert.Equal(t, block.Bedrock, loaded.GetBlock(0, 0, 0))
	assert.Equal(t, block.Stone, loaded.GetBlock(5, 64, 5))
	assert.Equal(t, block.Dirt, loaded.GetBlock(15, 255, 15))
	assert.Equal(t, pos, loaded.Pos)
}

func TestHasChunkTrueAfterSave(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	pos := mcmath.ChunkPos{X: 1, Z: 2}
	assert.False(t, s.HasChunk(pos))

	c := &chunk.Chunk{Pos: pos}
	c.SetBlock(0, 0, 0, block.Stone)
	err = s.SaveChunk(pos, c)
	assert.NoError(t, err)

	assert.True(t, s.HasChunk(pos))
}

func TestHasChunkFalseForMissing(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	pos := mcmath.ChunkPos{X: 99, Z: -99}
	assert.False(t, s.HasChunk(pos))
}

func TestLoadChunkMissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	_, err = s.LoadChunk(mcmath.ChunkPos{X: 0, Z: 0})
	assert.Error(t, err)
}

func TestSaveLoadLevelRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	level := LevelData{
		Seed:       123456789,
		SpawnX:     8,
		SpawnY:     64,
		SpawnZ:     8,
		GameTime:   12000,
		Difficulty: "normal",
	}

	err = s.SaveLevel(level)
	assert.NoError(t, err)

	loaded, err := s.LoadLevel()
	assert.NoError(t, err)
	assert.Equal(t, level, loaded)
}

func TestLoadLevelMissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	_, err = s.LoadLevel()
	assert.Error(t, err)
}

func TestSaveLoadPlayerRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	pd := PlayerData{
		X:      10.5,
		Y:      64.0,
		Z:      -20.3,
		Yaw:    1.57,
		Pitch:  -0.5,
		Health: 20.0,
		Hunger: 18.0,
		InventorySlots: []InventorySlotData{
			{ItemID: 1, Count: 32, Durability: 0},
			{ItemID: 5, Count: 1, Durability: 100},
		},
	}

	err = s.SavePlayer(pd)
	assert.NoError(t, err)

	loaded, err := s.LoadPlayer()
	assert.NoError(t, err)
	assert.Equal(t, pd, loaded)
}

func TestLoadPlayerMissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	_, err = s.LoadPlayer()
	assert.Error(t, err)
}

func TestListChunks(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	// Initially empty.
	positions, err := s.ListChunks()
	assert.NoError(t, err)
	assert.Empty(t, positions)

	// Save two chunks.
	c1 := &chunk.Chunk{Pos: mcmath.ChunkPos{X: 0, Z: 0}}
	c1.SetBlock(0, 0, 0, block.Stone)
	assert.NoError(t, s.SaveChunk(c1.Pos, c1))

	c2 := &chunk.Chunk{Pos: mcmath.ChunkPos{X: -1, Z: 5}}
	c2.SetBlock(0, 0, 0, block.Dirt)
	assert.NoError(t, s.SaveChunk(c2.Pos, c2))

	positions, err = s.ListChunks()
	assert.NoError(t, err)
	assert.Len(t, positions, 2)

	// Check both positions appear (order may vary).
	posSet := map[[2]int32]bool{}
	for _, p := range positions {
		posSet[p.Key()] = true
	}
	assert.True(t, posSet[[2]int32{0, 0}])
	assert.True(t, posSet[[2]int32{-1, 5}])
}

func TestHasSave(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	assert.False(t, s.HasSave())

	err = s.SaveLevel(LevelData{Seed: 42})
	assert.NoError(t, err)

	assert.True(t, s.HasSave())
}

func TestWorldSaveAllLoadAll(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	// Create a world, load some chunks, modify a block.
	w := NewWorld(testSeed)
	pos1 := mcmath.ChunkPos{X: 0, Z: 0}
	pos2 := mcmath.ChunkPos{X: 1, Z: 0}
	w.LoadChunk(pos1)
	w.LoadChunk(pos2)

	// Modify a block so we can verify it round-trips.
	testBlock := mcmath.BlockPos{X: 5, Y: 100, Z: 5}
	w.SetBlock(testBlock, block.GoldOre)

	// Save all chunks.
	err = w.SaveAll(s)
	assert.NoError(t, err)

	// Create a new empty world and load from storage.
	w2 := NewWorld(testSeed)
	err = w2.LoadAll(s)
	assert.NoError(t, err)

	assert.Equal(t, 2, w2.LoadedChunkCount())
	assert.Equal(t, block.GoldOre, w2.GetBlock(testBlock))
}

func TestSaveChunkNegativeCoordinates(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	assert.NoError(t, err)

	pos := mcmath.ChunkPos{X: -10, Z: -20}
	c := &chunk.Chunk{Pos: pos}
	c.SetBlock(0, 0, 0, block.Stone)

	err = s.SaveChunk(pos, c)
	assert.NoError(t, err)

	assert.True(t, s.HasChunk(pos))

	loaded, err := s.LoadChunk(pos)
	assert.NoError(t, err)
	assert.Equal(t, block.Stone, loaded.GetBlock(0, 0, 0))
}
