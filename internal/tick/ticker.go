// Package tick implements the random tick dispatcher and scheduled tick queue
// for block updates, matching vanilla Minecraft's tick behavior.
package tick

import (
	"container/heap"
	"math/rand/v2"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// randomTicksPerSection is the number of random blocks picked per section
// per tick, matching vanilla Minecraft's default randomTickSpeed of 3.
const randomTicksPerSection = 3

// BlockAccess defines the minimal interface for reading and writing blocks
// in the world. Defined at the point of use per Go idiom.
type BlockAccess interface {
	GetBlock(pos mcmath.BlockPos) uint16
	SetBlock(pos mcmath.BlockPos, id uint16)
}

// Ticker drives random ticks and scheduled ticks for a world.
type Ticker struct {
	world     BlockAccess
	registry  *HandlerRegistry
	scheduled scheduledQueue
	rng       *rand.Rand
}

// NewTicker creates a Ticker that operates on the given world.
func NewTicker(world BlockAccess, registry *HandlerRegistry) *Ticker {
	return &Ticker{
		world:    world,
		registry: registry,
		rng:      rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

// NewTickerWithSeed creates a Ticker with a deterministic RNG seed for testing.
func NewTickerWithSeed(world BlockAccess, registry *HandlerRegistry, seed uint64) *Ticker {
	return &Ticker{
		world:    world,
		registry: registry,
		rng:      rand.New(rand.NewPCG(seed, seed)),
	}
}

// RandomTick picks 3 random blocks per non-nil section in the given chunk
// and invokes any registered tick handler for the block type found.
func (t *Ticker) RandomTick(chunkX, chunkZ int32) {
	cp := mcmath.ChunkPos{X: chunkX, Z: chunkZ}
	numSections := mcmath.ChunkHeight / mcmath.SectionHeight

	for section := 0; section < numSections; section++ {
		baseY := int32(section * mcmath.SectionHeight)

		for i := 0; i < randomTicksPerSection; i++ {
			lx := int32(t.rng.IntN(mcmath.ChunkSize))
			ly := int32(t.rng.IntN(mcmath.SectionHeight))
			lz := int32(t.rng.IntN(mcmath.ChunkSize))

			pos := cp.BlockPosAt(lx, baseY+ly, lz)
			blockID := t.world.GetBlock(pos)

			handler, ok := t.registry.Get(blockID)
			if ok {
				handler(t.world, pos)
			}
		}
	}
}

// Schedule enqueues a block tick to fire after the given delay (in game ticks).
func (t *Ticker) Schedule(pos mcmath.BlockPos, blockID uint16, delay int, currentTick int64) {
	entry := scheduledEntry{
		pos:      pos,
		blockID:  blockID,
		fireTick: currentTick + int64(delay),
	}
	heap.Push(&t.scheduled, entry)
}

// ProcessScheduledTicks fires all scheduled ticks whose fire time is at or
// before currentTick.
func (t *Ticker) ProcessScheduledTicks(currentTick int64) {
	for t.scheduled.Len() > 0 {
		top := t.scheduled[0]
		if top.fireTick > currentTick {
			break
		}
		heap.Pop(&t.scheduled)

		handler, ok := t.registry.Get(top.blockID)
		if ok {
			handler(t.world, top.pos)
		}
	}
}

// ScheduledLen returns the number of pending scheduled ticks. Useful for tests.
func (t *Ticker) ScheduledLen() int {
	return t.scheduled.Len()
}

// --- priority queue for scheduled ticks ---

type scheduledEntry struct {
	pos      mcmath.BlockPos
	blockID  uint16
	fireTick int64
}

// scheduledQueue implements heap.Interface, ordered by fireTick ascending.
type scheduledQueue []scheduledEntry

func (q scheduledQueue) Len() int            { return len(q) }
func (q scheduledQueue) Less(i, j int) bool  { return q[i].fireTick < q[j].fireTick }
func (q scheduledQueue) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }

func (q *scheduledQueue) Push(x any) {
	*q = append(*q, x.(scheduledEntry))
}

func (q *scheduledQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}
