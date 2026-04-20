package network

import (
	"sync"
	"time"

	"github.com/fanxiyao/gomc/internal/protocol"
)

// EntitySnapshot captures the current state of an entity for synchronization.
type EntitySnapshot struct {
	EntityID uint32
	X        float32
	Y        float32
	Z        float32
	VelX     float32
	VelY     float32
	VelZ     float32
}

// playerState holds the last known position and orientation for a player.
type playerState struct {
	X     float32
	Y     float32
	Z     float32
	Yaw   float32
	Pitch float32
}

// playerSyncEntry pairs a connection with its entity ID and last known state
// for use during the sync tick broadcast.
type playerSyncEntry struct {
	conn     *Connection
	entityID uint32
	state    playerState
}

// EntitySyncManager tracks entity-to-client mapping, last known positions,
// and broadcasts entity updates to connected clients.
type EntitySyncManager struct {
	server *Server

	mu             sync.RWMutex
	lastPositions  map[uint32]EntitySnapshot // entityID -> last synced snapshot
	playerStates   map[*Connection]playerState
	playerEntities map[*Connection]uint32 // connection -> entityID for player entities

	ticker *time.Ticker
	done   chan struct{}
}

// NewEntitySyncManager creates an EntitySyncManager that broadcasts entity
// updates through the given server.
func NewEntitySyncManager(server *Server) *EntitySyncManager {
	return &EntitySyncManager{
		server:         server,
		lastPositions:  make(map[uint32]EntitySnapshot),
		playerStates:   make(map[*Connection]playerState),
		playerEntities: make(map[*Connection]uint32),
		done:           make(chan struct{}),
	}
}

// Start begins the entity sync tick loop at the given interval (e.g., 50ms).
func (m *EntitySyncManager) Start(interval time.Duration) {
	m.ticker = time.NewTicker(interval)
	go m.tickLoop()
}

// Stop halts the sync tick loop.
func (m *EntitySyncManager) Stop() {
	close(m.done)
	if m.ticker != nil {
		m.ticker.Stop()
	}
}

// RegisterPlayer associates a connection with a player entity ID.
func (m *EntitySyncManager) RegisterPlayer(conn *Connection, entityID uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playerEntities[conn] = entityID
}

// UnregisterPlayer removes a connection's player association.
func (m *EntitySyncManager) UnregisterPlayer(conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.playerEntities, conn)
	delete(m.playerStates, conn)
}

// BroadcastEntityUpdates sends MoveEntity packets to all connected clients
// for entities that have moved since the last sync.
func (m *EntitySyncManager) BroadcastEntityUpdates(entities []EntitySnapshot) {
	var changed []*protocol.MoveEntity

	m.mu.Lock()
	for _, snap := range entities {
		last, exists := m.lastPositions[snap.EntityID]
		if exists && last.X == snap.X && last.Y == snap.Y && last.Z == snap.Z &&
			last.VelX == snap.VelX && last.VelY == snap.VelY && last.VelZ == snap.VelZ {
			continue // no change
		}

		m.lastPositions[snap.EntityID] = snap

		changed = append(changed, &protocol.MoveEntity{
			EntityID: snap.EntityID,
			X:        snap.X,
			Y:        snap.Y,
			Z:        snap.Z,
			VelX:     snap.VelX,
			VelY:     snap.VelY,
			VelZ:     snap.VelZ,
		})
	}
	m.mu.Unlock()

	for _, pkt := range changed {
		m.server.Broadcast(pkt)
	}
}

// HandlePlayerPosition updates the server-side player position and relays
// it to all other clients as a MoveEntity packet. The sender is excluded
// from the broadcast.
func (m *EntitySyncManager) HandlePlayerPosition(conn *Connection, x, y, z, yaw, pitch float32) {
	m.mu.Lock()
	state := m.playerStates[conn]
	state.X = x
	state.Y = y
	state.Z = z
	state.Yaw = yaw
	state.Pitch = pitch
	m.playerStates[conn] = state

	entityID, hasEntity := m.playerEntities[conn]
	m.mu.Unlock()

	if !hasEntity {
		return
	}

	pkt := &protocol.MoveEntity{
		EntityID: entityID,
		X:        x,
		Y:        y,
		Z:        z,
	}
	m.server.BroadcastExcept(pkt, conn)
}

// GetPlayerState returns the last known position for a connection.
// Returns false if the connection has no recorded state.
func (m *EntitySyncManager) GetPlayerState(conn *Connection) (playerState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.playerStates[conn]
	return s, ok
}

func (m *EntitySyncManager) tickLoop() {
	for {
		select {
		case <-m.done:
			return
		case <-m.ticker.C:
			m.mu.RLock()
			entries := make([]playerSyncEntry, 0, len(m.playerEntities))
			for conn, eid := range m.playerEntities {
				if s, ok := m.playerStates[conn]; ok {
					entries = append(entries, playerSyncEntry{
						conn:     conn,
						entityID: eid,
						state:    s,
					})
				}
			}
			m.mu.RUnlock()

			for _, e := range entries {
				pkt := &protocol.MoveEntity{
					EntityID: e.entityID,
					X:        e.state.X,
					Y:        e.state.Y,
					Z:        e.state.Z,
				}
				m.server.BroadcastExcept(pkt, e.conn)
			}
		}
	}
}
