package chunk

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// Binary format (all little-endian):
//
//   Header:
//     ChunkX       int32
//     ChunkZ       int32
//     SectionMask  uint16   (bit i set → section i is present)
//
//   Per present section:
//     PaletteLen   uint16
//     Palette      [PaletteLen]uint16
//     Wide         uint8    (1 if indices are uint16, 0 for uint8)
//     Indices      [4096]uint8 or [4096]uint16
//
//   HeightMap:
//     [256]int32

// SerializeChunk encodes a chunk into a compact binary representation.
func SerializeChunk(c *Chunk) ([]byte, error) {
	var buf bytes.Buffer

	// Header.
	if err := binary.Write(&buf, binary.LittleEndian, c.Pos.X); err != nil {
		return nil, fmt.Errorf("failed to write chunk X: %w", err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, c.Pos.Z); err != nil {
		return nil, fmt.Errorf("failed to write chunk Z: %w", err)
	}

	// Section presence bitmask.
	var sectionMask uint16
	for i, sec := range c.Sections {
		if sec != nil && !sec.IsEmpty() {
			sectionMask |= 1 << uint(i)
		}
	}
	if err := binary.Write(&buf, binary.LittleEndian, sectionMask); err != nil {
		return nil, fmt.Errorf("failed to write section mask: %w", err)
	}

	// Sections.
	for i := 0; i < numSections; i++ {
		if sectionMask&(1<<uint(i)) == 0 {
			continue
		}
		sec := c.Sections[i]
		if err := serializeSection(&buf, sec); err != nil {
			return nil, fmt.Errorf("failed to serialize section %d: %w", i, err)
		}
	}

	// HeightMap.
	for i := 0; i < mcmath.ChunkSize*mcmath.ChunkSize; i++ {
		if err := binary.Write(&buf, binary.LittleEndian, int32(c.HeightMap[i])); err != nil {
			return nil, fmt.Errorf("failed to write height map: %w", err)
		}
	}

	return buf.Bytes(), nil
}

func serializeSection(buf *bytes.Buffer, sec *Section) error {
	pLen := uint16(len(sec.palette))
	if err := binary.Write(buf, binary.LittleEndian, pLen); err != nil {
		return fmt.Errorf("failed to write palette length: %w", err)
	}
	for _, id := range sec.palette {
		if err := binary.Write(buf, binary.LittleEndian, id); err != nil {
			return fmt.Errorf("failed to write palette entry: %w", err)
		}
	}

	wide := sec.wide != nil
	if wide {
		if err := buf.WriteByte(1); err != nil {
			return fmt.Errorf("failed to write wide flag: %w", err)
		}
		if err := binary.Write(buf, binary.LittleEndian, sec.wide[:sectionVolume]); err != nil {
			return fmt.Errorf("failed to write wide indices: %w", err)
		}
	} else {
		if err := buf.WriteByte(0); err != nil {
			return fmt.Errorf("failed to write wide flag: %w", err)
		}
		if _, err := buf.Write(sec.indices8[:]); err != nil {
			return fmt.Errorf("failed to write indices: %w", err)
		}
	}

	return nil
}

// DeserializeChunk decodes a chunk from its binary representation.
func DeserializeChunk(data []byte) (*Chunk, error) {
	r := bytes.NewReader(data)
	c := &Chunk{}

	if err := binary.Read(r, binary.LittleEndian, &c.Pos.X); err != nil {
		return nil, fmt.Errorf("failed to read chunk X: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Pos.Z); err != nil {
		return nil, fmt.Errorf("failed to read chunk Z: %w", err)
	}

	var sectionMask uint16
	if err := binary.Read(r, binary.LittleEndian, &sectionMask); err != nil {
		return nil, fmt.Errorf("failed to read section mask: %w", err)
	}

	for i := 0; i < numSections; i++ {
		if sectionMask&(1<<uint(i)) == 0 {
			continue
		}
		sec, err := deserializeSection(r)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize section %d: %w", i, err)
		}
		c.Sections[i] = sec
	}

	// HeightMap.
	for i := 0; i < mcmath.ChunkSize*mcmath.ChunkSize; i++ {
		var val int32
		if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
			return nil, fmt.Errorf("failed to read height map: %w", err)
		}
		c.HeightMap[i] = int(val)
	}

	return c, nil
}

func deserializeSection(r *bytes.Reader) (*Section, error) {
	var pLen uint16
	if err := binary.Read(r, binary.LittleEndian, &pLen); err != nil {
		return nil, fmt.Errorf("failed to read palette length: %w", err)
	}

	sec := &Section{
		palette: make([]uint16, pLen),
	}

	for i := uint16(0); i < pLen; i++ {
		if err := binary.Read(r, binary.LittleEndian, &sec.palette[i]); err != nil {
			return nil, fmt.Errorf("failed to read palette entry: %w", err)
		}
	}

	wideByte, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read wide flag: %w", err)
	}

	if wideByte == 1 {
		sec.wide = make([]uint16, sectionVolume)
		if err := binary.Read(r, binary.LittleEndian, sec.wide); err != nil {
			return nil, fmt.Errorf("failed to read wide indices: %w", err)
		}
	} else {
		if _, err := r.Read(sec.indices8[:]); err != nil {
			return nil, fmt.Errorf("failed to read indices: %w", err)
		}
	}

	// Recompute count from data.
	sec.count = 0
	for idx := 0; idx < sectionVolume; idx++ {
		var ci int
		if sec.wide != nil {
			ci = int(sec.wide[idx])
		} else {
			ci = int(sec.indices8[idx])
		}
		if sec.palette[ci] != 0 {
			sec.count++
		}
	}

	return sec, nil
}
