package ecs

import "sync"

// Entity is a unique identifier composed of an index (lower 32 bits) and a
// generation (upper 32 bits). The generation prevents aliasing after an entity
// is destroyed and its index slot is recycled.
type Entity = uint64

const (
	indexBits = 32
	indexMask = (1 << indexBits) - 1
	genShift = indexBits
)

// entityIndex returns the index portion of an Entity.
func entityIndex(e Entity) uint32 { return uint32(e & indexMask) }

// entityGen returns the generation portion of an Entity.
func entityGen(e Entity) uint32 { return uint32(e >> genShift) }

// makeEntity constructs an Entity from an index and generation.
func makeEntity(index, gen uint32) Entity {
	return Entity(gen)<<genShift | Entity(index)
}

// EntityManager creates, destroys, and tracks entities.
type EntityManager struct {
	mu sync.Mutex

	// generation[i] holds the current generation for slot i.
	generation []uint32

	// alive[i] is true when slot i is currently in use.
	alive []bool

	// freeList stores recycled slot indices.
	freeList []uint32
}

// NewEntityManager returns a ready-to-use EntityManager.
func NewEntityManager() *EntityManager {
	return &EntityManager{}
}

// Create allocates a new Entity, reusing a recycled slot when possible.
func (em *EntityManager) Create() Entity {
	em.mu.Lock()
	defer em.mu.Unlock()

	if len(em.freeList) > 0 {
		idx := em.freeList[len(em.freeList)-1]
		em.freeList = em.freeList[:len(em.freeList)-1]
		em.alive[idx] = true
		return makeEntity(idx, em.generation[idx])
	}

	idx := uint32(len(em.generation))
	em.generation = append(em.generation, 0)
	em.alive = append(em.alive, true)
	return makeEntity(idx, 0)
}

// Destroy marks an Entity as dead and recycles its index slot. The
// generation for the slot is incremented so that stale handles are
// detected.
func (em *EntityManager) Destroy(e Entity) {
	em.mu.Lock()
	defer em.mu.Unlock()

	idx := entityIndex(e)
	gen := entityGen(e)

	if int(idx) >= len(em.generation) {
		return
	}
	if em.generation[idx] != gen || !em.alive[idx] {
		return
	}

	em.alive[idx] = false
	em.generation[idx]++
	em.freeList = append(em.freeList, idx)
}

// Alive reports whether e refers to a currently living entity.
func (em *EntityManager) Alive(e Entity) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	idx := entityIndex(e)
	gen := entityGen(e)

	if int(idx) >= len(em.generation) {
		return false
	}
	return em.generation[idx] == gen && em.alive[idx]
}
