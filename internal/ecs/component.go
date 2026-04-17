package ecs

// ComponentStore is a sparse-set backed store for components of type T.
// Dense and data slices are kept in parallel so iteration is cache-friendly.
// Removals use swap-remove to maintain density.
type ComponentStore[T any] struct {
	dense  []Entity
	data   []T
	sparse map[Entity]int // entity -> index in dense/data
}

// NewComponentStore returns an empty ComponentStore[T].
func NewComponentStore[T any]() *ComponentStore[T] {
	return &ComponentStore[T]{
		sparse: make(map[Entity]int),
	}
}

// Set associates component value with entity, overwriting any previous value.
func (s *ComponentStore[T]) Set(e Entity, v T) {
	if idx, ok := s.sparse[e]; ok {
		s.data[idx] = v
		return
	}
	s.sparse[e] = len(s.dense)
	s.dense = append(s.dense, e)
	s.data = append(s.data, v)
}

// Get returns a pointer to the component for entity and true, or nil and
// false when the entity has no component in this store.
func (s *ComponentStore[T]) Get(e Entity) (*T, bool) {
	idx, ok := s.sparse[e]
	if !ok {
		return nil, false
	}
	return &s.data[idx], true
}

// Remove deletes the component for entity using swap-remove and returns
// true, or returns false if the entity was not present.
func (s *ComponentStore[T]) Remove(e Entity) bool {
	idx, ok := s.sparse[e]
	if !ok {
		return false
	}

	last := len(s.dense) - 1
	if idx != last {
		// Move the last element into the vacated slot.
		movedEntity := s.dense[last]
		s.dense[idx] = movedEntity
		s.data[idx] = s.data[last]
		s.sparse[movedEntity] = idx
	}

	// Shrink slices and remove sparse entry.
	s.dense = s.dense[:last]
	s.data = s.data[:last]
	delete(s.sparse, e)
	return true
}

// Has reports whether entity has a component in this store.
func (s *ComponentStore[T]) Has(e Entity) bool {
	_, ok := s.sparse[e]
	return ok
}

// Len returns the number of stored components.
func (s *ComponentStore[T]) Len() int {
	return len(s.dense)
}

// Each calls fn for every (entity, component) pair in dense order.
func (s *ComponentStore[T]) Each(fn func(Entity, *T)) {
	for i := range s.dense {
		fn(s.dense[i], &s.data[i])
	}
}
