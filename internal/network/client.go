package network

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fanxiyao/gomc/internal/protocol"
)

const (
	keepAliveInterval = 10 * time.Second
	keepAliveTimeout  = 30 * time.Second
)

// Client connects to a remote server, reads packets, and sends keep-alives.
type Client struct {
	conn     *Connection
	registry *protocol.PacketRegistry

	onPacket func(packet protocol.Packet)

	connected atomic.Bool

	// lastPong tracks the most recent Pong timestamp received from the server.
	lastPong   time.Time
	lastPongMu sync.Mutex

	done chan struct{}
	wg   sync.WaitGroup
}

// NewClient creates a disconnected Client.
func NewClient() *Client {
	return &Client{
		registry: protocol.DefaultRegistry(),
		done:     make(chan struct{}),
	}
}

// OnPacket registers a callback invoked for every incoming packet from the server.
func (c *Client) OnPacket(fn func(packet protocol.Packet)) {
	c.onPacket = fn
}

// Connect dials the server at address:port over TCP.
func (c *Client) Connect(address string, port int) error {
	addr := net.JoinHostPort(address, fmt.Sprintf("%d", port))
	rawConn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.conn = NewConnection(rawConn, c.registry)
	c.connected.Store(true)

	c.lastPongMu.Lock()
	c.lastPong = time.Now()
	c.lastPongMu.Unlock()

	c.wg.Add(2)
	go c.readLoop()
	go c.keepAliveLoop()

	return nil
}

// Disconnect closes the connection and stops background goroutines.
func (c *Client) Disconnect() error {
	if !c.connected.CompareAndSwap(true, false) {
		return nil
	}
	close(c.done)
	err := c.conn.Close()
	c.wg.Wait()
	return err
}

// Send writes a packet to the server.
func (c *Client) Send(packet protocol.Packet) error {
	if !c.connected.Load() {
		return net.ErrClosed
	}
	return c.conn.Send(packet)
}

// IsConnected reports whether the client is currently connected.
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

func (c *Client) readLoop() {
	defer c.wg.Done()

	for c.connected.Load() {
		pkt, err := c.conn.ReadPacket()
		if err != nil {
			c.connected.Store(false)
			return
		}

		// Track Pong for keep-alive monitoring.
		if _, ok := pkt.(*protocol.Pong); ok {
			c.lastPongMu.Lock()
			c.lastPong = time.Now()
			c.lastPongMu.Unlock()
		}

		if c.onPacket != nil {
			c.onPacket(pkt)
		}
	}
}

func (c *Client) keepAliveLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			// Check whether the server has responded recently.
			c.lastPongMu.Lock()
			elapsed := time.Since(c.lastPong)
			c.lastPongMu.Unlock()

			if elapsed > keepAliveTimeout {
				// Server is unresponsive; disconnect.
				c.connected.Store(false)
				_ = c.conn.Close()
				return
			}

			_ = c.Send(&protocol.KeepAlive{
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}
}
