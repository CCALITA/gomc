package tick

import (
	"sync"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// TickHandler is a function invoked when a block receives a random or
// scheduled tick.
type TickHandler func(world BlockAccess, pos mcmath.BlockPos)

// HandlerRegistry maps block IDs to their tick handlers.
type HandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[uint16]TickHandler
}

// NewHandlerRegistry creates an empty handler registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[uint16]TickHandler),
	}
}

// Register associates a tick handler with the given block ID.
func (r *HandlerRegistry) Register(blockID uint16, handler TickHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[blockID] = handler
}

// Get returns the handler for the given block ID, if any.
func (r *HandlerRegistry) Get(blockID uint16) (TickHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[blockID]
	return h, ok
}

// RegisterDefaults registers the standard vanilla-style tick handlers:
// grass spread and leaf decay.
func RegisterDefaults(r *HandlerRegistry) {
	r.Register(block.Dirt, grassSpreadHandler)
	r.Register(block.OakLeaves, leafDecayHandler)
}

// grassSpreadHandler converts a dirt block to grass if an adjacent block is
// grass and the block above the dirt has a light filter of 0 (transparent to
// light, approximating the vanilla "light level >= 9 above dirt" check).
func grassSpreadHandler(w BlockAccess, pos mcmath.BlockPos) {
	above := pos.Above()
	if above.Y >= mcmath.ChunkHeight {
		return
	}

	aboveBlock := w.GetBlock(above)
	props := block.GetProperties(aboveBlock)
	if props.LightFilter > 0 && aboveBlock != block.Air {
		return
	}

	if hasAdjacentGrass(w, pos) {
		w.SetBlock(pos, block.Grass)
	}
}

// hasAdjacentGrass checks whether any of the six neighbors is a grass block.
func hasAdjacentGrass(w BlockAccess, pos mcmath.BlockPos) bool {
	for _, neighbor := range pos.Neighbors() {
		if w.GetBlock(neighbor) == block.Grass {
			return true
		}
	}
	return false
}

// leafDecayMaxDistance is the maximum taxicab distance from a leaf to a
// supporting oak log before the leaf decays.
const leafDecayMaxDistance = 6

// leafDecayHandler removes an oak leaf block if there is no oak log within
// leafDecayMaxDistance taxicab distance.
func leafDecayHandler(w BlockAccess, pos mcmath.BlockPos) {
	if hasNearbyLog(w, pos, leafDecayMaxDistance) {
		return
	}
	w.SetBlock(pos, block.Air)
}

// hasNearbyLog performs a bounded search for an oak log within the given
// taxicab distance from pos. Loop bounds are tightened so only positions
// within the taxicab diamond are visited.
func hasNearbyLog(w BlockAccess, center mcmath.BlockPos, maxDist int32) bool {
	for dx := -maxDist; dx <= maxDist; dx++ {
		remAfterX := maxDist - mcmath.Abs(dx)
		for dy := -remAfterX; dy <= remAfterX; dy++ {
			remAfterXY := remAfterX - mcmath.Abs(dy)
			for dz := -remAfterXY; dz <= remAfterXY; dz++ {
				if w.GetBlock(center.Offset(dx, dy, dz)) == block.OakLog {
					return true
				}
			}
		}
	}
	return false
}

