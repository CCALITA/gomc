package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// Packet defines the interface for all network packets.
type Packet interface {
	ID() byte
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

// PacketRegistry maps packet IDs to factory functions that create new Packet instances.
type PacketRegistry struct {
	mu        sync.RWMutex
	factories map[byte]func() Packet
}

// NewPacketRegistry returns an empty PacketRegistry.
func NewPacketRegistry() *PacketRegistry {
	return &PacketRegistry{
		factories: make(map[byte]func() Packet),
	}
}

// Register associates a packet ID with a factory function.
func (r *PacketRegistry) Register(id byte, factory func() Packet) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[id] = factory
}

// Create returns a new Packet instance for the given ID.
func (r *PacketRegistry) Create(id byte) (Packet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[id]
	if !ok {
		return nil, fmt.Errorf("unknown packet ID: 0x%02X", id)
	}
	return factory(), nil
}

// Frame format: [ID:1byte][Length:4bytes-big-endian][Payload:N-bytes]

// WritePacket encodes a packet and writes the framed data to w.
func WritePacket(w io.Writer, p Packet) error {
	var buf bytes.Buffer
	if err := p.Encode(&buf); err != nil {
		return fmt.Errorf("failed to encode packet 0x%02X: %w", p.ID(), err)
	}

	payload := buf.Bytes()

	// Write ID
	if err := binary.Write(w, binary.BigEndian, p.ID()); err != nil {
		return fmt.Errorf("failed to write packet ID: %w", err)
	}

	// Write length
	length := uint32(len(payload))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write packet length: %w", err)
	}

	// Write payload
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("failed to write packet payload: %w", err)
	}

	return nil
}

// ReadPacket reads a framed packet from r, using reg to create the appropriate Packet type.
func ReadPacket(r io.Reader, reg *PacketRegistry) (Packet, error) {
	// Read ID
	var id byte
	if err := binary.Read(r, binary.BigEndian, &id); err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}

	// Read length
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read packet length: %w", err)
	}

	// Read payload
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("failed to read packet payload: %w", err)
	}

	// Create packet from registry
	pkt, err := reg.Create(id)
	if err != nil {
		return nil, err
	}

	// Decode payload into packet
	if err := pkt.Decode(bytes.NewReader(payload)); err != nil {
		return nil, fmt.Errorf("failed to decode packet 0x%02X: %w", id, err)
	}

	return pkt, nil
}
