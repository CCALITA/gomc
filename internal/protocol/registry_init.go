package protocol

// DefaultRegistry returns a PacketRegistry with all known packet types registered.
func DefaultRegistry() *PacketRegistry {
	reg := NewPacketRegistry()

	// Login packets
	reg.Register(0x01, func() Packet { return &LoginRequest{} })
	reg.Register(0x02, func() Packet { return &LoginResponse{} })
	reg.Register(0x03, func() Packet { return &Disconnect{} })

	// Play packets
	reg.Register(0x10, func() Packet { return &PlayerPosition{} })
	reg.Register(0x11, func() Packet { return &PlayerAction{} })
	reg.Register(0x12, func() Packet { return &ChatMessage{} })

	// World packets
	reg.Register(0x20, func() Packet { return &ChunkData{} })
	reg.Register(0x21, func() Packet { return &BlockChange{} })
	reg.Register(0x22, func() Packet { return &SpawnEntity{} })

	// System packets
	reg.Register(0x30, func() Packet { return &KeepAlive{} })
	reg.Register(0x31, func() Packet { return &Ping{} })
	reg.Register(0x32, func() Packet { return &Pong{} })

	return reg
}
