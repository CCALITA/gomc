package chunk

import "github.com/fanxiyao/gomc/internal/mcmath"

const (
	numSections = mcmath.ChunkHeight / mcmath.SectionHeight // 16
)

// Chunk represents a 16x256x16 column of blocks divided into 16 sections.
type Chunk struct {
	Pos       mcmath.ChunkPos
	Sections  [numSections]*Section
	HeightMap [mcmath.ChunkSize * mcmath.ChunkSize]int // 16x16 — highest non-air Y+1 per column
}

// heightIndex converts (x, z) into a flat HeightMap index.
func heightIndex(x, z int) int {
	return z*mcmath.ChunkSize + x
}

// getOrCreateSection returns the section at the given index, creating it lazily
// if it does not yet exist.
func (c *Chunk) getOrCreateSection(idx int) *Section {
	if c.Sections[idx] == nil {
		c.Sections[idx] = NewSection()
	}
	return c.Sections[idx]
}

// GetSection returns the section at the given index (0..15), or nil.
func (c *Chunk) GetSection(index int) *Section {
	if index < 0 || index >= numSections {
		return nil
	}
	return c.Sections[index]
}

// GetBlock returns the block ID at local coordinates (x, y, z).
// x and z must be in [0, 16); y must be in [0, 256).
// Returns 0 (air) for coordinates that fall in an uninitialised section.
func (c *Chunk) GetBlock(x, y, z int) uint16 {
	if y < 0 || y >= mcmath.ChunkHeight {
		return 0
	}
	si := y / mcmath.SectionHeight
	sec := c.Sections[si]
	if sec == nil {
		return 0
	}
	return sec.GetBlock(x, y%mcmath.SectionHeight, z)
}

// SetBlock sets the block ID at local coordinates (x, y, z).
// x and z must be in [0, 16); y must be in [0, 256).
// Sections are lazily initialised on first write.
func (c *Chunk) SetBlock(x, y, z int, id uint16) {
	if y < 0 || y >= mcmath.ChunkHeight {
		return
	}
	si := y / mcmath.SectionHeight
	sec := c.getOrCreateSection(si)
	sec.SetBlock(x, y%mcmath.SectionHeight, z, id)
	c.updateHeightMap(x, y, z, id)
}

// updateHeightMap maintains the per-column maximum-Y tracker.
func (c *Chunk) updateHeightMap(x, y, z int, id uint16) {
	hi := heightIndex(x, z)
	if id != 0 {
		if y+1 > c.HeightMap[hi] {
			c.HeightMap[hi] = y + 1
		}
	} else if y+1 >= c.HeightMap[hi] {
		// Removed the highest block — scan downward to find the new max.
		c.HeightMap[hi] = 0
		for ny := y; ny >= 0; ny-- {
			if c.GetBlock(x, ny, z) != 0 {
				c.HeightMap[hi] = ny + 1
				break
			}
		}
	}
}

// HighestBlock returns the Y coordinate of the highest non-air block in the
// column (x, z), or -1 if the column is empty.
func (c *Chunk) HighestBlock(x, z int) int {
	h := c.HeightMap[heightIndex(x, z)]
	if h == 0 {
		return -1
	}
	return h - 1
}

// IsEmpty returns true when the chunk contains no non-air blocks.
func (c *Chunk) IsEmpty() bool {
	for _, sec := range c.Sections {
		if sec != nil && !sec.IsEmpty() {
			return false
		}
	}
	return true
}
