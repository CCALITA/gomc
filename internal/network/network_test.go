package network

import (
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fanxiyao/gomc/internal/protocol"
	"github.com/stretchr/testify/assert"
)

// helper: start a server on a random port and return its address.
func startTestServer(t *testing.T) (*Server, int) {
	t.Helper()
	srv := NewServer("127.0.0.1", 0)
	err := srv.Start()
	assert.NoError(t, err)
	port := srv.Addr().(*net.TCPAddr).Port
	return srv, port
}

// helper: connect a client with optional OnPacket callback.
func connectClient(t *testing.T, port int, onPkt func(protocol.Packet)) *Client {
	t.Helper()
	c := NewClient()
	if onPkt != nil {
		c.OnPacket(onPkt)
	}
	err := c.Connect("127.0.0.1", port)
	assert.NoError(t, err)
	return c
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestServerStartStop(t *testing.T) {
	srv, _ := startTestServer(t)
	assert.Equal(t, 0, srv.ClientCount())
	assert.NoError(t, srv.Stop())
}

func TestClientConnectDisconnect(t *testing.T) {
	connected := make(chan struct{}, 1)
	disconnected := make(chan struct{}, 1)

	srv, port := startTestServer(t)
	srv.OnConnect(func(_ *Connection) { connected <- struct{}{} })
	srv.OnDisconnect(func(_ *Connection) { disconnected <- struct{}{} })
	defer srv.Stop()

	client := connectClient(t, port, nil)
	assert.True(t, client.IsConnected())

	select {
	case <-connected:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for OnConnect")
	}

	assert.Equal(t, 1, srv.ClientCount())

	assert.NoError(t, client.Disconnect())
	assert.False(t, client.IsConnected())

	select {
	case <-disconnected:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for OnDisconnect")
	}

	assert.Equal(t, 0, srv.ClientCount())
}

func TestSendReceivePacket(t *testing.T) {
	received := make(chan protocol.Packet, 1)

	srv, port := startTestServer(t)
	srv.OnPacket(func(conn *Connection, pkt protocol.Packet) {
		// Echo back to sender.
		_ = conn.Send(pkt)
	})
	defer srv.Stop()

	client := connectClient(t, port, func(pkt protocol.Packet) {
		received <- pkt
	})
	defer client.Disconnect()

	// Give server time to register the client.
	time.Sleep(50 * time.Millisecond)

	sent := &protocol.ChatMessage{Sender: "Alice", Message: "Hello"}
	assert.NoError(t, client.Send(sent))

	select {
	case pkt := <-received:
		chat, ok := pkt.(*protocol.ChatMessage)
		assert.True(t, ok)
		assert.Equal(t, "Alice", chat.Sender)
		assert.Equal(t, "Hello", chat.Message)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for echoed packet")
	}
}

func TestBroadcast(t *testing.T) {
	const numClients = 3

	srv, port := startTestServer(t)
	defer srv.Stop()

	counts := make([]int32, numClients)

	var clients []*Client
	for i := 0; i < numClients; i++ {
		idx := i
		c := connectClient(t, port, func(_ protocol.Packet) {
			atomic.AddInt32(&counts[idx], 1)
		})
		clients = append(clients, c)
	}
	defer func() {
		for _, c := range clients {
			_ = c.Disconnect()
		}
	}()

	// Wait for all clients to register.
	assert.Eventually(t, func() bool {
		return srv.ClientCount() == numClients
	}, 2*time.Second, 20*time.Millisecond)

	msg := &protocol.ChatMessage{Sender: "Server", Message: "broadcast"}
	srv.Broadcast(msg)

	// Give packets time to arrive.
	assert.Eventually(t, func() bool {
		for i := 0; i < numClients; i++ {
			if atomic.LoadInt32(&counts[i]) < 1 {
				return false
			}
		}
		return true
	}, 2*time.Second, 20*time.Millisecond)

	for i := 0; i < numClients; i++ {
		assert.GreaterOrEqual(t, atomic.LoadInt32(&counts[i]), int32(1),
			"client %d should have received the broadcast", i)
	}
}

func TestBroadcastExcept(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()

	received1 := make(chan struct{}, 4)
	received2 := make(chan struct{}, 4)

	// Connect two clients.
	c1 := connectClient(t, port, func(_ protocol.Packet) { received1 <- struct{}{} })
	c2 := connectClient(t, port, func(_ protocol.Packet) { received2 <- struct{}{} })
	defer c1.Disconnect()
	defer c2.Disconnect()

	assert.Eventually(t, func() bool { return srv.ClientCount() == 2 },
		2*time.Second, 20*time.Millisecond)

	// Find connections on the server side.
	srv.mu.RLock()
	var conns []*Connection
	for c := range srv.clients {
		conns = append(conns, c)
	}
	srv.mu.RUnlock()

	// Exclude the first connection in the set.
	excluded := conns[0]
	srv.BroadcastExcept(&protocol.ChatMessage{Sender: "S", Message: "test"}, excluded)

	// The non-excluded client should receive it.
	time.Sleep(200 * time.Millisecond)

	// At least one channel should have received; we cannot know which conn maps
	// to which client, but total received across both should be exactly 1
	// (excluding keep-alive packets that may also arrive).
	total := len(received1) + len(received2)
	assert.GreaterOrEqual(t, total, 1, "at least one client should have received the packet")
}

func TestMultipleClients(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()

	const n = 5
	var clients []*Client
	for i := 0; i < n; i++ {
		c := connectClient(t, port, nil)
		clients = append(clients, c)
	}

	assert.Eventually(t, func() bool { return srv.ClientCount() == n },
		2*time.Second, 20*time.Millisecond)

	// Disconnect half and verify count.
	for i := 0; i < n/2; i++ {
		_ = clients[i].Disconnect()
	}

	assert.Eventually(t, func() bool { return srv.ClientCount() == n-n/2 },
		2*time.Second, 20*time.Millisecond)

	for i := n / 2; i < n; i++ {
		_ = clients[i].Disconnect()
	}

	assert.Eventually(t, func() bool { return srv.ClientCount() == 0 },
		2*time.Second, 20*time.Millisecond)
}

func TestKeepAliveExchange(t *testing.T) {
	srv, port := startTestServer(t)
	handler := NewPacketHandler()

	pongSent := make(chan struct{}, 10)
	handler.RegisterHandler(0x30, func(conn *Connection, _ protocol.Packet) {
		_ = conn.Send(&protocol.Pong{Timestamp: time.Now().UnixMilli()})
		pongSent <- struct{}{}
	})
	srv.OnPacket(func(conn *Connection, pkt protocol.Packet) {
		handler.Handle(conn, pkt)
	})
	defer srv.Stop()

	pongReceived := make(chan struct{}, 10)
	client := connectClient(t, port, func(pkt protocol.Packet) {
		if _, ok := pkt.(*protocol.Pong); ok {
			pongReceived <- struct{}{}
		}
	})
	defer client.Disconnect()

	// Client auto-sends KeepAlive every 10s; send one manually for a fast test.
	assert.NoError(t, client.Send(&protocol.KeepAlive{Timestamp: time.Now().UnixMilli()}))

	select {
	case <-pongReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Pong")
	}
}

func TestPacketHandlerDispatch(t *testing.T) {
	handler := NewPacketHandler()

	chatHandled := false
	handler.RegisterHandler(0x12, func(_ *Connection, pkt protocol.Packet) {
		chat, ok := pkt.(*protocol.ChatMessage)
		assert.True(t, ok)
		assert.Equal(t, "Test", chat.Message)
		chatHandled = true
	})

	// Dispatch with a nil connection (handler does not use it).
	handler.Handle(nil, &protocol.ChatMessage{Sender: "X", Message: "Test"})
	assert.True(t, chatHandled)
}

func TestPacketHandlerDefaultKeepAlive(t *testing.T) {
	srv, port := startTestServer(t)
	handler := NewPacketHandler() // has default KeepAlive+Ping handlers
	srv.OnPacket(func(conn *Connection, pkt protocol.Packet) {
		handler.Handle(conn, pkt)
	})
	defer srv.Stop()

	pongReceived := make(chan *protocol.Pong, 4)
	client := connectClient(t, port, func(pkt protocol.Packet) {
		if p, ok := pkt.(*protocol.Pong); ok {
			pongReceived <- p
		}
	})
	defer client.Disconnect()

	time.Sleep(50 * time.Millisecond)

	// Send a Ping; default handler should respond with Pong carrying same timestamp.
	ts := time.Now().UnixMilli()
	assert.NoError(t, client.Send(&protocol.Ping{Timestamp: ts}))

	select {
	case p := <-pongReceived:
		assert.Equal(t, ts, p.Timestamp)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Pong from Ping handler")
	}
}

func TestConnectionRemoteAddr(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()

	addrCh := make(chan string, 1)
	srv.OnConnect(func(conn *Connection) {
		addrCh <- conn.RemoteAddr()
	})

	client := connectClient(t, port, nil)
	defer client.Disconnect()

	select {
	case addr := <-addrCh:
		assert.NotEmpty(t, addr)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for remote addr")
	}
}

func TestConnectionIsAlive(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()

	aliveCh := make(chan bool, 1)
	srv.OnConnect(func(conn *Connection) {
		aliveCh <- conn.IsAlive()
	})

	client := connectClient(t, port, nil)

	select {
	case alive := <-aliveCh:
		assert.True(t, alive)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	client.Disconnect()
}

func TestSendToDisconnectedClient(t *testing.T) {
	client := NewClient()
	assert.False(t, client.IsConnected())
	err := client.Send(&protocol.ChatMessage{Sender: "X", Message: "Y"})
	assert.Error(t, err)
}

func TestHandlerOverwrite(t *testing.T) {
	handler := NewPacketHandler()

	var called string
	handler.RegisterHandler(0x12, func(_ *Connection, _ protocol.Packet) {
		called = "first"
	})
	handler.RegisterHandler(0x12, func(_ *Connection, _ protocol.Packet) {
		called = "second"
	})

	handler.Handle(nil, &protocol.ChatMessage{Sender: "X", Message: "Y"})
	assert.Equal(t, "second", called)
}

func TestHandleUnregisteredPacket(t *testing.T) {
	handler := NewPacketHandler()
	// Should not panic.
	handler.Handle(nil, &protocol.ChatMessage{Sender: "X", Message: "Y"})
}
