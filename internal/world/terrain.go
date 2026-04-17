// Package world implements world management, terrain generation, and async
// chunk loading for a Minecraft-like voxel game.
package world

import (
	"math/rand"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/noise"
)

const (
	seaLevel       = 62
	baseHeight     = 64
	heightAmp      = 32
	heightScale    = 0.01
	caveThreshold  = 0.6
	caveMinY       = 5
	bedrockRandMax = 4
)

// oreConfig describes how a single ore type is generated.
type oreConfig struct {
	blockID   block.BlockID
	maxY      int
	threshold float64
	scale     float64
}

// TerrainGenerator produces chunks with realistic terrain using layered noise.
type TerrainGenerator struct {
	seed      int64
	heightMap *noise.HeightMap
	caveNoise *noise.OctaveNoise
	treeNoise *noise.OctaveNoise
	ores      []oreConfig
	oreNoises []*noise.OctaveNoise
}

// NewTerrainGenerator creates a TerrainGenerator with the given seed.
// It initialises noise generators for height, caves, trees, and ores.
func NewTerrainGenerator(seed int64) *TerrainGenerator {
	heightGen := noise.NewNoiseGenerator(seed)
	octave := noise.NewOctaveNoise(heightGen, 6, 0.5, 2.0)
	hm := noise.NewHeightMap(octave, baseHeight, heightAmp)

	caveGen := noise.NewNoiseGenerator(seed + 1)
	caveOctave := noise.NewOctaveNoise(caveGen, 4, 0.5, 2.0)

	treeGen := noise.NewNoiseGenerator(seed + 2)
	treeOctave := noise.NewOctaveNoise(treeGen, 2, 0.5, 2.0)

	ores := []oreConfig{
		{blockID: block.CoalOre, maxY: 80, threshold: 0.55, scale: 0.08},
		{blockID: block.IronOre, maxY: 64, threshold: 0.6, scale: 0.07},
		{blockID: block.GoldOre, maxY: 32, threshold: 0.7, scale: 0.06},
		{blockID: block.DiamondOre, maxY: 16, threshold: 0.8, scale: 0.05},
	}

	oreNoises := make([]*noise.OctaveNoise, len(ores))
	for i := range ores {
		gen := noise.NewNoiseGenerator(seed + 100 + int64(i))
		oreNoises[i] = noise.NewOctaveNoise(gen, 3, 0.5, 2.0)
	}

	return &TerrainGenerator{
		seed:      seed,
		heightMap: hm,
		caveNoise: caveOctave,
		treeNoise: treeOctave,
		ores:      ores,
		oreNoises: oreNoises,
	}
}

