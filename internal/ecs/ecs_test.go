package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Entity tests
// ---------------------------------------------------------------------------

func TestEntityCreateAndAlive(t *testing.T) {
	em := NewEntityManager()

	e1 := em.Create()
	e2 := em.Create()

	assert.True(t, em.Alive(e1))
	assert.True(t, em.Alive(e2))
	assert.NotEqual(t, e1, e2, "two entities must have different IDs")
}

func TestEntityDestroy(t *testing.T) {
	em := NewEntityManager()
	e := em.Create()

	em.Destroy(e)
	assert.False(t, em.Alive(e), "destroyed entity must not be alive")
}

func TestEntityDestroyIdempotent(t *testing.T) {
	em := NewEntityManager()
	e := em.Create()
	em.Destroy(e)
	em.Destroy(e) // second destroy should be a no-op
	assert.False(t, em.Alive(e))
}

func TestEntityGenerationReuse(t *testing.T) {
	em := NewEntityManager()
	e1 := em.Create()
	idx1 := entityIndex(e1)
	gen1 := entityGen(e1)

	em.Destroy(e1)

	e2 := em.Create()
	idx2 := entityIndex(e2)
	gen2 := entityGen(e2)

	assert.Equal(t, idx1, idx2, "index slot should be recycled")
	assert.Equal(t, gen1+1, gen2, "generation must be incremented after reuse")
	assert.NotEqual(t, e1, e2, "reused slot must produce a different Entity value")

	assert.False(t, em.Alive(e1), "old generation handle must be dead")
	assert.True(t, em.Alive(e2), "new generation handle must be alive")
}

func TestEntityAliveOutOfRange(t *testing.T) {
	em := NewEntityManager()
	assert.False(t, em.Alive(makeEntity(999, 0)), "non-existent index must return false")
}

func TestEntityDestroyOutOfRange(t *testing.T) {
	em := NewEntityManager()
	em.Destroy(makeEntity(999, 0)) // should not panic
}

// ---------------------------------------------------------------------------
// ComponentStore tests
// ---------------------------------------------------------------------------

type Position struct {
	X, Y float64
}

type Velocity struct {
	DX, DY float64
}

type Health struct {
	HP int
}

func TestComponentSetGet(t *testing.T) {
	s := NewComponentStore[Position]()
	var e Entity = 1

	s.Set(e, Position{10, 20})
	p, ok := s.Get(e)

	assert.True(t, ok)
	assert.Equal(t, 10.0, p.X)
	assert.Equal(t, 20.0, p.Y)
}

func TestComponentOverwrite(t *testing.T) {
	s := NewComponentStore[Position]()
	var e Entity = 1

	s.Set(e, Position{1, 2})
	s.Set(e, Position{3, 4})

	p, ok := s.Get(e)
	assert.True(t, ok)
	assert.Equal(t, 3.0, p.X)
	assert.Equal(t, 4.0, p.Y)
	assert.Equal(t, 1, s.Len(), "overwrite must not increase length")
}

func TestComponentRemove(t *testing.T) {
	s := NewComponentStore[Position]()
	var e Entity = 1

	s.Set(e, Position{1, 2})
	removed := s.Remove(e)

	assert.True(t, removed)
	assert.False(t, s.Has(e))
	assert.Equal(t, 0, s.Len())
}

func TestComponentRemoveNonExistent(t *testing.T) {
	s := NewComponentStore[Position]()
	assert.False(t, s.Remove(42))
}

func TestComponentHas(t *testing.T) {
	s := NewComponentStore[Position]()
	var e Entity = 1

	assert.False(t, s.Has(e))
	s.Set(e, Position{})
	assert.True(t, s.Has(e))
}

func TestComponentLen(t *testing.T) {
	s := NewComponentStore[Position]()
	assert.Equal(t, 0, s.Len())

	s.Set(1, Position{})
	s.Set(2, Position{})
	assert.Equal(t, 2, s.Len())
}

