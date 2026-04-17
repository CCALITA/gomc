// Package registry provides a generic, thread-safe registry that maps
// string names to sequential uint16 IDs and their associated items.
// Once frozen, no further registrations are accepted.
package registry

import (
	"errors"
	"fmt"
	"sync"
)

// Sentinel errors returned by Register.
var (
	ErrFrozen        = errors.New("registry is frozen")
	ErrDuplicateName = errors.New("duplicate name")
)

// entry holds a single registered item together with its name.
type entry[T any] struct {
	name string
	item T
}

// Registry is a generic, thread-safe collection that assigns sequential
// uint16 IDs (starting from 0) to named items.  After Freeze is called
// the registry becomes read-only.
type Registry[T any] struct {
	mu      sync.RWMutex
	entries []entry[T]       // indexed by id
	byName  map[string]uint16 // name → id
	frozen  bool
}

// New creates an empty Registry.
func New[T any]() *Registry[T] {
	return &Registry[T]{
		byName: make(map[string]uint16),
	}
}

// Register adds an item under the given name and returns its newly
// assigned sequential ID.  It returns ErrFrozen if the registry has been
// frozen and ErrDuplicateName if name is already registered.
func (r *Registry[T]) Register(name string, item T) (uint16, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.frozen {
		return 0, fmt.Errorf("register %q: %w", name, ErrFrozen)
	}
	if _, exists := r.byName[name]; exists {
		return 0, fmt.Errorf("register %q: %w", name, ErrDuplicateName)
	}

	id := uint16(len(r.entries))
	r.entries = append(r.entries, entry[T]{name: name, item: item})
	r.byName[name] = id
	return id, nil
}

// Freeze locks the registry so that future Register calls return ErrFrozen.
func (r *Registry[T]) Freeze() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
}

// Get returns the item for the given ID.
func (r *Registry[T]) Get(id uint16) (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if int(id) >= len(r.entries) {
		var zero T
		return zero, false
	}
	return r.entries[id].item, true
}

// GetByName returns the item and its ID for the given name.
func (r *Registry[T]) GetByName(name string) (T, uint16, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byName[name]
	if !ok {
		var zero T
		return zero, 0, false
	}
	return r.entries[id].item, id, true
}

// ID returns the numeric ID for the given name.
func (r *Registry[T]) ID(name string) (uint16, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byName[name]
	return id, ok
}

// Name returns the string name for the given ID.
func (r *Registry[T]) Name(id uint16) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if int(id) >= len(r.entries) {
		return "", false
	}
	return r.entries[id].name, true
}

// Len returns the number of registered items.
func (r *Registry[T]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// Range iterates over every entry in ID order.  The callback receives
// the ID, name, and item.  Return false from fn to stop iteration early.
func (r *Registry[T]) Range(fn func(id uint16, name string, item T) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i, e := range r.entries {
		if !fn(uint16(i), e.name, e.item) {
			return
		}
	}
}
