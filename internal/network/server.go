package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/fanxiyao/gomc/internal/protocol"
)

// Server listens for TCP connections, manages clients, and dispatches packets.
type Server struct {
	address  string
	port     int
	listener net.Listener
	registry *protocol.PacketRegistry

	mu      sync.RWMutex
	clients map[*Connection]struct{}

	cbMu         sync.RWMutex
	onConnect    func(conn *Connection)
	onDisconnect func(conn *Connection)
	onPacket     func(conn *Connection, packet protocol.Packet)

	done chan struct{}
}

// NewServer returns a Server that will bind to address:port.
func NewServer(address string, port int) *Server {
	return &Server{
		address:  address,
		port:     port,
		registry: protocol.DefaultRegistry(),
		clients:  make(map[*Connection]struct{}),
		done:     make(chan struct{}),
	}
}

// OnConnect registers a callback invoked when a new client connects.
func (s *Server) OnConnect(fn func(conn *Connection)) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	s.onConnect = fn
}

// OnDisconnect registers a callback invoked when a client disconnects.
func (s *Server) OnDisconnect(fn func(conn *Connection)) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	s.onDisconnect = fn
}

// OnPacket registers a callback invoked for every incoming packet.
func (s *Server) OnPacket(fn func(conn *Connection, packet protocol.Packet)) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	s.onPacket = fn
}

// Start begins accepting TCP connections. It blocks briefly to bind
// the listener, then spawns a goroutine that accepts in the background.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = ln

	go s.acceptLoop()
	return nil
}

// Stop closes the listener and disconnects every client.
func (s *Server) Stop() error {
	close(s.done)

	var firstErr error
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			firstErr = err
		}
	}

	s.mu.RLock()
	snapshot := make([]*Connection, 0, len(s.clients))
	for c := range s.clients {
		snapshot = append(snapshot, c)
	}
	s.mu.RUnlock()

	for _, c := range snapshot {
		_ = c.Close()
	}
	return firstErr
}

// Broadcast sends a packet to every connected client.
func (s *Server) Broadcast(packet protocol.Packet) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for c := range s.clients {
		_ = c.Send(packet)
	}
}

// BroadcastExcept sends a packet to every connected client except exclude.
func (s *Server) BroadcastExcept(packet protocol.Packet, exclude *Connection) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for c := range s.clients {
		if c != exclude {
			_ = c.Send(packet)
		}
	}
}

// ClientCount returns the number of currently connected clients.
func (s *Server) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// Addr returns the listener's address, useful when binding to port 0.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) acceptLoop() {
	for {
		rawConn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}
		conn := NewConnection(rawConn, s.registry)
		s.addClient(conn)
		go s.handleClient(conn)
	}
}

func (s *Server) addClient(conn *Connection) {
	s.mu.Lock()
	s.clients[conn] = struct{}{}
	s.mu.Unlock()

	s.cbMu.RLock()
	fn := s.onConnect
	s.cbMu.RUnlock()
	if fn != nil {
		fn(conn)
	}
}

func (s *Server) removeClient(conn *Connection) {
	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()

	_ = conn.Close()

	s.cbMu.RLock()
	fn := s.onDisconnect
	s.cbMu.RUnlock()
	if fn != nil {
		fn(conn)
	}
}

func (s *Server) handleClient(conn *Connection) {
	defer s.removeClient(conn)

	for {
		pkt, err := conn.ReadPacket()
		if err != nil {
			return
		}
		s.cbMu.RLock()
		fn := s.onPacket
		s.cbMu.RUnlock()
		if fn != nil {
			fn(conn, pkt)
		}
	}
}