func TestComponentEach(t *testing.T) {
	s := NewComponentStore[Position]()
	s.Set(1, Position{1, 0})
	s.Set(2, Position{2, 0})
	s.Set(3, Position{3, 0})

	sum := 0.0
	s.Each(func(_ Entity, p *Position) {
		sum += p.X
	})
	assert.Equal(t, 6.0, sum)
}

func TestComponentSwapRemoveCorrectness(t *testing.T) {
	s := NewComponentStore[int]()

	// Insert entities 10, 20, 30 with values 100, 200, 300.
	s.Set(Entity(10), 100)
	s.Set(Entity(20), 200)
	s.Set(Entity(30), 300)

	// Remove the middle element (entity 20). The last element (entity 30)
	// should be swapped into its slot.
	s.Remove(Entity(20))

	assert.Equal(t, 2, s.Len())
	assert.False(t, s.Has(Entity(20)))

	v10, ok := s.Get(Entity(10))
	assert.True(t, ok)
	assert.Equal(t, 100, *v10)

	v30, ok := s.Get(Entity(30))
	assert.True(t, ok)
	assert.Equal(t, 300, *v30)
}

func TestComponentSwapRemoveLast(t *testing.T) {
	s := NewComponentStore[int]()
	s.Set(Entity(1), 10)
	s.Set(Entity(2), 20)

	// Remove the last element -- no swap needed.
	s.Remove(Entity(2))

	assert.Equal(t, 1, s.Len())
	v, ok := s.Get(Entity(1))
	assert.True(t, ok)
	assert.Equal(t, 10, *v)
}

func TestComponentGetMissing(t *testing.T) {
	s := NewComponentStore[int]()
	v, ok := s.Get(Entity(999))
	assert.False(t, ok)
	assert.Nil(t, v)
}

// ---------------------------------------------------------------------------
// World tests
// ---------------------------------------------------------------------------

func TestWorldNewEntity(t *testing.T) {
	w := NewWorld()
	e1 := w.NewEntity()
	e2 := w.NewEntity()

	assert.True(t, w.Alive(e1))
	assert.True(t, w.Alive(e2))
	assert.NotEqual(t, e1, e2)
}

func TestWorldDestroyEntity(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()
	w.DestroyEntity(e)
	assert.False(t, w.Alive(e))
}

func TestGetStoreCreatesOnFirstAccess(t *testing.T) {
	w := NewWorld()
	s1 := GetStore[Position](w)
	s2 := GetStore[Position](w)

	assert.Same(t, s1, s2, "must return the same store instance")
}

func TestGetStoreDifferentTypes(t *testing.T) {
	w := NewWorld()
	sp := GetStore[Position](w)
	sv := GetStore[Velocity](w)

	// They should be independent.
	e := w.NewEntity()
	sp.Set(e, Position{1, 2})

	assert.True(t, sp.Has(e))
	assert.False(t, sv.Has(e))
}

// ---------------------------------------------------------------------------
// Query tests
// ---------------------------------------------------------------------------

func TestQuery2(t *testing.T) {
	w := NewWorld()

	e1 := w.NewEntity()
	e2 := w.NewEntity()
	e3 := w.NewEntity() // position only

	ps := GetStore[Position](w)
	vs := GetStore[Velocity](w)

	ps.Set(e1, Position{1, 0})
	vs.Set(e1, Velocity{10, 0})

	ps.Set(e2, Position{2, 0})
	vs.Set(e2, Velocity{20, 0})

	ps.Set(e3, Position{3, 0})

	var results []Entity
	Query2[Position, Velocity](w, func(e Entity, p *Position, v *Velocity) {
		p.X += v.DX
		results = append(results, e)
	})

	assert.Len(t, results, 2)
	assert.Contains(t, results, e1)
	assert.Contains(t, results, e2)

	p1, _ := ps.Get(e1)
	assert.Equal(t, 11.0, p1.X)
	p2, _ := ps.Get(e2)
	assert.Equal(t, 22.0, p2.X)
}

