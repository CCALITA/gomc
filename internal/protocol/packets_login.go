package protocol

import "io"

// LoginRequest is sent by the client to authenticate with a username.
type LoginRequest struct {
	Username string
}

func (p *LoginRequest) ID() byte { return 0x01 }

func (p *LoginRequest) Encode(w io.Writer) error {
	return WriteString(w, p.Username)
}

func (p *LoginRequest) Decode(r io.Reader) error {
	var err error
	p.Username, err = ReadString(r)
	return err
}

// LoginResponse is sent by the server upon successful login.
type LoginResponse struct {
	EntityID     uint32
	SpawnX       float32
	SpawnY       float32
	SpawnZ       float32
}

func (p *LoginResponse) ID() byte { return 0x02 }

func (p *LoginResponse) Encode(w io.Writer) error {
	if err := WriteUint32(w, p.EntityID); err != nil {
		return err
	}
	if err := WriteFloat32(w, p.SpawnX); err != nil {
		return err
	}
	if err := WriteFloat32(w, p.SpawnY); err != nil {
		return err
	}
	return WriteFloat32(w, p.SpawnZ)
}

func (p *LoginResponse) Decode(r io.Reader) error {
	var err error
	if p.EntityID, err = ReadUint32(r); err != nil {
		return err
	}
	if p.SpawnX, err = ReadFloat32(r); err != nil {
		return err
	}
	if p.SpawnY, err = ReadFloat32(r); err != nil {
		return err
	}
	p.SpawnZ, err = ReadFloat32(r)
	return err
}

// Disconnect is sent to terminate a connection with a reason.
type Disconnect struct {
	Reason string
}

func (p *Disconnect) ID() byte { return 0x03 }

func (p *Disconnect) Encode(w io.Writer) error {
	return WriteString(w, p.Reason)
}

func (p *Disconnect) Decode(r io.Reader) error {
	var err error
	p.Reason, err = ReadString(r)
	return err
}
