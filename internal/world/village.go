package world

import (
	"math/rand"

	"github.com/fanxiyao/gomc/internal/biome"
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

const (
	// villageCellSize is the side length (in chunks) of each grid cell used to
	// determine village placement. At most one village spawns per cell.
	villageCellSize = 32

	// villageMinBuildings is the minimum number of buildings per village.
	villageMinBuildings = 3

	// villageMaxBuildings is the maximum number of buildings per village.
	villageMaxBuildings = 7

	// buildingSpacing is the gap in blocks between adjacent buildings on the grid.
	buildingSpacing = 3
)

// buildingType identifies the kind of structure to generate.
type buildingType int

const (
	buildingSmallHouse buildingType = iota
	buildingWell
	buildingFarm
)

// buildingFootprint stores the width (X) and depth (Z) of a building type.
type buildingFootprint struct {
	w, d int
}

// footprints maps each building type to its X/Z footprint.
var footprints = map[buildingType]buildingFootprint{
	buildingSmallHouse: {w: 5, d: 5},
	buildingWell:       {w: 3, d: 3},
	buildingFarm:       {w: 7, d: 3},
}

// VillageGenerator determines whether a village should appear in a chunk and
// produces the corresponding block data.
type VillageGenerator struct {
	Seed int64
}

// ShouldGenerateVillage reports whether a village center should be placed at
// the given chunk position. Villages appear only in Plains (ID 0) or Desert
// biomes and are limited to one per 32x32 chunk grid cell, with the exact
// chunk within the cell selected deterministically from the seed.
func (vg *VillageGenerator) ShouldGenerateVillage(chunkPos mcmath.ChunkPos, biomeID uint8) bool {
	if biomeID != uint8(biome.Plains) && biomeID != uint8(biome.Desert) {
		return false
	}

	cellX, cellZ := cellCoords(chunkPos)
	rng := rand.New(rand.NewSource(vg.Seed ^ int64(cellX)*6364136223846793005 ^ int64(cellZ)*1442695040888963407))
	targetX := cellX*villageCellSize + rng.Intn(villageCellSize)
	targetZ := cellZ*villageCellSize + rng.Intn(villageCellSize)

	return int(chunkPos.X) == targetX && int(chunkPos.Z) == targetZ
}

// GenerateVillage places a village in the given chunk. It writes buildings
// and gravel paths directly into the chunk. The biomeID selects material
// variants (sandstone in deserts, oak planks otherwise).
func (vg *VillageGenerator) GenerateVillage(c *chunk.Chunk, biomeID uint8, rng *rand.Rand) {
	count := villageMinBuildings + rng.Intn(villageMaxBuildings-villageMinBuildings+1)
	plan := pickBuildings(count, rng)
	placements := layoutBuildings(plan)

	wallBlock := block.OakPlanks
	if biomeID == uint8(biome.Desert) {
		wallBlock = block.Sandstone
	}

	for _, p := range placements {
		surfaceY := c.HighestBlock(p.lx, p.lz)
		if surfaceY < 1 || surfaceY >= mcmath.ChunkHeight-10 {
			continue
		}
		switch p.bt {
		case buildingSmallHouse:
			placeSmallHouse(c, p.lx, surfaceY, p.lz, wallBlock)
		case buildingWell:
			placeWell(c, p.lx, surfaceY, p.lz)
		case buildingFarm:
			placeFarm(c, p.lx, surfaceY, p.lz, rng)
		}
	}

	connectPaths(c, placements)
}

// placement records where a building sits within the chunk.
type placement struct {
	bt     buildingType
	lx, lz int // local X/Z origin inside the chunk
}

// cellCoords returns the grid-cell coordinates for a chunk position, using
// floored division so negative coordinates map correctly.
func cellCoords(pos mcmath.ChunkPos) (cellX, cellZ int) {
	cellX = floorDivInt(int(pos.X), villageCellSize)
	cellZ = floorDivInt(int(pos.Z), villageCellSize)
	return
}

// floorDivInt performs floored integer division (rounds toward negative infinity).
func floorDivInt(a, b int) int {
	d := a / b
	if (a^b) < 0 && d*b != a {
		d--
	}
	return d
}

// pickBuildings selects which building types to generate. The first building
// is always a well; the rest are a random mix of houses and farms.
func pickBuildings(count int, rng *rand.Rand) []buildingType {
	types := make([]buildingType, 0, count)
	types = append(types, buildingWell)
	for i := 1; i < count; i++ {
		if rng.Intn(3) == 0 {
			types = append(types, buildingFarm)
		} else {
			types = append(types, buildingSmallHouse)
		}
	}
	return types
}

// layoutBuildings arranges the selected buildings in a compact grid with
// buildingSpacing-block gaps, centered roughly in the chunk.
func layoutBuildings(types []buildingType) []placement {
	placements := make([]placement, 0, len(types))

	// Start near the center of the chunk.
	startX := 2
	curX := startX
	curZ := 2

	rowMaxD := 0

	for _, bt := range types {
		fp := footprints[bt]
		// If this building would overflow the chunk width, start a new row.
		if curX+fp.w > mcmath.ChunkSize-1 {
			curX = startX
			curZ += rowMaxD + buildingSpacing
			rowMaxD = 0
		}
		// If this building would overflow the chunk depth, skip it.
		if curZ+fp.d > mcmath.ChunkSize-1 {
			continue
		}
		placements = append(placements, placement{bt: bt, lx: curX, lz: curZ})
		if fp.d > rowMaxD {
			rowMaxD = fp.d
		}
		curX += fp.w + buildingSpacing
	}
	return placements
}

// placeSmallHouse builds a 5x5x4 house with walls, cobblestone foundation,
// a door on the front face, a torch inside, and a flat plank roof.
func placeSmallHouse(c *chunk.Chunk, ox, baseY, oz int, wallBlock block.BlockID) {
	h := 4 // wall height (including foundation row)

	for dx := 0; dx < 5; dx++ {
		for dz := 0; dz < 5; dz++ {
			lx := ox + dx
			lz := oz + dz
			if lx >= mcmath.ChunkSize || lz >= mcmath.ChunkSize {
				continue
			}
			// Foundation (Y = baseY).
			c.SetBlock(lx, baseY, lz, block.Cobblestone)

			isWall := dx == 0 || dx == 4 || dz == 0 || dz == 4
			for dy := 1; dy < h; dy++ {
				y := baseY + dy
				if y >= mcmath.ChunkHeight {
					break
				}
				if isWall {
					c.SetBlock(lx, y, lz, wallBlock)
				} else {
					c.SetBlock(lx, y, lz, block.Air)
				}
			}

			// Roof at top.
			roofY := baseY + h
			if roofY < mcmath.ChunkHeight {
				c.SetBlock(lx, roofY, lz, block.OakPlanks)
			}
		}
	}

	// Door: cut a 1x2 opening on the front wall (dz == 0), center column (dx == 2).
	doorX := ox + 2
	if doorX < mcmath.ChunkSize {
		for dy := 1; dy <= 2; dy++ {
			y := baseY + dy
			if y < mcmath.ChunkHeight {
				c.SetBlock(doorX, y, oz, block.OakDoor)
			}
		}
	}

	// Torch in the center at floor+2.
	torchX := ox + 2
	torchZ := oz + 2
	torchY := baseY + 2
	if torchX < mcmath.ChunkSize && torchZ < mcmath.ChunkSize && torchY < mcmath.ChunkHeight {
		c.SetBlock(torchX, torchY, torchZ, block.Torch)
	}
}

// placeWell builds a 3x3 cobblestone ring with water inside and oak fence
// posts on the four corners.
func placeWell(c *chunk.Chunk, ox, baseY, oz int) {
	for dx := 0; dx < 3; dx++ {
		for dz := 0; dz < 3; dz++ {
			lx := ox + dx
			lz := oz + dz
			if lx >= mcmath.ChunkSize || lz >= mcmath.ChunkSize {
				continue
			}
			isCorner := (dx == 0 || dx == 2) && (dz == 0 || dz == 2)
			isCenter := dx == 1 && dz == 1

			// Ring wall at surface level.
			c.SetBlock(lx, baseY, lz, block.Cobblestone)

			if isCenter {
				// Water one block below the rim.
				if baseY-1 >= 0 {
					c.SetBlock(lx, baseY-1, lz, block.Water)
				}
			}

			if isCorner {
				// Fence post above the rim.
				if baseY+1 < mcmath.ChunkHeight {
					c.SetBlock(lx, baseY+1, lz, block.OakFence)
				}
			}
		}
	}
}

// placeFarm builds a 7x3 farm: a water channel in the center row with
// farmland and wheat on each side.
func placeFarm(c *chunk.Chunk, ox, baseY, oz int, rng *rand.Rand) {
	for dx := 0; dx < 7; dx++ {
		for dz := 0; dz < 3; dz++ {
			lx := ox + dx
			lz := oz + dz
			if lx >= mcmath.ChunkSize || lz >= mcmath.ChunkSize {
				continue
			}
			if dz == 1 {
				// Center row: water channel.
				c.SetBlock(lx, baseY, lz, block.Water)
			} else {
				// Outer rows: farmland with wheat above.
				c.SetBlock(lx, baseY, lz, block.Farmland)
				if baseY+1 < mcmath.ChunkHeight {
					rng.Intn(8) // advance RNG per crop for stable output across building counts
					c.SetBlock(lx, baseY+1, lz, block.WheatCrop)
				}
			}
		}
	}
}

// connectPaths lays gravel between consecutive buildings at the surface
// level. The path runs from the center of one building to the center of
// the next.
func connectPaths(c *chunk.Chunk, placements []placement) {
	for i := 0; i+1 < len(placements); i++ {
		a := placements[i]
		b := placements[i+1]
		fpA := footprints[a.bt]
		fpB := footprints[b.bt]

		ax := a.lx + fpA.w/2
		az := a.lz + fpA.d/2
		bx := b.lx + fpB.w/2
		bz := b.lz + fpB.d/2

		drawPath(c, ax, az, bx, bz)
	}
}

// drawPath lays a gravel L-shaped path between two local XZ points,
// adapting to the terrain height at each column.
func drawPath(c *chunk.Chunk, x1, z1, x2, z2 int) {
	// Horizontal segment (vary X, keep Z = z1).
	stepX := 1
	if x2 < x1 {
		stepX = -1
	}
	for x := x1; x != x2; x += stepX {
		placeGravel(c, x, z1)
	}
	// Vertical segment (vary Z, keep X = x2).
	stepZ := 1
	if z2 < z1 {
		stepZ = -1
	}
	for z := z1; z != z2+stepZ; z += stepZ {
		placeGravel(c, x2, z)
	}
}

// placeGravel sets a single gravel block at the surface level if the
// coordinates are within chunk bounds.
func placeGravel(c *chunk.Chunk, lx, lz int) {
	if lx < 0 || lx >= mcmath.ChunkSize || lz < 0 || lz >= mcmath.ChunkSize {
		return
	}
	y := c.HighestBlock(lx, lz)
	if y >= 0 && y < mcmath.ChunkHeight {
		c.SetBlock(lx, y, lz, block.Gravel)
	}
}
