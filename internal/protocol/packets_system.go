package protocol

import "io"

// KeepAlive is sent periodically to maintain the connection.
type KeepAlive struct {
	Timestamp int64
}

func (p *KeepAlive) ID() byte { return 0x30 }

func (p *KeepAlive) Encode(w io.Writer) error {
	return WriteInt64(w, p.Timestamp)
}

func (p *KeepAlive) Decode(r io.Reader) error {
	var err error
	p.Timestamp, err = ReadInt64(r)
	return err
}

// Ping is sent to measure round-trip latency.
type Ping struct {
	Timestamp int64
}

func (p *Ping) ID() byte { return 0x31 }

func (p *Ping) Encode(w io.Writer) error {
	return WriteInt64(w, p.Timestamp)
}

func (p *Ping) Decode(r io.Reader) error {
	var err error
	p.Timestamp, err = ReadInt64(r)
	return err
}

// Pong is the response to a Ping.
type Pong struct {
	Timestamp int64
}

func (p *Pong) ID() byte { return 0x32 }

func (p *Pong) Encode(w io.Writer) error {
	return WriteInt64(w, p.Timestamp)
}

func (p *Pong) Decode(r io.Reader) error {
	var err error
	p.Timestamp, err = ReadInt64(r)
	return err
}
