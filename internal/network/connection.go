package network

import (
	"bufio"
	"net"
	"sync"
	"sync/atomic"

	"github.com/fanxiyao/gomc/internal/protocol"
)

// Connection wraps a net.Conn with packet-level read/write and thread safety.
type Connection struct {
	conn     net.Conn
	reader   *bufio.Reader
	registry *protocol.PacketRegistry
	writeMu  sync.Mutex
	alive    atomic.Bool
}

// NewConnection creates a Connection that uses registry for packet framing.
func NewConnection(conn net.Conn, registry *protocol.PacketRegistry) *Connection {
	c := &Connection{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		registry: registry,
	}
	c.alive.Store(true)
	return c
}

// Send encodes and writes a packet to the underlying connection.
// It is safe for concurrent use.
func (c *Connection) Send(packet protocol.Packet) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if !c.alive.Load() {
		return net.ErrClosed
	}
	return protocol.WritePacket(c.conn, packet)
}

// ReadPacket performs a blocking read of the next framed packet.
func (c *Connection) ReadPacket() (protocol.Packet, error) {
	return protocol.ReadPacket(c.reader, c.registry)
}

// Close shuts down the connection.
func (c *Connection) Close() error {
	if c.alive.CompareAndSwap(true, false) {
		return c.conn.Close()
	}
	return nil
}

// RemoteAddr returns the remote address as a string.
func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

// IsAlive reports whether the connection has not been closed.
func (c *Connection) IsAlive() bool {
	return c.alive.Load()
}
