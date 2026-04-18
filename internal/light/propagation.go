// Package light implements BFS flood-fill light propagation for block light
// and sky light in a Minecraft-like voxel world.
package light

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

const (
	maxLight    = 15
	sectionSize = mcmath.ChunkSize // 16
	numSections = mcmath.ChunkHeight / mcmath.SectionHeight
)

// sectionsArray is the fixed-size array of section pointers used by a chunk.
type sectionsArray = [numSections]*chunk.Section

// lightNode represents a position in the BFS queue with its current light level.
type lightNode struct {
	x, y, z int
	level   uint8
}

// neighbor offsets for the six cardinal directions.
var neighborOffsets = [6][3]int{
	{1, 0, 0}, {-1, 0, 0},
	{0, 1, 0}, {0, -1, 0},
	{0, 0, 1}, {0, 0, -1},
}

// lightAccessor bundles the get/set operations for a single light channel
// (block light or sky light) so the BFS can be shared.
type lightAccessor struct {
	get func(sections *sectionsArray, x, y, z int) uint8
	set func(sections *sectionsArray, x, y, z int, level uint8)
}

var blockLightAccessor = lightAccessor{get: getBlockLightAt, set: setBlockLightAt}
var skyLightAccessor = lightAccessor{get: getSkyLightAt, set: setSkyLightAt}

// PropagateBlockLight performs a BFS flood fill from a single light source at
// (x, y, z) with the given starting level. Light decreases by 1 per transparent
// block and is blocked entirely by solid blocks.
//
// Coordinates are in chunk-local space: x,z in [0,16), y in [0,256).
// The sections pointer must refer to the chunk's Sections field so that lazily
// created sections are visible to the caller.
func PropagateBlockLight(sections *sectionsArray, x, y, z int, level uint8) {
	if level == 0 {
		return
	}

	queue := make([]lightNode, 0, 256)
	setBlockLightAt(sections, x, y, z, level)
	queue = append(queue, lightNode{x, y, z, level})

	bfsPropagate(sections, queue, blockLightAccessor)
}

// PropagateSkyLight fills sky light from the top of the chunk downward.
// Columns receive sky light level 15 from the highest unobstructed block down
// to the first opaque block. Only boundary cells (those adjacent to opaque
// blocks or chunk edges) are enqueued for the sideways BFS pass.
func PropagateSkyLight(sections *sectionsArray, heightMap [mcmath.ChunkSize * mcmath.ChunkSize]int) {
	queue := make([]lightNode, 0, 1024)

	// Phase 1: fill vertical columns with sky light = 15 above the heightmap.
	for x := 0; x < sectionSize; x++ {
		for z := 0; z < sectionSize; z++ {
			topY := heightMap[z*sectionSize+x]
			for y := mcmath.ChunkHeight - 1; y >= topY; y-- {
				ensureSection(sections, y)
				setSkyLightAt(sections, x, y, z, maxLight)
			}
		}
	}

	// Phase 2: enqueue only boundary cells that could propagate into darker areas.
	for x := 0; x < sectionSize; x++ {
		for z := 0; z < sectionSize; z++ {
			topY := heightMap[z*sectionSize+x]
			for y := mcmath.ChunkHeight - 1; y >= topY; y-- {
				if isSkyBoundary(sections, heightMap, x, y, z) {
					queue = append(queue, lightNode{x, y, z, maxLight})
				}
			}
		}
	}

	// Phase 3: BFS sideways/downward propagation through transparent blocks.
	bfsPropagate(sections, queue, skyLightAccessor)
}

