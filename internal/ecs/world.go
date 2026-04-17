package ecs

import "reflect"

// World is the top-level container for entities and their component stores.
type World struct {
	em     *EntityManager
	stores map[reflect.Type]any
}

// NewWorld creates a World with an initialised EntityManager.
func NewWorld() *World {
	return &World{
		em:     NewEntityManager(),
		stores: make(map[reflect.Type]any),
	}
}

// NewEntity allocates a new entity through the underlying EntityManager.
func (w *World) NewEntity() Entity {
	return w.em.Create()
}

// DestroyEntity marks the entity as dead.
func (w *World) DestroyEntity(e Entity) {
	w.em.Destroy(e)
}

// Alive reports whether the entity is still alive.
func (w *World) Alive(e Entity) bool {
	return w.em.Alive(e)
}

// EntityManager exposes the underlying manager for advanced use.
func (w *World) EntityManager() *EntityManager {
	return w.em
}

// Stores returns the raw store map (used by generic accessor / queries).
func (w *World) Stores() map[reflect.Type]any {
	return w.stores
}

// GetStore returns the ComponentStore for type T, creating it on first access.
func GetStore[T any](w *World) *ComponentStore[T] {
	t := reflect.TypeFor[T]()
	if s, ok := w.stores[t]; ok {
		return s.(*ComponentStore[T])
	}
	s := NewComponentStore[T]()
	w.stores[t] = s
	return s
}
