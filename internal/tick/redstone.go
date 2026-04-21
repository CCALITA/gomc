package tick

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// buttonTimeoutTicks is the number of ticks a stone button stays active.
const buttonTimeoutTicks = 20

// PropagateRedstone performs a BFS from pos outward through redstone wire,
// setting each wire's power to max(adjacent_power - 1, 0). Only horizontal
// neighbors (4 directions) plus up/down are considered for wire connectivity.
func PropagateRedstone(w BlockAccess, pos mcmath.BlockPos) {
	srcID := w.GetBlock(pos)
	srcPower := sourcepower(srcID)

	type entry struct {
		pos   mcmath.BlockPos
		power int
	}

	visited := map[mcmath.BlockPos]bool{pos: true}
	queue := []entry{}

	for _, n := range pos.Neighbors() {
		nid := w.GetBlock(n)
		if block.IsRedstoneWire(nid) {
			queue = append(queue, entry{pos: n, power: srcPower - 1})
			visited[n] = true
		}
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if cur.power < 0 {
			cur.power = 0
		}

		// Read current wire power; only update if we bring more power.
		curID := w.GetBlock(cur.pos)
		existing := block.RedstonePowerLevel(curID)
		if cur.power <= existing {
			continue
		}

		w.SetBlock(cur.pos, block.WithRedstonePower(block.RedstoneWire, cur.power))

		if cur.power <= 1 {
			continue
		}

		for _, n := range cur.pos.Neighbors() {
			if visited[n] {
				continue
			}
			nid := w.GetBlock(n)
			if block.IsRedstoneWire(nid) {
				visited[n] = true
				queue = append(queue, entry{pos: n, power: cur.power - 1})
			}
		}
	}
}

// ResetRedstone sets all connected wire to power 0, then re-propagates from
// any remaining power sources. Used when a source is turned off.
func ResetRedstone(w BlockAccess, pos mcmath.BlockPos) {
	visited := map[mcmath.BlockPos]bool{pos: true}
	wirePositions := []mcmath.BlockPos{}
	queue := []mcmath.BlockPos{}

	for _, n := range pos.Neighbors() {
		if block.IsRedstoneWire(w.GetBlock(n)) {
			queue = append(queue, n)
			visited[n] = true
		}
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		wirePositions = append(wirePositions, cur)

		for _, n := range cur.Neighbors() {
			if visited[n] {
				continue
			}
			if block.IsRedstoneWire(w.GetBlock(n)) {
				visited[n] = true
				queue = append(queue, n)
			}
		}
	}

	// Zero out all wire, then re-propagate from remaining adjacent sources.
	for _, wp := range wirePositions {
		w.SetBlock(wp, block.RedstoneWire)
	}

	repropagate := map[mcmath.BlockPos]bool{}
	for _, wp := range wirePositions {
		for _, n := range wp.Neighbors() {
			if repropagate[n] {
				continue
			}
			nid := w.GetBlock(n)
			if block.IsPowerSource(nid) {
				repropagate[n] = true
			}
		}
	}
	for src := range repropagate {
		PropagateRedstone(w, src)
	}
}

// sourcepower returns the power level a block emits as a source.
func sourcepower(id uint16) int {
	if block.IsPowerSource(id) {
		return block.MaxRedstonePower
	}
	return 0
}

// RegisterRedstoneHandlers registers tick handlers for redstone components.
// Lever: toggles power on right-click (handled externally via interaction).
// Button: scheduled tick deactivates after buttonTimeoutTicks.
// Torch: re-propagates on tick.
func RegisterRedstoneHandlers(r *HandlerRegistry) {
	r.Register(block.StoneButton, buttonTimeoutHandler)
	r.Register(block.RedstoneTorch, torchUpdateHandler)
}

// buttonTimeoutHandler deactivates a stone button by setting its power to 0
// and resetting connected wire.
func buttonTimeoutHandler(w BlockAccess, pos mcmath.BlockPos) {
	id := w.GetBlock(pos)
	if block.RedstonePowerLevel(id) == 0 {
		return
	}
	w.SetBlock(pos, block.StoneButton) // power 0
	ResetRedstone(w, pos)
}

// torchUpdateHandler checks if the block below the torch is powered;
// if so, the torch turns off (inverted behavior). Otherwise it stays on
// and propagates power.
func torchUpdateHandler(w BlockAccess, pos mcmath.BlockPos) {
	below := pos.Below()
	belowID := w.GetBlock(below)

	// Torch inversion: if the block supporting the torch receives power,
	// the torch turns off.
	powered := block.IsRedstone(belowID) && block.RedstonePowerLevel(belowID) > 0
	if powered {
		// Turn torch off (power 0) and reset downstream wire.
		w.SetBlock(pos, block.RedstoneTorch) // power 0 = off
		ResetRedstone(w, pos)
	} else {
		// Torch is on — ensure it propagates.
		w.SetBlock(pos, block.WithRedstonePower(block.RedstoneTorch, block.MaxRedstonePower))
		PropagateRedstone(w, pos)
	}
}