func TestQuery2IteratesSmallerStore(t *testing.T) {
	w := NewWorld()

	// Create 5 entities with Position, 2 with Velocity.
	entities := make([]Entity, 5)
	ps := GetStore[Position](w)
	vs := GetStore[Velocity](w)

	for i := range entities {
		entities[i] = w.NewEntity()
		ps.Set(entities[i], Position{float64(i), 0})
	}
	vs.Set(entities[0], Velocity{1, 0})
	vs.Set(entities[1], Velocity{1, 0})

	count := 0
	Query2[Position, Velocity](w, func(_ Entity, _ *Position, _ *Velocity) {
		count++
	})
	assert.Equal(t, 2, count)
}

func TestQuery3(t *testing.T) {
	w := NewWorld()

	e1 := w.NewEntity()
	e2 := w.NewEntity()
	e3 := w.NewEntity()

	ps := GetStore[Position](w)
	vs := GetStore[Velocity](w)
	hs := GetStore[Health](w)

	ps.Set(e1, Position{1, 0})
	vs.Set(e1, Velocity{1, 0})
	hs.Set(e1, Health{100})

	ps.Set(e2, Position{2, 0})
	vs.Set(e2, Velocity{2, 0})
	// e2 has no Health

	ps.Set(e3, Position{3, 0})
	vs.Set(e3, Velocity{3, 0})
	hs.Set(e3, Health{50})

	var results []Entity
	Query3[Position, Velocity, Health](w, func(e Entity, p *Position, v *Velocity, h *Health) {
		results = append(results, e)
	})

	assert.Len(t, results, 2)
	assert.Contains(t, results, e1)
	assert.Contains(t, results, e3)
}

func TestQuery3SmallestStoreSelection(t *testing.T) {
	w := NewWorld()

	// Create many Position and Velocity, few Health.
	e := w.NewEntity()
	ps := GetStore[Position](w)
	vs := GetStore[Velocity](w)
	hs := GetStore[Health](w)

	for i := 0; i < 10; i++ {
		eid := w.NewEntity()
		ps.Set(eid, Position{})
		vs.Set(eid, Velocity{})
	}
	ps.Set(e, Position{})
	vs.Set(e, Velocity{})
	hs.Set(e, Health{42})

	count := 0
	Query3[Position, Velocity, Health](w, func(_ Entity, _ *Position, _ *Velocity, _ *Health) {
		count++
	})
	assert.Equal(t, 1, count)
}

func TestQuery3IteratesBWhenSmallest(t *testing.T) {
	w := NewWorld()

	// A has 5, B has 1, C has 5 -> B is smallest
	entities := make([]Entity, 5)
	sa := GetStore[Position](w)
	sb := GetStore[Velocity](w)
	sc := GetStore[Health](w)

	for i := range entities {
		entities[i] = w.NewEntity()
		sa.Set(entities[i], Position{})
		sc.Set(entities[i], Health{})
	}
	// Only entity[2] has all three
	sb.Set(entities[2], Velocity{})

	var results []Entity
	Query3[Position, Velocity, Health](w, func(e Entity, _ *Position, _ *Velocity, _ *Health) {
		results = append(results, e)
	})
	assert.Len(t, results, 1)
	assert.Equal(t, entities[2], results[0])
}

func TestQuery3IteratesCWhenSmallest(t *testing.T) {
	w := NewWorld()

	// A has 5, B has 5, C has 1 -> C is smallest
	entities := make([]Entity, 5)
	sa := GetStore[Position](w)
	sb := GetStore[Velocity](w)
	sc := GetStore[Health](w)

	for i := range entities {
		entities[i] = w.NewEntity()
		sa.Set(entities[i], Position{})
		sb.Set(entities[i], Velocity{})
	}
	// Only entity[3] has all three
	sc.Set(entities[3], Health{})

	var results []Entity
	Query3[Position, Velocity, Health](w, func(e Entity, _ *Position, _ *Velocity, _ *Health) {
		results = append(results, e)
	})
	assert.Len(t, results, 1)
	assert.Equal(t, entities[3], results[0])
}

// ---------------------------------------------------------------------------
// System / Scheduler tests
// ---------------------------------------------------------------------------