// GenerateChunk generates a full terrain chunk at the given chunk position.
func (tg *TerrainGenerator) GenerateChunk(pos mcmath.ChunkPos) *chunk.Chunk {
	c := &chunk.Chunk{Pos: pos}
	rng := rand.New(rand.NewSource(tg.seed ^ int64(pos.X)*397 ^ int64(pos.Z)*7901))

	// Pre-compute height map for the 16x16 column.
	var heights [mcmath.ChunkSize][mcmath.ChunkSize]int
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			wx := int(pos.WorldBlockX()) + lx
			wz := int(pos.WorldBlockZ()) + lz
			heights[lx][lz] = tg.heightMap.HeightAt(wx, wz)
		}
	}

	// Pass 1: base terrain, bedrock, dirt/grass/stone layers.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			h := heights[lx][lz]
			for y := 0; y <= h && y < mcmath.ChunkHeight; y++ {
				var id block.BlockID
				switch {
				case y == 0:
					id = block.Bedrock
				case y <= bedrockRandMax && rng.Intn(y+1) == 0:
					id = block.Bedrock
				case y == h:
					id = block.Grass
				case y >= h-3:
					id = block.Dirt
				default:
					id = block.Stone
				}
				c.SetBlock(lx, y, lz, id)
			}
		}
	}

	// Pass 2: water at sea level and sand at shorelines.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			h := heights[lx][lz]
			// Fill air below sea level with water.
			for y := h + 1; y <= seaLevel; y++ {
				if y >= 0 && y < mcmath.ChunkHeight {
					c.SetBlock(lx, y, lz, block.Water)
				}
			}
			// Sand at shorelines: if surface is at or just below sea level.
			if h <= seaLevel && h >= seaLevel-2 {
				for y := h; y >= h-3 && y >= 0; y-- {
					bid := c.GetBlock(lx, y, lz)
					if bid == block.Grass || bid == block.Dirt {
						c.SetBlock(lx, y, lz, block.Sand)
					}
				}
			}
		}
	}

	// Pass 3: caves — carve air using 3D noise.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			h := heights[lx][lz]
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)
			for y := caveMinY; y < h && y < mcmath.ChunkHeight; y++ {
				n := tg.caveNoise.Sample3D(wx*0.05, float64(y)*0.05, wz*0.05)
				if n > caveThreshold {
					c.SetBlock(lx, y, lz, block.Air)
				}
			}
		}
	}

	// Pass 4: ores — place ore blocks in stone using 3D noise with per-ore seeds.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)
			for i, ore := range tg.ores {
				for y := 1; y < ore.maxY && y < mcmath.ChunkHeight; y++ {
					if c.GetBlock(lx, y, lz) != block.Stone {
						continue
					}
					n := tg.oreNoises[i].Sample3D(wx*ore.scale, float64(y)*ore.scale, wz*ore.scale)
					if n > ore.threshold {
						c.SetBlock(lx, y, lz, ore.blockID)
					}
				}
			}
		}
	}

	// Pass 5: trees — scatter on grass blocks using 2D noise.
	tg.generateTrees(c, pos, heights, rng)

	return c
}

// generateTrees attempts to place oak trees on suitable grass blocks.
func (tg *TerrainGenerator) generateTrees(c *chunk.Chunk, pos mcmath.ChunkPos, heights [mcmath.ChunkSize][mcmath.ChunkSize]int, rng *rand.Rand) {
	for lx := 2; lx < mcmath.ChunkSize-2; lx++ {
		for lz := 2; lz < mcmath.ChunkSize-2; lz++ {
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)
			n := tg.treeNoise.Sample2D(wx*0.1, wz*0.1)
			if n < 0.7 {
				continue
			}
			h := heights[lx][lz]
			if h <= seaLevel || h >= mcmath.ChunkHeight-10 {
				continue
			}
			if c.GetBlock(lx, h, lz) != block.Grass {
				continue
			}
			tg.placeTree(c, lx, h+1, lz, rng)
		}
	}
}

// placeTree places an oak tree template at the given local position.
// The trunk is 5-7 blocks tall; leaves form a 5x5x3 canopy at the top.
func (tg *TerrainGenerator) placeTree(c *chunk.Chunk, lx, baseY, lz int, rng *rand.Rand) {
	trunkHeight := 5 + rng.Intn(3) // 5, 6, or 7
	topY := baseY + trunkHeight - 1

	if topY+3 >= mcmath.ChunkHeight {
		return
	}

	// Place trunk.
	for y := baseY; y <= topY; y++ {
		c.SetBlock(lx, y, lz, block.OakLog)
	}

	// Place leaf canopy: 5x5x3 centred on the trunk, starting 2 blocks below the top.
	leafStartY := topY - 1
	for dy := 0; dy < 3; dy++ {
		for dx := -2; dx <= 2; dx++ {
			for dz := -2; dz <= 2; dz++ {
				nx := lx + dx
				nz := lz + dz
				ny := leafStartY + dy
				if nx < 0 || nx >= mcmath.ChunkSize || nz < 0 || nz >= mcmath.ChunkSize {
					continue
				}
				if ny >= mcmath.ChunkHeight {
					continue
				}
				// Skip the corners on the top layer for a rounded look.
				if dy == 2 && abs(dx) == 2 && abs(dz) == 2 {
					continue
				}
				// Do not overwrite trunk.
				if c.GetBlock(nx, ny, nz) == block.OakLog {
					continue
				}
				c.SetBlock(nx, ny, nz, block.OakLeaves)
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