// isSkyBoundary reports whether the sky-lit cell at (x, y, z) has at least one
// neighbor that is in-bounds, passable, and not already at sky light 15. Cells
// at the chunk's horizontal edges always qualify because light could propagate
// to a neighboring chunk (not modeled yet, but the BFS handles it via bounds).
func isSkyBoundary(sections *sectionsArray, heightMap [256]int, x, y, z int) bool {
	if x == 0 || x == sectionSize-1 || z == 0 || z == sectionSize-1 {
		return true
	}
	for _, off := range neighborOffsets {
		nx, ny, nz := x+off[0], y+off[1], z+off[2]
		if !inBounds(nx, ny, nz) {
			continue
		}
		blockID := getBlockAt(sections, nx, ny, nz)
		if block.IsSolid(blockID) && !block.IsTransparent(blockID) {
			continue
		}
		if getSkyLightAt(sections, nx, ny, nz) < maxLight {
			return true
		}
	}
	return false
}

// bfsPropagate is the shared BFS loop for both block light and sky light. It
// uses an index to advance through the queue, avoiding repeated slice header
// copies.
func bfsPropagate(sections *sectionsArray, queue []lightNode, acc lightAccessor) {
	for i := 0; i < len(queue); i++ {
		node := queue[i]

		for _, off := range neighborOffsets {
			nx, ny, nz := node.x+off[0], node.y+off[1], node.z+off[2]
			if !inBounds(nx, ny, nz) {
				continue
			}

			blockID := getBlockAt(sections, nx, ny, nz)
			if block.IsSolid(blockID) && !block.IsTransparent(blockID) {
				continue
			}

			filter := block.GetProperties(blockID).LightFilter
			reduction := uint8(1)
			if filter > reduction {
				reduction = filter
			}
			if node.level <= reduction {
				continue
			}
			newLevel := node.level - reduction

			current := acc.get(sections, nx, ny, nz)
			if newLevel > current {
				ensureSection(sections, ny)
				acc.set(sections, nx, ny, nz, newLevel)
				queue = append(queue, lightNode{nx, ny, nz, newLevel})
			}
		}
	}
}

// inBounds reports whether (x, y, z) is within a single chunk's local space.
func inBounds(x, y, z int) bool {
	return x >= 0 && x < sectionSize &&
		z >= 0 && z < sectionSize &&
		y >= 0 && y < mcmath.ChunkHeight
}

// ensureSection lazily initialises the section at the given world-Y if nil.
func ensureSection(sections *sectionsArray, y int) {
	si := y / mcmath.SectionHeight
	if si < 0 || si >= numSections {
		return
	}
	if sections[si] == nil {
		sections[si] = chunk.NewSection()
	}
}

// getBlockAt returns the block ID at chunk-local (x, y, z).
func getBlockAt(sections *sectionsArray, x, y, z int) uint16 {
	si := y / mcmath.SectionHeight
	sec := sections[si]
	if sec == nil {
		return 0
	}
	return sec.GetBlock(x, y%mcmath.SectionHeight, z)
}

// getBlockLightAt returns the block light level at chunk-local (x, y, z).
func getBlockLightAt(sections *sectionsArray, x, y, z int) uint8 {
	si := y / mcmath.SectionHeight
	sec := sections[si]
	if sec == nil {
		return 0
	}
	return sec.GetBlockLight(x, y%mcmath.SectionHeight, z)
}

// setBlockLightAt sets the block light level, lazily creating the section.
func setBlockLightAt(sections *sectionsArray, x, y, z int, level uint8) {
	si := y / mcmath.SectionHeight
	if sections[si] == nil {
		sections[si] = chunk.NewSection()
	}
	sections[si].SetBlockLight(x, y%mcmath.SectionHeight, z, level)
}

// getSkyLightAt returns the sky light level at chunk-local (x, y, z).
func getSkyLightAt(sections *sectionsArray, x, y, z int) uint8 {
	si := y / mcmath.SectionHeight
	sec := sections[si]
	if sec == nil {
		return 0
	}
	return sec.GetSkyLight(x, y%mcmath.SectionHeight, z)
}

// setSkyLightAt sets the sky light level at chunk-local (x, y, z).
func setSkyLightAt(sections *sectionsArray, x, y, z int, level uint8) {
	si := y / mcmath.SectionHeight
	if sections[si] == nil {
		sections[si] = chunk.NewSection()
	}
	sections[si].SetSkyLight(x, y%mcmath.SectionHeight, z, level)
}
