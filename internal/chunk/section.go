package chunk

import "github.com/fanxiyao/gomc/internal/mcmath"

const sectionVolume = mcmath.ChunkSize * mcmath.ChunkSize * mcmath.SectionHeight // 4096

// Section stores a 16x16x16 cube of blocks using a palette + indices scheme.
// The palette maps compact indices to real block IDs, and each voxel stores
// only the compact index, keeping memory usage low for sections with few
// distinct block types.
//
// BlockLight and SkyLight store per-voxel light levels (0-15).
type Section struct {
	palette    []uint16
	indices8   [sectionVolume]uint8 // compact indices when len(palette) <= 256
	wide       []uint16             // compact indices when len(palette) > 256 (lazy-allocated)
	count      int                  // number of non-air (non-zero) blocks
	BlockLight [sectionVolume]uint8
	SkyLight   [sectionVolume]uint8
}

// blockIndex converts local (x, y, z) coordinates into a flat array index.
// Layout: Y * 256 + Z * 16 + X  (matches Minecraft's section ordering).
func blockIndex(x, y, z int) int {
	return y*mcmath.ChunkSize*mcmath.ChunkSize + z*mcmath.ChunkSize + x
}

// NewSection returns an empty section. Palette entry 0 is always air (block ID 0).
func NewSection() *Section {
	return &Section{
		palette: []uint16{0}, // index 0 → air
	}
}

// paletteIndex returns the compact index for the given block ID, adding it to
// the palette if it is not already present. When the palette outgrows 256
// entries, the section migrates from uint8 to uint16 storage.
func (s *Section) paletteIndex(id uint16) int {
	for i, pid := range s.palette {
		if pid == id {
			return i
		}
	}
	// New block type — append to palette.
	idx := len(s.palette)
	s.palette = append(s.palette, id)

	// Migrate to wide storage when palette exceeds 256 entries.
	if idx == 256 && s.wide == nil {
		s.wide = make([]uint16, sectionVolume)
		for i, v := range s.indices8 {
			s.wide[i] = uint16(v)
		}
	}
	return idx
}

// GetBlock returns the block ID at the given local coordinates.
// Coordinates must be in [0, 16).
func (s *Section) GetBlock(x, y, z int) uint16 {
	idx := blockIndex(x, y, z)
	var ci int
	if s.wide != nil {
		ci = int(s.wide[idx])
	} else {
		ci = int(s.indices8[idx])
	}
	return s.palette[ci]
}

// SetBlock sets the block ID at the given local coordinates.
// Coordinates must be in [0, 16). Passing id 0 counts as air.
func (s *Section) SetBlock(x, y, z int, id uint16) {
	idx := blockIndex(x, y, z)
	old := s.GetBlock(x, y, z)

	// Update non-air block count.
	if old == 0 && id != 0 {
		s.count++
	} else if old != 0 && id == 0 {
		s.count--
	}

	ci := s.paletteIndex(id)
	if s.wide != nil {
		s.wide[idx] = uint16(ci)
	} else {
		s.indices8[idx] = uint8(ci)
	}
}

// IsEmpty returns true when the section contains only air.
func (s *Section) IsEmpty() bool {
	return s.count == 0
}

// PaletteSize returns the number of distinct block types tracked by this section.
func (s *Section) PaletteSize() int {
	return len(s.palette)
}

// GetBlockLight returns the block light level at local coordinates (x, y, z).
// Coordinates must be in [0, 16).
func (s *Section) GetBlockLight(x, y, z int) uint8 {
	return s.BlockLight[blockIndex(x, y, z)]
}

// SetBlockLight sets the block light level at local coordinates (x, y, z).
// Coordinates must be in [0, 16). Level is clamped to [0, 15].
func (s *Section) SetBlockLight(x, y, z int, level uint8) {
	if level > 15 {
		level = 15
	}
	s.BlockLight[blockIndex(x, y, z)] = level
}

// GetSkyLight returns the sky light level at local coordinates (x, y, z).
// Coordinates must be in [0, 16).
func (s *Section) GetSkyLight(x, y, z int) uint8 {
	return s.SkyLight[blockIndex(x, y, z)]
}

// SetSkyLight sets the sky light level at local coordinates (x, y, z).
// Coordinates must be in [0, 16). Level is clamped to [0, 15].
func (s *Section) SetSkyLight(x, y, z int, level uint8) {
	if level > 15 {
		level = 15
	}
	s.SkyLight[blockIndex(x, y, z)] = level
}

// ClearBlockLight zeroes all block light values in this section.
func (s *Section) ClearBlockLight() {
	s.BlockLight = [sectionVolume]uint8{}
}

// ClearSkyLight zeroes all sky light values in this section.
func (s *Section) ClearSkyLight() {
	s.SkyLight = [sectionVolume]uint8{}
}
