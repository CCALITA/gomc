package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// LevelData holds world-level metadata persisted to level.json.
type LevelData struct {
	Seed       int64  `json:"seed"`
	SpawnX     int32  `json:"spawn_x"`
	SpawnY     int32  `json:"spawn_y"`
	SpawnZ     int32  `json:"spawn_z"`
	GameTime   int64  `json:"game_time"`
	Difficulty string `json:"difficulty"`
}

// InventorySlotData represents a single inventory slot in saved player data.
type InventorySlotData struct {
	ItemID     uint16 `json:"item_id"`
	Count      int    `json:"count"`
	Durability int    `json:"durability"`
}

// PlayerData holds player state persisted to player.json.
type PlayerData struct {
	X              float32             `json:"x"`
	Y              float32             `json:"y"`
	Z              float32             `json:"z"`
	Yaw            float32             `json:"yaw"`
	Pitch          float32             `json:"pitch"`
	Health         float32             `json:"health"`
	Hunger         float32             `json:"hunger"`
	InventorySlots []InventorySlotData `json:"inventory_slots"`
}

// Storage manages file-based world persistence using a directory structure:
//
//	savePath/
//	  level.json
//	  player.json
//	  chunks/
//	    X_Z.bin
type Storage struct {
	savePath  string
	chunksDir string
}

// NewStorage creates a Storage rooted at savePath, ensuring the directory
// structure exists. Returns an error if directory creation fails.
func NewStorage(savePath string) (*Storage, error) {
	chunksDir := filepath.Join(savePath, "chunks")
	if err := os.MkdirAll(chunksDir, 0o755); err != nil {
		return nil, fmt.Errorf("NewStorage: create directory: %w", err)
	}
	return &Storage{
		savePath:  savePath,
		chunksDir: chunksDir,
	}, nil
}

// chunkFileName returns the file path for a chunk at the given position.
func (s *Storage) chunkFileName(pos mcmath.ChunkPos) string {
	return filepath.Join(s.chunksDir, fmt.Sprintf("%d_%d.bin", pos.X, pos.Z))
}

// SaveChunk serializes and writes a chunk to disk.
func (s *Storage) SaveChunk(pos mcmath.ChunkPos, c *chunk.Chunk) error {
	data, err := chunk.SerializeChunk(c)
	if err != nil {
		return fmt.Errorf("SaveChunk: serialize chunk %v: %w", pos, err)
	}
	path := s.chunkFileName(pos)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("SaveChunk: write file %s: %w", path, err)
	}
	return nil
}

// LoadChunk reads and deserializes a chunk from disk.
func (s *Storage) LoadChunk(pos mcmath.ChunkPos) (*chunk.Chunk, error) {
	path := s.chunkFileName(pos)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("LoadChunk: read file %s: %w", path, err)
	}
	c, err := chunk.DeserializeChunk(data)
	if err != nil {
		return nil, fmt.Errorf("LoadChunk: deserialize chunk %v: %w", pos, err)
	}
	return c, nil
}

// HasChunk reports whether a saved chunk file exists for the given position.
func (s *Storage) HasChunk(pos mcmath.ChunkPos) bool {
	_, err := os.Stat(s.chunkFileName(pos))
	return err == nil
}

// SaveLevel writes world metadata to level.json.
func (s *Storage) SaveLevel(level LevelData) error {
	data, err := json.MarshalIndent(level, "", "  ")
	if err != nil {
		return fmt.Errorf("SaveLevel: marshal level data: %w", err)
	}
	path := filepath.Join(s.savePath, "level.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("SaveLevel: write file: %w", err)
	}
	return nil
}

// LoadLevel reads world metadata from level.json.
func (s *Storage) LoadLevel() (LevelData, error) {
	path := filepath.Join(s.savePath, "level.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return LevelData{}, fmt.Errorf("LoadLevel: read file: %w", err)
	}
	var level LevelData
	if err := json.Unmarshal(data, &level); err != nil {
		return LevelData{}, fmt.Errorf("LoadLevel: unmarshal data: %w", err)
	}
	return level, nil
}

// SavePlayer writes player state to player.json.
func (s *Storage) SavePlayer(player PlayerData) error {
	data, err := json.MarshalIndent(player, "", "  ")
	if err != nil {
		return fmt.Errorf("SavePlayer: marshal player data: %w", err)
	}
	path := filepath.Join(s.savePath, "player.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("SavePlayer: write file: %w", err)
	}
	return nil
}

// LoadPlayer reads player state from player.json.
func (s *Storage) LoadPlayer() (PlayerData, error) {
	path := filepath.Join(s.savePath, "player.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return PlayerData{}, fmt.Errorf("LoadPlayer: read file: %w", err)
	}
	var player PlayerData
	if err := json.Unmarshal(data, &player); err != nil {
		return PlayerData{}, fmt.Errorf("LoadPlayer: unmarshal data: %w", err)
	}
	return player, nil
}

// ListChunks returns the positions of all chunk files in the save directory.
// Files that do not match the expected naming pattern are skipped, and a
// combined error describing all skipped entries is returned alongside any
// successfully parsed positions.
func (s *Storage) ListChunks() ([]mcmath.ChunkPos, error) {
	entries, err := os.ReadDir(s.chunksDir)
	if err != nil {
		return nil, fmt.Errorf("ListChunks: read directory: %w", err)
	}
	var positions []mcmath.ChunkPos
	var parseErrs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var x, z int32
		name := entry.Name()
		if _, err := fmt.Sscanf(name, "%d_%d.bin", &x, &z); err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("ListChunks: parse filename %q: %w", name, err))
			continue
		}
		positions = append(positions, mcmath.ChunkPos{X: x, Z: z})
	}
	return positions, errors.Join(parseErrs...)
}

// HasSave reports whether a save directory with a level.json exists.
func (s *Storage) HasSave() bool {
	path := filepath.Join(s.savePath, "level.json")
	_, err := os.Stat(path)
	return err == nil
}
