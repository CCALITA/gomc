package world

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

const (
	// maxFluidLevel is the maximum flow level before a fluid stops spreading.
	maxFluidLevel = 7

	// waterSpreadInterval is how many ticks between water spread updates.
	waterSpreadInterval = 5

	// lavaSpreadInterval is how many ticks between lava spread updates.
	lavaSpreadInterval = 30
)

// fluidUpdate represents a pending fluid update at a position.
type fluidUpdate struct {
	pos  mcmath.BlockPos
	tick int
}

// FluidSimulator handles water and lava flow mechanics using a BFS queue.
type FluidSimulator struct {
	world   *World
	queue   []fluidUpdate
	tick    int
}

// NewFluidSimulator creates a FluidSimulator bound to the given world.
func NewFluidSimulator(world *World) *FluidSimulator {
	return &FluidSimulator{
		world: world,
	}
}

// Update processes one tick of fluid simulation, spreading and draining fluids
// whose scheduled tick has arrived.
func (fs *FluidSimulator) Update() {
	fs.tick++

	var remaining []fluidUpdate
	var ready []fluidUpdate

	for _, u := range fs.queue {
		if u.tick <= fs.tick {
			ready = append(ready, u)
		} else {
			remaining = append(remaining, u)
		}
	}
	fs.queue = remaining

	for _, u := range ready {
		fs.processUpdate(u.pos)
	}
}

// OnBlockChange is called when a block changes. It schedules fluid updates as needed.
func (fs *FluidSimulator) OnBlockChange(pos mcmath.BlockPos, oldID, newID uint16) {
	// Source placed: schedule spread.
	if newID == block.Water {
		fs.scheduleUpdate(pos, waterSpreadInterval)
	}
	if newID == block.Lava {
		fs.scheduleUpdate(pos, lavaSpreadInterval)
	}

	// Source removed: schedule drain of connected flowing blocks.
	if (oldID == block.Water || oldID == block.Lava) && !block.IsFluid(newID) {
		fs.drainConnected(pos, oldID)
	}

	// Check for lava/water interactions at the changed position.
	fs.checkInteractions(pos)
}

// scheduleUpdate adds a fluid update to the BFS queue.
func (fs *FluidSimulator) scheduleUpdate(pos mcmath.BlockPos, delay int) {
	fs.queue = append(fs.queue, fluidUpdate{
		pos:  pos,
		tick: fs.tick + delay,
	})
}

// processUpdate handles a single fluid update at a position by spreading the fluid.
func (fs *FluidSimulator) processUpdate(pos mcmath.BlockPos) {
	id := fs.world.GetBlock(pos)
	if !block.IsFluid(id) {
		return
	}

	isSource := id == block.Water || id == block.Lava
	level := block.FluidLevel(id)

	// Sources spread at level 1; flowing blocks spread at level+1.
	var nextLevel int
	if isSource {
		nextLevel = 1
	} else {
		nextLevel = level + 1
	}

	if nextLevel > maxFluidLevel {
		return
	}

	isWater := block.IsWater(id)
	interval := lavaSpreadInterval
	if isWater {
		interval = waterSpreadInterval
	}

	// Try to flow down first.
	below := pos.Below()
	belowID := fs.world.GetBlock(below)
	if below.Y >= 0 && canFluidReplace(belowID) {
		// Flowing down creates a level 1 block (nearly full).
		var flowID uint16
		if isWater {
			flowID = block.FlowingWaterLevel(1)
		} else {
			flowID = block.FlowingLavaLevel(1)
		}
		fs.world.SetBlock(below, flowID)
		fs.checkInteractions(below)
		fs.scheduleUpdate(below, interval)
	}

	// Spread horizontally.
	horizontalNeighbors := [4]mcmath.BlockPos{
		{X: pos.X + 1, Y: pos.Y, Z: pos.Z},
		{X: pos.X - 1, Y: pos.Y, Z: pos.Z},
		{X: pos.X, Y: pos.Y, Z: pos.Z + 1},
		{X: pos.X, Y: pos.Y, Z: pos.Z - 1},
	}

	for _, neighbor := range horizontalNeighbors {
		neighborID := fs.world.GetBlock(neighbor)
		if !canFluidReplace(neighborID) {
			continue
		}

		var flowID uint16
		if isWater {
			flowID = block.FlowingWaterLevel(nextLevel)
		} else {
			flowID = block.FlowingLavaLevel(nextLevel)
		}

		fs.world.SetBlock(neighbor, flowID)
		fs.checkInteractions(neighbor)
		if nextLevel < maxFluidLevel {
			fs.scheduleUpdate(neighbor, interval)
		}
	}
}

// canFluidReplace reports whether a fluid can replace the block at a given position.
func canFluidReplace(id uint16) bool {
	return id == block.Air || (!block.IsSolid(id) && !block.IsFluid(id))
}

// drainConnected removes all flowing blocks connected to a removed source
// using BFS. This is a simplified drain: it removes all connected flowing
// blocks of the same fluid type.
func (fs *FluidSimulator) drainConnected(sourcePos mcmath.BlockPos, sourceID uint16) {
	isWater := sourceID == block.Water

	visited := map[mcmath.BlockPos]struct{}{sourcePos: {}}
	queue := []mcmath.BlockPos{sourcePos}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range current.Neighbors() {
			if _, seen := visited[neighbor]; seen {
				continue
			}
			visited[neighbor] = struct{}{}

			nID := fs.world.GetBlock(neighbor)
			base := block.BaseID(nID)

			var isMatchingFlow bool
			if isWater {
				isMatchingFlow = base == block.FlowingWater
			} else {
				isMatchingFlow = base == block.FlowingLava
			}

			if isMatchingFlow {
				fs.world.SetBlock(neighbor, block.Air)
				queue = append(queue, neighbor)
			}
		}
	}
}

// checkInteractions handles lava/water interactions at the given position.
func (fs *FluidSimulator) checkInteractions(pos mcmath.BlockPos) {
	id := fs.world.GetBlock(pos)
	if !block.IsFluid(id) {
		return
	}

	for _, neighbor := range pos.Neighbors() {
		nID := fs.world.GetBlock(neighbor)
		if !block.IsFluid(nID) {
			continue
		}

		waterMeetsLava := block.IsWater(id) && block.IsLava(nID)
		lavaMeetsWater := block.IsLava(id) && block.IsWater(nID)
		if waterMeetsLava || lavaMeetsWater {
			fs.resolveInteraction(pos, id, neighbor, nID)
			return
		}
	}
}

// resolveInteraction handles the result of water and lava meeting.
func (fs *FluidSimulator) resolveInteraction(posA mcmath.BlockPos, idA uint16, posB mcmath.BlockPos, idB uint16) {
	// Determine which is lava and which is water.
	var lavaPos mcmath.BlockPos
	var lavaID uint16
	if block.IsLava(idA) {
		lavaPos = posA
		lavaID = idA
	} else {
		lavaPos = posB
		lavaID = idB
	}

	// Lava source + water => Obsidian (replaces the lava).
	// Flowing lava + water => Cobblestone (replaces the flowing lava).
	if lavaID == block.Lava {
		fs.world.SetBlock(lavaPos, block.Obsidian)
	} else {
		fs.world.SetBlock(lavaPos, block.Cobblestone)
	}
}
