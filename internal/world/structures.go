package world

import (
	"math/rand"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// Dungeon generation constants.
const (
	dungeonChance = 256 // 1-in-N chance per chunk
	dungeonMinY   = 10
	dungeonMaxY   = 50
	dungeonSmallW = 5
	dungeonLargeW = 7
	dungeonHeight = 4
)

// DungeonGenerator places dungeon rooms inside cave systems.
type DungeonGenerator struct{}

// lootEntry describes a single possible item in dungeon chest loot.
type lootEntry struct {
	ItemID    item.ItemID
	MinCount  int
	MaxCount  int
	WeightPct int // cumulative weight boundary (0-100)
}

// LootTable generates randomised loot stacks for dungeon chests.
type LootTable struct {
	entries []lootEntry
}

// NewDungeonLootTable returns a LootTable configured with standard dungeon loot.
//
// Probabilities (cumulative):
//
//	iron ingot  1-3  30%  (0-29)
//	gold ingot  1-3  15%  (30-44)
//	bread       1    25%  (45-69)
//	wheat seeds 2-4  20%  (70-89)
//	string      1-4  10%  (90-99)
func NewDungeonLootTable() LootTable {
	return LootTable{
		entries: []lootEntry{
			{ItemID: item.IronIngot, MinCount: 1, MaxCount: 3, WeightPct: 30},
			{ItemID: item.GoldIngot, MinCount: 1, MaxCount: 3, WeightPct: 45},
			{ItemID: item.Bread, MinCount: 1, MaxCount: 1, WeightPct: 70},
			{ItemID: item.WheatSeeds, MinCount: 2, MaxCount: 4, WeightPct: 90},
			{ItemID: item.StringItem, MinCount: 1, MaxCount: 4, WeightPct: 100},
		},
	}
}

// GenerateLoot returns a random set of 1-4 item stacks drawn from the table.
func (lt LootTable) GenerateLoot(rng *rand.Rand) []item.ItemStack {
	count := 1 + rng.Intn(4) // 1-4 stacks
	stacks := make([]item.ItemStack, 0, count)
	for i := 0; i < count; i++ {
		roll := rng.Intn(100)
		for _, e := range lt.entries {
			if roll < e.WeightPct {
				qty := e.MinCount
				if e.MaxCount > e.MinCount {
					qty += rng.Intn(e.MaxCount - e.MinCount + 1)
				}
				stacks = append(stacks, item.NewItemStack(e.ItemID, qty))
				break
			}
		}
	}
	return stacks
}

// TryPlaceDungeon attempts to place a dungeon in the given chunk.
// It returns true if a dungeon was placed.
func (dg *DungeonGenerator) TryPlaceDungeon(c *chunk.Chunk, rng *rand.Rand) bool {
	if rng.Intn(dungeonChance) != 0 {
		return false
	}

	// Pick a random size: 5x5 or 7x7.
	width := dungeonSmallW
	if rng.Intn(2) == 0 {
		width = dungeonLargeW
	}

	half := width / 2

	// Ensure the room fits within the chunk.
	maxOffset := mcmath.ChunkSize - width
	if maxOffset < 0 {
		return false
	}
	x0 := rng.Intn(maxOffset + 1)
	z0 := rng.Intn(maxOffset + 1)
	y0 := dungeonMinY + rng.Intn(dungeonMaxY-dungeonMinY+1)

	// The room spans [x0, x0+width) x [y0, y0+dungeonHeight) x [z0, z0+width).
	if y0+dungeonHeight >= mcmath.ChunkHeight {
		return false
	}

	// Verify the center column is in a cave (air) at the floor level.
	cx := x0 + half
	cz := z0 + half
	if c.GetBlock(cx, y0, cz) != block.Air {
		return false
	}

	// Build the room.
	placeRoom(c, x0, y0, z0, width)

	// Place mob spawner at the center, one block above the floor.
	c.SetBlock(cx, y0+1, cz, block.MobSpawner)

	// Place 1-2 chests against walls.
	placeChests(c, x0, y0, z0, width, rng)

	return true
}

// placeRoom builds the dungeon shell: mossy cobblestone floor, cobblestone
// walls and ceiling, air interior.
func placeRoom(c *chunk.Chunk, x0, y0, z0, width int) {
	for dx := 0; dx < width; dx++ {
		for dz := 0; dz < width; dz++ {
			lx := x0 + dx
			lz := z0 + dz

			isEdge := dx == 0 || dx == width-1 || dz == 0 || dz == width-1

			for dy := 0; dy < dungeonHeight; dy++ {
				ly := y0 + dy

				switch {
				case dy == 0:
					c.SetBlock(lx, ly, lz, block.MossyCobblestone)
				case dy == dungeonHeight-1:
					c.SetBlock(lx, ly, lz, block.Cobblestone)
				case isEdge:
					c.SetBlock(lx, ly, lz, block.Cobblestone)
				default:
					c.SetBlock(lx, ly, lz, block.Air)
				}
			}
		}
	}
}

// placeChests places 1-2 chests along the inner wall perimeter.
func placeChests(c *chunk.Chunk, x0, y0, z0, width int, rng *rand.Rand) {
	chestCount := 1 + rng.Intn(2) // 1-2

	// Candidate positions: blocks adjacent to a wall on the interior, at floor+1.
	type pos struct{ x, z int }
	var candidates []pos

	for dx := 1; dx < width-1; dx++ {
		for dz := 1; dz < width-1; dz++ {
			// Must be next to at least one wall.
			if dx == 1 || dx == width-2 || dz == 1 || dz == width-2 {
				candidates = append(candidates, pos{x0 + dx, z0 + dz})
			}
		}
	}

	if len(candidates) == 0 {
		return
	}

	// Shuffle and pick.
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for i := 0; i < chestCount && i < len(candidates); i++ {
		p := candidates[i]
		c.SetBlock(p.x, y0+1, p.z, block.Chest)
	}
}
