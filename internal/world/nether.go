package world

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/noise"
)

const (
	netherHeight       = 128
	netherLavaLevel    = 31
	netherCaveScale    = 0.05
	netherCaveThresh   = 0.3
	netherCeilingY     = 127
	netherGlowMinY     = 120
	netherGlowScale    = 0.1
	netherGlowThresh   = 0.6
	netherSoulMinY     = 32
	netherSoulMaxY     = 40
	netherSoulScale    = 0.08
	netherSoulThresh   = 0.5
	netherQuartzMaxY   = 120
	netherQuartzScale  = 0.06
	netherQuartzThresh = 0.55
)

// NetherTerrainGenerator produces Nether-style chunks.
type NetherTerrainGenerator struct {
	seed        int64
	caveNoise   *noise.OctaveNoise
	glowNoise   *noise.OctaveNoise
	soulNoise   *noise.OctaveNoise
	quartzNoise *noise.OctaveNoise
}

// NewNetherTerrainGenerator creates a NetherTerrainGenerator with the given seed.
func NewNetherTerrainGenerator(seed int64) *NetherTerrainGenerator {
	return &NetherTerrainGenerator{
		seed:        seed,
		caveNoise:   noise.NewOctaveNoise(noise.NewNoiseGenerator(seed+200), 4, 0.5, 2.0),
		glowNoise:   noise.NewOctaveNoise(noise.NewNoiseGenerator(seed+201), 2, 0.5, 2.0),
		soulNoise:   noise.NewOctaveNoise(noise.NewNoiseGenerator(seed+202), 2, 0.5, 2.0),
		quartzNoise: noise.NewOctaveNoise(noise.NewNoiseGenerator(seed+203), 3, 0.5, 2.0),
	}
}

// GenerateNetherChunk produces a Nether chunk at the given position.
func (ng *NetherTerrainGenerator) GenerateNetherChunk(pos mcmath.ChunkPos) *chunk.Chunk {
	c := &chunk.Chunk{Pos: pos}

	// Pass 1: Fill Y 0-127 with netherrack, bedrock floor/ceiling.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 0; y < netherHeight; y++ {
				if y == 0 || y == netherCeilingY {
					c.SetBlock(lx, y, lz, block.Bedrock)
				} else {
					c.SetBlock(lx, y, lz, block.NetherRack)
				}
			}
		}
	}

	// Pass 2: Carve caves, then decorate (glowstone, soul sand, quartz).
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)

			for y := 1; y < netherCeilingY; y++ {
				n := ng.caveNoise.Sample3D(wx*netherCaveScale, float64(y)*netherCaveScale, wz*netherCaveScale)
				if n > netherCaveThresh {
					c.SetBlock(lx, y, lz, block.Air)
				}
			}
		}
	}

	// Pass 3: Lava ocean at Y <= netherLavaLevel (fill carved air with lava).
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			for y := 1; y <= netherLavaLevel; y++ {
				if c.GetBlock(lx, y, lz) == block.Air {
					c.SetBlock(lx, y, lz, block.Lava)
				}
			}
		}
	}

	// Pass 4: Decorate netherrack — glowstone, soul sand, quartz ore.
	for lx := 0; lx < mcmath.ChunkSize; lx++ {
		for lz := 0; lz < mcmath.ChunkSize; lz++ {
			wx := float64(int(pos.WorldBlockX()) + lx)
			wz := float64(int(pos.WorldBlockZ()) + lz)

			for y := 1; y < netherCeilingY; y++ {
				if c.GetBlock(lx, y, lz) != block.NetherRack {
					continue
				}
				fy := float64(y)

				// Glowstone clusters near ceiling.
				if y >= netherGlowMinY {
					if ng.glowNoise.Sample3D(wx*netherGlowScale, fy*netherGlowScale, wz*netherGlowScale) > netherGlowThresh {
						c.SetBlock(lx, y, lz, block.Glowstone)
						continue
					}
				}

				// Soul sand patches.
				if y >= netherSoulMinY && y <= netherSoulMaxY {
					if ng.soulNoise.Sample3D(wx*netherSoulScale, fy*netherSoulScale, wz*netherSoulScale) > netherSoulThresh {
						c.SetBlock(lx, y, lz, block.SoulSand)
						continue
					}
				}

				// Nether quartz ore.
				if y < netherQuartzMaxY {
					if ng.quartzNoise.Sample3D(wx*netherQuartzScale, fy*netherQuartzScale, wz*netherQuartzScale) > netherQuartzThresh {
						c.SetBlock(lx, y, lz, block.NetherQuartzOre)
					}
				}
			}
		}
	}

	return c
}
