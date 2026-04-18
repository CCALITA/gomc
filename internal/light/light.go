package light

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// LightEngine computes block light and sky light for a chunk.
type LightEngine struct{}

// NewLightEngine returns a new LightEngine.
func NewLightEngine() *LightEngine {
	return &LightEngine{}
}

// ComputeChunkLight calculates block light and sky light for all sections in
// the given chunk. Neighbor chunks (N, E, S, W) are currently reserved for
// cross-chunk propagation but not yet used.
//
// Block light is emitted by light-source blocks (Torch=14, Lava=15) and
// decreases by 1 per transparent block, blocked by solid opaque blocks.
// Sky light starts at 15 from the highest unobstructed block and propagates
// downward and sideways through transparent blocks.
func (e *LightEngine) ComputeChunkLight(c *chunk.Chunk, _ [4]*chunk.Chunk) {
	resetLight(c)
	computeBlockLight(c)
	computeSkyLight(c)
}

// resetLight zeroes all light arrays in every initialised section.
func resetLight(c *chunk.Chunk) {
	for i := 0; i < numSections; i++ {
		sec := c.GetSection(i)
		if sec == nil {
			continue
		}
		sec.ClearBlockLight()
		sec.ClearSkyLight()
	}
}

// computeBlockLight scans the chunk for light-emitting blocks and propagates
// each source via BFS.
func computeBlockLight(c *chunk.Chunk) {
	sources := findLightSources(c)

	for _, src := range sources {
		PropagateBlockLight(&c.Sections, src.x, src.y, src.z, src.level)
	}
}

// findLightSources returns all light-emitting blocks in the chunk.
func findLightSources(c *chunk.Chunk) []lightNode {
	sources := make([]lightNode, 0, 16)
	for si := 0; si < numSections; si++ {
		sec := c.GetSection(si)
		if sec == nil {
			continue
		}
		baseY := si * mcmath.SectionHeight
		for x := 0; x < sectionSize; x++ {
			for y := 0; y < mcmath.SectionHeight; y++ {
				for z := 0; z < sectionSize; z++ {
					id := sec.GetBlock(x, y, z)
					emission := block.GetProperties(id).LightEmission
					if emission > 0 {
						sources = append(sources, lightNode{
							x: x, y: baseY + y, z: z, level: emission,
						})
					}
				}
			}
		}
	}
	return sources
}

// computeSkyLight fills sky light from the top down using the chunk's height map.
func computeSkyLight(c *chunk.Chunk) {
	PropagateSkyLight(&c.Sections, c.HeightMap)
}
