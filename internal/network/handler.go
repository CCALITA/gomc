package network

import (
	"sync"
	"time"

	"github.com/fanxiyao/gomc/internal/protocol"
)

// PacketHandler dispatches incoming packets to registered handler functions.
type PacketHandler struct {
	mu       sync.RWMutex
	handlers map[byte]func(*Connection, protocol.Packet)
}

// NewPacketHandler returns a PacketHandler pre-loaded with default handlers
// for KeepAlive and Ping.
func NewPacketHandler() *PacketHandler {
	h := &PacketHandler{
		handlers: make(map[byte]func(*Connection, protocol.Packet)),
	}
	h.registerDefaults()
	return h
}

// RegisterHandler maps a packet ID to a handler function.
// It overwrites any previously registered handler for the same ID.
func (h *PacketHandler) RegisterHandler(packetID byte, handler func(*Connection, protocol.Packet)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[packetID] = handler
}

// Handle dispatches the packet to the registered handler, if any.
func (h *PacketHandler) Handle(conn *Connection, packet protocol.Packet) {
	h.mu.RLock()
	fn, ok := h.handlers[packet.ID()]
	h.mu.RUnlock()

	if ok {
		fn(conn, packet)
	}
}

// registerDefaults installs built-in handlers for system packets.
func (h *PacketHandler) registerDefaults() {
	// KeepAlive -> respond with Pong
	h.handlers[0x30] = func(conn *Connection, _ protocol.Packet) {
		_ = conn.Send(&protocol.Pong{
			Timestamp: time.Now().UnixMilli(),
		})
	}

	// Ping -> respond with Pong
	h.handlers[0x31] = func(conn *Connection, pkt protocol.Packet) {
		ping, ok := pkt.(*protocol.Ping)
		if !ok {
			return
		}
		_ = conn.Send(&protocol.Pong{
			Timestamp: ping.Timestamp,
		})
	}
}
