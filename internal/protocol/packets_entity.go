package protocol

import "io"

// SpawnEntitySync notifies clients about a new entity in the multiplayer session.
type SpawnEntitySync struct {
	EntityID   uint32
	EntityType uint8
	X          float32
	Y          float32
	Z          float32
}

func (p *SpawnEntitySync) ID() byte { return 0x40 }

func (p *SpawnEntitySync) Encode(w io.Writer) error {
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
	return WriteFloat32(w, p.Z)
}

func (p *SpawnEntitySync) Decode(r io.Reader) error {
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
	p.Z, err = ReadFloat32(r)
	return err
}

// MoveEntity updates an entity's position and velocity.
type MoveEntity struct {
	EntityID uint32
	X        float32
	Y        float32
	Z        float32
	VelX     float32
	VelY     float32
	VelZ     float32
}

func (p *MoveEntity) ID() byte { return 0x41 }

func (p *MoveEntity) Encode(w io.Writer) error {
	if err := WriteUint32(w, p.EntityID); err != nil {
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
	if err := WriteFloat32(w, p.VelX); err != nil {
		return err
	}
	if err := WriteFloat32(w, p.VelY); err != nil {
		return err
	}
	return WriteFloat32(w, p.VelZ)
}

func (p *MoveEntity) Decode(r io.Reader) error {
	var err error
	if p.EntityID, err = ReadUint32(r); err != nil {
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
	if p.VelX, err = ReadFloat32(r); err != nil {
		return err
	}
	if p.VelY, err = ReadFloat32(r); err != nil {
		return err
	}
	p.VelZ, err = ReadFloat32(r)
	return err
}

// RemoveEntity notifies clients that an entity has been removed.
type RemoveEntity struct {
	EntityID uint32
}

func (p *RemoveEntity) ID() byte { return 0x42 }

func (p *RemoveEntity) Encode(w io.Writer) error {
	return WriteUint32(w, p.EntityID)
}

func (p *RemoveEntity) Decode(r io.Reader) error {
	var err error
	p.EntityID, err = ReadUint32(r)
	return err
}

// PlayerPositionSync relays a player's position and orientation to other clients.
type PlayerPositionSync struct {
	X     float32
	Y     float32
	Z     float32
	Yaw   float32
	Pitch float32
}

func (p *PlayerPositionSync) ID() byte { return 0x43 }

func (p *PlayerPositionSync) Encode(w io.Writer) error {
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

func (p *PlayerPositionSync) Decode(r io.Reader) error {
	var err error
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
