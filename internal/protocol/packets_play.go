package protocol

import "io"

// PlayerPosition reports the player's current position and orientation.
type PlayerPosition struct {
	X        float32
	Y        float32
	Z        float32
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (p *PlayerPosition) ID() byte { return 0x10 }

func (p *PlayerPosition) Encode(w io.Writer) error {
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
	if err := WriteFloat32(w, p.Pitch); err != nil {
		return err
	}
	return WriteBool(w, p.OnGround)
}

func (p *PlayerPosition) Decode(r io.Reader) error {
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
	if p.Pitch, err = ReadFloat32(r); err != nil {
		return err
	}
	p.OnGround, err = ReadBool(r)
	return err
}

// PlayerAction represents an action the player takes on a block.
type PlayerAction struct {
	ActionType byte
	BlockX     int32
	BlockY     int32
	BlockZ     int32
	Face       byte
}

func (p *PlayerAction) ID() byte { return 0x11 }

func (p *PlayerAction) Encode(w io.Writer) error {
	if err := WriteByte_(w, p.ActionType); err != nil {
		return err
	}
	if err := WriteInt32(w, p.BlockX); err != nil {
		return err
	}
	if err := WriteInt32(w, p.BlockY); err != nil {
		return err
	}
	if err := WriteInt32(w, p.BlockZ); err != nil {
		return err
	}
	return WriteByte_(w, p.Face)
}

func (p *PlayerAction) Decode(r io.Reader) error {
	var err error
	if p.ActionType, err = ReadByte_(r); err != nil {
		return err
	}
	if p.BlockX, err = ReadInt32(r); err != nil {
		return err
	}
	if p.BlockY, err = ReadInt32(r); err != nil {
		return err
	}
	if p.BlockZ, err = ReadInt32(r); err != nil {
		return err
	}
	p.Face, err = ReadByte_(r)
	return err
}

// ChatMessage carries a chat message from a sender.
type ChatMessage struct {
	Sender  string
	Message string
}

func (p *ChatMessage) ID() byte { return 0x12 }

func (p *ChatMessage) Encode(w io.Writer) error {
	if err := WriteString(w, p.Sender); err != nil {
		return err
	}
	return WriteString(w, p.Message)
}

func (p *ChatMessage) Decode(r io.Reader) error {
	var err error
	if p.Sender, err = ReadString(r); err != nil {
		return err
	}
	p.Message, err = ReadString(r)
	return err
}
