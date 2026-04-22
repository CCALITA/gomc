package world

import (
	"sync"
	"sync/atomic"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

const (
	defaultLoadRadius = 8
	workerCount       = 4
	unloadPadding     = 2
)

// chunkRequest is sent to worker goroutines to request chunk generation.
type chunkRequest struct {
	pos mcmath.ChunkPos
}

// ChunkLoader manages asynchronous chunk loading around a moving center position.
type ChunkLoader struct {
	world         *World
	LoadRadius    int
	center        mcmath.ChunkPos
	centerMu      sync.Mutex
	pending       atomic.Int64
	loaded        atomic.Int64
	requests      chan chunkRequest
	stopOnce      sync.Once
	stop          chan struct{}
	workersDone   sync.WaitGroup
	OnChunkLoaded func(pos mcmath.ChunkPos)
}

// NewChunkLoader creates a ChunkLoader that will load chunks from the given world.
func NewChunkLoader(w *World) *ChunkLoader {
	cl := &ChunkLoader{
		world:      w,
		LoadRadius: defaultLoadRadius,
		requests:   make(chan chunkRequest, 256),
		stop:       make(chan struct{}),
	}

	cl.workersDone.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go cl.worker()
	}

	return cl
}

// worker processes chunk generation requests from the channel.
func (cl *ChunkLoader) worker() {
	defer cl.workersDone.Done()
	for {
		select {
		case req, ok := <-cl.requests:
			if !ok {
				return
			}
			cl.world.LoadChunk(req.pos)
			cl.pending.Add(-1)
			cl.loaded.Add(1)
			if cl.OnChunkLoaded != nil {
				cl.OnChunkLoaded(req.pos)
			}
		case <-cl.stop:
			return
		}
	}
}

// Update loads chunks within LoadRadius of centerChunkPos and unloads chunks
// outside LoadRadius + unloadPadding.
func (cl *ChunkLoader) Update(centerChunkPos mcmath.ChunkPos) {
	cl.centerMu.Lock()
	cl.center = centerChunkPos
	cl.centerMu.Unlock()

	radius := int32(cl.LoadRadius)
	unloadRadius := int32(cl.LoadRadius + unloadPadding)

	// Request loading for chunks within radius.
	for dx := -radius; dx <= radius; dx++ {
		for dz := -radius; dz <= radius; dz++ {
			cp := mcmath.ChunkPos{
				X: centerChunkPos.X + dx,
				Z: centerChunkPos.Z + dz,
			}
			if centerChunkPos.Distance(cp) > float64(radius) {
				continue
			}
			if cl.world.GetChunk(cp) != nil {
				continue
			}
			cl.pending.Add(1)
			select {
			case cl.requests <- chunkRequest{pos: cp}:
			case <-cl.stop:
				cl.pending.Add(-1)
				return
			}
		}
	}

	// Unload chunks outside radius + padding.
	for _, cp := range cl.world.LoadedChunkPositions() {
		if centerChunkPos.Distance(cp) > float64(unloadRadius) {
			cl.world.UnloadChunk(cp)
		}
	}
}

// PendingLoads returns the number of chunks currently queued for generation.
func (cl *ChunkLoader) PendingLoads() int {
	v := cl.pending.Load()
	if v < 0 {
		return 0
	}
	return int(v)
}

// LoadedCount returns the total number of chunks that have been loaded by this loader.
func (cl *ChunkLoader) LoadedCount() int {
	return int(cl.loaded.Load())
}

// Stop shuts down the worker pool. It should be called when the loader is no
// longer needed. It is safe to call multiple times.
func (cl *ChunkLoader) Stop() {
	cl.stopOnce.Do(func() {
		close(cl.stop)
		cl.workersDone.Wait()
	})
}