type recordingSystem struct {
	id    int
	order *[]int
}

func (s *recordingSystem) Update(_ *World, _ float64) {
	*s.order = append(*s.order, s.id)
}

func TestSchedulerExecutionOrder(t *testing.T) {
	sched := NewScheduler()
	var order []int

	sched.Add(&recordingSystem{id: 1, order: &order})
	sched.Add(&recordingSystem{id: 2, order: &order})
	sched.Add(&recordingSystem{id: 3, order: &order})

	w := NewWorld()
	sched.Update(w, 0.016)

	assert.Equal(t, []int{1, 2, 3}, order)
}

type movementSystem struct{}

func (s *movementSystem) Update(w *World, dt float64) {
	Query2[Position, Velocity](w, func(_ Entity, p *Position, v *Velocity) {
		p.X += v.DX * dt
		p.Y += v.DY * dt
	})
}

func TestSystemIntegration(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()

	ps := GetStore[Position](w)
	vs := GetStore[Velocity](w)

	ps.Set(e, Position{0, 0})
	vs.Set(e, Velocity{100, 50})

	sched := NewScheduler()
	sched.Add(&movementSystem{})

	sched.Update(w, 1.0)

	p, _ := ps.Get(e)
	assert.Equal(t, 100.0, p.X)
	assert.Equal(t, 50.0, p.Y)
}

func TestSchedulerDtPropagation(t *testing.T) {
	var received float64
	sys := systemFunc(func(_ *World, dt float64) {
		received = dt
	})

	sched := NewScheduler()
	sched.Add(sys)
	sched.Update(NewWorld(), 0.033)

	assert.InDelta(t, 0.033, received, 1e-9)
}

// systemFunc is a helper adapter so a plain function satisfies System.
type systemFunc func(*World, float64)

func (f systemFunc) Update(w *World, dt float64) { f(w, dt) }

// ---------------------------------------------------------------------------
// Edge cases and additional coverage
// ---------------------------------------------------------------------------

func TestEntityManagerMultipleCreateDestroy(t *testing.T) {
	em := NewEntityManager()
	entities := make([]Entity, 100)
	for i := range entities {
		entities[i] = em.Create()
	}

	// Destroy odd-indexed entities.
	for i := 1; i < len(entities); i += 2 {
		em.Destroy(entities[i])
	}

	// Verify alive status.
	for i, e := range entities {
		if i%2 == 0 {
			assert.True(t, em.Alive(e), "even entity %d should be alive", i)
		} else {
			assert.False(t, em.Alive(e), "odd entity %d should be dead", i)
		}
	}

	// Create new entities -- should reuse slots.
	for i := 0; i < 50; i++ {
		ne := em.Create()
		assert.True(t, em.Alive(ne))
	}
}

func TestComponentEachMutate(t *testing.T) {
	s := NewComponentStore[int]()
	s.Set(Entity(1), 10)
	s.Set(Entity(2), 20)

	s.Each(func(_ Entity, v *int) {
		*v *= 2
	})

	v1, _ := s.Get(Entity(1))
	v2, _ := s.Get(Entity(2))
	assert.Equal(t, 20, *v1)
	assert.Equal(t, 40, *v2)
}

func TestQuery2EmptyStores(t *testing.T) {
	w := NewWorld()
	count := 0
	Query2[Position, Velocity](w, func(_ Entity, _ *Position, _ *Velocity) {
		count++
	})
	assert.Equal(t, 0, count)
}

func TestQuery3EmptyStores(t *testing.T) {
	w := NewWorld()
	count := 0
	Query3[Position, Velocity, Health](w, func(_ Entity, _ *Position, _ *Velocity, _ *Health) {
		count++
	})
	assert.Equal(t, 0, count)
}

func TestWorldEntityManager(t *testing.T) {
	w := NewWorld()
	assert.NotNil(t, w.EntityManager())
}

func TestWorldStores(t *testing.T) {
	w := NewWorld()
	assert.NotNil(t, w.Stores())
	GetStore[Position](w)
	assert.Equal(t, 1, len(w.Stores()))
}
