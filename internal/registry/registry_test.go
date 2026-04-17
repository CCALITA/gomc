package registry_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/fanxiyao/gomc/internal/registry"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newPopulated(t *testing.T) *registry.Registry[string] {
	t.Helper()
	r := registry.New[string]()
	id0, err := r.Register("air", "block_air")
	assert.NoError(t, err)
	assert.Equal(t, uint16(0), id0)

	id1, err := r.Register("stone", "block_stone")
	assert.NoError(t, err)
	assert.Equal(t, uint16(1), id1)

	id2, err := r.Register("dirt", "block_dirt")
	assert.NoError(t, err)
	assert.Equal(t, uint16(2), id2)

	return r
}

// ---------------------------------------------------------------------------
// Register / Get round-trip
// ---------------------------------------------------------------------------

func TestRegisterAndGet(t *testing.T) {
	r := newPopulated(t)

	item, ok := r.Get(0)
	assert.True(t, ok)
	assert.Equal(t, "block_air", item)

	item, ok = r.Get(1)
	assert.True(t, ok)
	assert.Equal(t, "block_stone", item)

	item, ok = r.Get(2)
	assert.True(t, ok)
	assert.Equal(t, "block_dirt", item)

	// out of range
	_, ok = r.Get(999)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Sequential IDs
// ---------------------------------------------------------------------------

func TestIDsAreSequential(t *testing.T) {
	r := registry.New[int]()
	for i := 0; i < 10; i++ {
		id, err := r.Register(string(rune('a'+i)), i*10)
		assert.NoError(t, err)
		assert.Equal(t, uint16(i), id)
	}
	assert.Equal(t, 10, r.Len())
}

// ---------------------------------------------------------------------------
// Freeze blocks registration
// ---------------------------------------------------------------------------

func TestFreezeBlocksRegistration(t *testing.T) {
	r := newPopulated(t)
	r.Freeze()

	_, err := r.Register("grass", "block_grass")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, registry.ErrFrozen))

	// existing items still accessible after freeze
	item, ok := r.Get(0)
	assert.True(t, ok)
	assert.Equal(t, "block_air", item)
}

// ---------------------------------------------------------------------------
// Duplicate name
// ---------------------------------------------------------------------------

func TestDuplicateNameError(t *testing.T) {
	r := registry.New[string]()
	_, err := r.Register("stone", "v1")
	assert.NoError(t, err)

	_, err = r.Register("stone", "v2")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, registry.ErrDuplicateName))
}

// ---------------------------------------------------------------------------
// GetByName
// ---------------------------------------------------------------------------

func TestGetByName(t *testing.T) {
	r := newPopulated(t)

	item, id, ok := r.GetByName("stone")
	assert.True(t, ok)
	assert.Equal(t, uint16(1), id)
	assert.Equal(t, "block_stone", item)

	// missing name
	_, _, ok = r.GetByName("missing")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// ID and Name
// ---------------------------------------------------------------------------

func TestIDAndName(t *testing.T) {
	r := newPopulated(t)

	id, ok := r.ID("dirt")
	assert.True(t, ok)
	assert.Equal(t, uint16(2), id)

	name, ok := r.Name(2)
	assert.True(t, ok)
	assert.Equal(t, "dirt", name)

	// missing
	_, ok = r.ID("nope")
	assert.False(t, ok)

	_, ok = r.Name(999)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Range iteration
// ---------------------------------------------------------------------------

func TestRangeFull(t *testing.T) {
	r := newPopulated(t)

	var ids []uint16
	var names []string
	var items []string

	r.Range(func(id uint16, name string, item string) bool {
		ids = append(ids, id)
		names = append(names, name)
		items = append(items, item)
		return true
	})

	assert.Equal(t, []uint16{0, 1, 2}, ids)
	assert.Equal(t, []string{"air", "stone", "dirt"}, names)
	assert.Equal(t, []string{"block_air", "block_stone", "block_dirt"}, items)
}

func TestRangeEarlyStop(t *testing.T) {
	r := newPopulated(t)

	count := 0
	r.Range(func(id uint16, name string, item string) bool {
		count++
		return id < 1 // stop after id 1
	})
	assert.Equal(t, 2, count)
}

func TestRangeEmpty(t *testing.T) {
	r := registry.New[int]()
	called := false
	r.Range(func(id uint16, name string, item int) bool {
		called = true
		return true
	})
	assert.False(t, called)
}

// ---------------------------------------------------------------------------
// Concurrent register safety
// ---------------------------------------------------------------------------

func TestConcurrentRegisterSafety(t *testing.T) {
	r := registry.New[int]()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errs := make([]error, goroutines)
	ids := make([]uint16, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := string(rune('A'+idx/26)) + string(rune('a'+idx%26))
			ids[idx], errs[idx] = r.Register(name, idx)
		}(i)
	}
	wg.Wait()

	// All registrations should succeed (names are unique).
	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d failed", i)
	}
	assert.Equal(t, goroutines, r.Len())

	// IDs should cover [0, goroutines) with no duplicates.
	seen := make(map[uint16]bool)
	for _, id := range ids {
		assert.False(t, seen[id], "duplicate id %d", id)
		seen[id] = true
	}
}

// ---------------------------------------------------------------------------
// Concurrent reads after freeze
// ---------------------------------------------------------------------------

func TestConcurrentReadsAfterFreeze(t *testing.T) {
	r := newPopulated(t)
	r.Freeze()

	var wg sync.WaitGroup
	const readers = 50
	wg.Add(readers)

	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			_, _ = r.Get(0)
			_, _, _ = r.GetByName("stone")
			_, _ = r.ID("dirt")
			_, _ = r.Name(1)
			_ = r.Len()
			r.Range(func(id uint16, name string, item string) bool { return true })
		}()
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Len
// ---------------------------------------------------------------------------

func TestLenEmpty(t *testing.T) {
	r := registry.New[float64]()
	assert.Equal(t, 0, r.Len())
}

func TestLenAfterRegistrations(t *testing.T) {
	r := newPopulated(t)
	assert.Equal(t, 3, r.Len())
}

// ---------------------------------------------------------------------------
// Error message wrapping
// ---------------------------------------------------------------------------

func TestErrorMessagesContainName(t *testing.T) {
	r := registry.New[string]()
	_, _ = r.Register("dup", "first")

	_, err := r.Register("dup", "second")
	assert.Contains(t, err.Error(), "dup")

	r.Freeze()
	_, err = r.Register("new", "val")
	assert.Contains(t, err.Error(), "new")
}
