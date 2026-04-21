package tick

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// pistonMaxBlocks is the maximum number of blocks a piston can push.
const pistonMaxBlocks = 12

// pistonTickDelay is the number of game ticks before a piston extends or
// retracts. Used by callers when scheduling piston ticks.
const pistonTickDelay = 2

// CanPistonPush checks whether a piston at pos facing dir can push up to
// maxBlocks blocks. Returns false if any block in the chain is immovable
// or the chain exceeds maxBlocks.
func CanPistonPush(w BlockAccess, pos mcmath.BlockPos, dir mcmath.Direction, maxBlocks int) bool {
	current := pos.Neighbors()[dir]
	for i := 0; i < maxBlocks; i++ {
		id := w.GetBlock(current)
		if id == block.Air {
			return true
		}
		if block.IsImmovable(id) {
			return false
		}
		current = current.Neighbors()[dir]
	}
	// Check if the block after the chain is air (room to push into).
	return w.GetBlock(current) == block.Air
}

// PistonExtend pushes blocks in front of a piston and places a piston head.
// It collects the chain of blocks, then moves them from the far end toward
// the piston so blocks are not overwritten.
func PistonExtend(w BlockAccess, pos mcmath.BlockPos, dir mcmath.Direction) {
	if !CanPistonPush(w, pos, dir, pistonMaxBlocks) {
		return
	}

	// Collect the chain of blocks to push.
	var chain []mcmath.BlockPos
	current := pos.Neighbors()[dir]
	for {
		id := w.GetBlock(current)
		if id == block.Air {
			break
		}
		chain = append(chain, current)
		current = current.Neighbors()[dir]
	}

	// Move blocks from the far end toward the piston.
	for i := len(chain) - 1; i >= 0; i-- {
		src := chain[i]
		dst := src.Neighbors()[dir]
		w.SetBlock(dst, w.GetBlock(src))
	}

	// Place piston head at the extension position.
	headPos := pos.Neighbors()[dir]
	w.SetBlock(headPos, block.PistonHead)

	// Mark piston as extended.
	pistonID := w.GetBlock(pos)
	w.SetBlock(pos, block.WithPistonExtended(pistonID, true))
}

// PistonRetract removes the piston head and optionally pulls the adjacent
// block back (for sticky pistons).
func PistonRetract(w BlockAccess, pos mcmath.BlockPos, dir mcmath.Direction, sticky bool) {
	headPos := pos.Neighbors()[dir]

	// Only retract if there is actually a piston head in front.
	if w.GetBlock(headPos) != block.PistonHead {
		return
	}

	if sticky {
		// The block beyond the piston head should be pulled back.
		pullPos := headPos.Neighbors()[dir]
		pullID := w.GetBlock(pullPos)
		if pullID != block.Air && !block.IsImmovable(pullID) {
			w.SetBlock(headPos, pullID)
			w.SetBlock(pullPos, block.Air)
		} else {
			w.SetBlock(headPos, block.Air)
		}
	} else {
		w.SetBlock(headPos, block.Air)
	}

	// Mark piston as retracted.
	pistonID := w.GetBlock(pos)
	w.SetBlock(pos, block.WithPistonExtended(pistonID, false))
}

// RegisterPistonHandlers registers scheduled tick handlers for pistons.
// The handlers fire after a 2-tick delay when the piston's powered state
// changes.
func RegisterPistonHandlers(r *HandlerRegistry) {
	r.Register(block.Piston, makePistonTickHandler(false))
	r.Register(block.StickyPiston, makePistonTickHandler(true))
}

// makePistonTickHandler returns a tick handler that toggles a piston between
// extended and retracted. The sticky parameter controls whether the piston
// pulls an adjacent block on retract.
// NOTE: Direction is hardcoded to North because BlockAccess provides only
// uint16 block IDs without orientation state. A full implementation would
// read orientation from BlockState.
func makePistonTickHandler(sticky bool) TickHandler {
	return func(w BlockAccess, pos mcmath.BlockPos) {
		id := w.GetBlock(pos)
		if block.IsPistonExtended(id) {
			PistonRetract(w, pos, mcmath.North, sticky)
		} else {
			PistonExtend(w, pos, mcmath.North)
		}
	}
}
