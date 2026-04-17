package mcmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAABB_Intersects(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{2, 2, 2}}
	b := AABB{Vec3{1, 1, 1}, Vec3{3, 3, 3}}
	c := AABB{Vec3{5, 5, 5}, Vec3{6, 6, 6}}

	assert.True(t, a.Intersects(b))
	assert.True(t, b.Intersects(a))
	assert.False(t, a.Intersects(c))
	assert.False(t, c.Intersects(a))
}

func TestAABB_Intersects_Touching(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{1, 1, 1}}
	b := AABB{Vec3{1, 0, 0}, Vec3{2, 1, 1}}
	// Touching faces should not count as intersection (strict inequality)
	assert.False(t, a.Intersects(b))
}

func TestAABB_Contains(t *testing.T) {
	outer := AABB{Vec3{0, 0, 0}, Vec3{10, 10, 10}}
	inner := AABB{Vec3{1, 1, 1}, Vec3{5, 5, 5}}
	outside := AABB{Vec3{-1, 0, 0}, Vec3{5, 5, 5}}

	assert.True(t, outer.Contains(inner))
	assert.False(t, inner.Contains(outer))
	assert.False(t, outer.Contains(outside))
}

func TestAABB_Contains_Self(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{5, 5, 5}}
	assert.True(t, a.Contains(a))
}

func TestAABB_ContainsPoint(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{5, 5, 5}}
	assert.True(t, a.ContainsPoint(Vec3{2.5, 2.5, 2.5}))
	assert.True(t, a.ContainsPoint(Vec3{0, 0, 0})) // boundary
	assert.True(t, a.ContainsPoint(Vec3{5, 5, 5})) // boundary
	assert.False(t, a.ContainsPoint(Vec3{-1, 0, 0}))
	assert.False(t, a.ContainsPoint(Vec3{0, 0, 6}))
}

func TestAABB_Expand(t *testing.T) {
	a := AABB{Vec3{1, 1, 1}, Vec3{3, 3, 3}}
	expanded := a.Expand(Vec3{0, 0, 0})
	assert.Equal(t, Vec3{0, 0, 0}, expanded.Min)
	assert.Equal(t, Vec3{3, 3, 3}, expanded.Max)

	expanded2 := a.Expand(Vec3{5, 5, 5})
	assert.Equal(t, Vec3{1, 1, 1}, expanded2.Min)
	assert.Equal(t, Vec3{5, 5, 5}, expanded2.Max)
}

func TestAABB_Offset(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{1, 1, 1}}
	moved := a.Offset(Vec3{5, 5, 5})
	assert.Equal(t, Vec3{5, 5, 5}, moved.Min)
	assert.Equal(t, Vec3{6, 6, 6}, moved.Max)
}

func TestAABB_Size(t *testing.T) {
	a := AABB{Vec3{1, 2, 3}, Vec3{4, 6, 8}}
	assert.Equal(t, Vec3{3, 4, 5}, a.Size())
}

func TestAABB_Center(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{4, 4, 4}}
	c := a.Center()
	assert.InDelta(t, float32(2), c.X, 1e-5)
	assert.InDelta(t, float32(2), c.Y, 1e-5)
	assert.InDelta(t, float32(2), c.Z, 1e-5)
}

func TestAABB_Grow(t *testing.T) {
	a := AABB{Vec3{1, 1, 1}, Vec3{3, 3, 3}}
	grown := a.Grow(0.5)
	assert.InDelta(t, float32(0.5), grown.Min.X, 1e-5)
	assert.InDelta(t, float32(0.5), grown.Min.Y, 1e-5)
	assert.InDelta(t, float32(0.5), grown.Min.Z, 1e-5)
	assert.InDelta(t, float32(3.5), grown.Max.X, 1e-5)
	assert.InDelta(t, float32(3.5), grown.Max.Y, 1e-5)
	assert.InDelta(t, float32(3.5), grown.Max.Z, 1e-5)
}

func TestAABB_String(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{1, 1, 1}}
	s := a.String()
	assert.Contains(t, s, "AABB")
}

func TestBlockAABB(t *testing.T) {
	pos := BlockPos{5, 10, 3}
	box := BlockAABB(pos)
	assert.Equal(t, Vec3{5, 10, 3}, box.Min)
	assert.Equal(t, Vec3{6, 11, 4}, box.Max)
}

func TestBlockAABB_Negative(t *testing.T) {
	pos := BlockPos{-1, 0, -1}
	box := BlockAABB(pos)
	assert.Equal(t, Vec3{-1, 0, -1}, box.Min)
	assert.Equal(t, Vec3{0, 1, 0}, box.Max)
}

func TestAABB_Immutable(t *testing.T) {
	a := AABB{Vec3{0, 0, 0}, Vec3{1, 1, 1}}
	original := a
	a.Offset(Vec3{5, 5, 5})
	assert.Equal(t, original, a, "Offset should not mutate the original")
	a.Grow(1)
	assert.Equal(t, original, a, "Grow should not mutate the original")
}
