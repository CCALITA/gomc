package protocol

import (
	"fmt"
	"io"
)

// ChunkData carries the raw voxel data for a chunk column.
type ChunkData struct {
	ChunkX int32
	ChunkZ int32
	Data   []byte
}

func (p *ChunkData) ID() byte { return 0x20 }

func (p *ChunkData) Encode(w io.Writer) error {
	if err := WriteInt32(w, p.ChunkX); err != nil {
		return err
	}
	if err := WriteInt32(w, p.ChunkZ); err != nil {
		return err
	}
	// Write data length as uint32, then raw bytes
	if err := WriteUint32(w, uint32(len(p.Data))); err != nil {
		return err
	}
	if _, err := w.Write(p.Data); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}
	return nil
}

func (p *ChunkData) Decode(r io.Reader) error {
	var err error
	if p.ChunkX, err = ReadInt32(r); err != nil {
		return err
	}
	if p.ChunkZ, err = ReadInt32(r); err != nil {
		return err
	}
	dataLen, err := ReadUint32(r)
	if err != nil {
		return err
	}
	p.Data = make([]byte, dataLen)
	if _, err := io.ReadFull(r, p.Data); err != nil {
		return fmt.Errorf("failed to read chunk data: %w", err)
	}
	return nil
}

// BlockChange represents a single block update in the world.
type BlockChange struct {
	X       int32
	Y       int32
	Z       int32
	BlockID uint16
}

func (p *BlockChange) ID() byte { return 0x21 }

func (p *BlockChange) Encode(w io.Writer) error {
	if err := WriteInt32(w, p.X); err != nil {
		return err
	}
	if err := WriteInt32(w, p.Y); err != nil {
		return err
	}
	if err := WriteInt32(w, p.Z); err != nil {
		return err
	}
	return WriteUint16(w, p.BlockID)
}

func (p *BlockChange) Decode(r io.Reader) error {
	var err error
	if p.X, err = ReadInt32(r); err != nil {
		return err
	}
	if p.Y, err = ReadInt32(r); err != nil {
		return err
	}
	if p.Z, err = ReadInt32(r); err != nil {
		return err
	}
	p.BlockID, err = ReadUint16(r)
	return err
}

// SpawnEntity notifies the client about a new entity in the world.
type SpawnEntity struct {
	EntityID   uint32
	EntityType byte
	X          float32
	Y          float32
	Z          float32
	Yaw        float32
	Pitch      float32
}

func (p *SpawnEntity) ID() byte { return 0x22 }

func (p *SpawnEntity) Encode(w io.Writer) error {
	if err := WriteUint32(w, p.EntityID); err != nil {
		return err
	}
	if err := WriteByte_(w, p.EntityType); err != nil {
		return err
	}
	if err := WriteFloat32(w, p.X); err != nil {
		return err
	}
	if err := WriteFloat32(w, p.Y); err != nil {
		return err
	}
	if err := WriteFloat32(w, p.Z); err != nil {
		return err
	}
	if err := WriteFloat32(w, p.Yaw); err != nil {
		return err
	}
	return WriteFloat32(w, p.Pitch)
}

func (p *SpawnEntity) Decode(r io.Reader) error {
	var err error
	if p.EntityID, err = ReadUint32(r); err != nil {
		return err
	}
	if p.EntityType, err = ReadByte_(r); err != nil {
		return err
	}
	if p.X, err = ReadFloat32(r); err != nil {
		return err
	}
	if p.Y, err = ReadFloat32(r); err != nil {
		return err
	}
	if p.Z, err = ReadFloat32(r); err != nil {
		return err
	}
	if p.Yaw, err = ReadFloat32(r); err != nil {
		return err
	}
	p.Pitch, err = ReadFloat32(r)
	return err
}
