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

// ---------------------------------------------------------------------------
// StairAABBs tests
// ---------------------------------------------------------------------------

func TestStairAABBs_ReturnsTwo(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	for orient := 0; orient <= 3; orient++ {
		aabbs := StairAABBs(pos, orient)
		assert.Len(t, aabbs, 2, "orientation %d should return 2 AABBs", orient)
	}
}

func TestStairAABBs_BottomHalf(t *testing.T) {
	pos := BlockPos{5, 10, 3}
	// Bottom half is the same regardless of orientation.
	for orient := 0; orient <= 3; orient++ {
		aabbs := StairAABBs(pos, orient)
		bottom := aabbs[0]
		assert.Equal(t, Vec3{5, 10, 3}, bottom.Min, "orient %d bottom min", orient)
		assert.Equal(t, Vec3{6, 10.5, 4}, bottom.Max, "orient %d bottom max", orient)
	}
}

func TestStairAABBs_NorthStep(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	aabbs := StairAABBs(pos, 0) // North: step on +Z side
	top := aabbs[1]
	assert.Equal(t, Vec3{0, 0.5, 0.5}, top.Min)
	assert.Equal(t, Vec3{1, 1, 1}, top.Max)
}

func TestStairAABBs_SouthStep(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	aabbs := StairAABBs(pos, 1) // South: step on -Z side
	top := aabbs[1]
	assert.Equal(t, Vec3{0, 0.5, 0}, top.Min)
	assert.Equal(t, Vec3{1, 1, 0.5}, top.Max)
}

func TestStairAABBs_EastStep(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	aabbs := StairAABBs(pos, 2) // East: step on -X side
	top := aabbs[1]
	assert.Equal(t, Vec3{0, 0.5, 0}, top.Min)
	assert.Equal(t, Vec3{0.5, 1, 1}, top.Max)
}

func TestStairAABBs_WestStep(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	aabbs := StairAABBs(pos, 3) // West: step on +X side
	top := aabbs[1]
	assert.Equal(t, Vec3{0.5, 0.5, 0}, top.Min)
	assert.Equal(t, Vec3{1, 1, 1}, top.Max)
}

func TestStairAABBs_DefaultOrientation(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	// Unknown orientation values should default to North.
	aabbs := StairAABBs(pos, 99)
	top := aabbs[1]
	assert.Equal(t, Vec3{0, 0.5, 0.5}, top.Min)
	assert.Equal(t, Vec3{1, 1, 1}, top.Max)
}

func TestStairAABBs_NegativePosition(t *testing.T) {
	pos := BlockPos{-1, 0, -1}
	aabbs := StairAABBs(pos, 0)
	assert.Equal(t, Vec3{-1, 0, -1}, aabbs[0].Min)
	assert.Equal(t, Vec3{0, 0.5, 0}, aabbs[0].Max)
	assert.Equal(t, Vec3{-1, 0.5, -0.5}, aabbs[1].Min)
	assert.Equal(t, Vec3{0, 1, 0}, aabbs[1].Max)
}

// ---------------------------------------------------------------------------
// SlabAABB tests
// ---------------------------------------------------------------------------

func TestSlabAABB_Bottom(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	slab := SlabAABB(pos, false)
	assert.Equal(t, Vec3{0, 0, 0}, slab.Min)
	assert.Equal(t, Vec3{1, 0.5, 1}, slab.Max)
}

func TestSlabAABB_Top(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	slab := SlabAABB(pos, true)
	assert.Equal(t, Vec3{0, 0.5, 0}, slab.Min)
	assert.Equal(t, Vec3{1, 1, 1}, slab.Max)
}

func TestSlabAABB_WithOffset(t *testing.T) {
	pos := BlockPos{3, 10, 7}
	slab := SlabAABB(pos, false)
	assert.Equal(t, Vec3{3, 10, 7}, slab.Min)
	assert.Equal(t, Vec3{4, 10.5, 8}, slab.Max)
}

func TestSlabAABB_NegativePosition(t *testing.T) {
	pos := BlockPos{-2, 0, -3}
	slab := SlabAABB(pos, true)
	assert.Equal(t, Vec3{-2, 0.5, -3}, slab.Min)
	assert.Equal(t, Vec3{-1, 1, -2}, slab.Max)
}

func TestSlabAABB_HalfHeight(t *testing.T) {
	pos := BlockPos{0, 0, 0}
	bottom := SlabAABB(pos, false)
	top := SlabAABB(pos, true)

	// Both should have exactly 0.5 height.
	assert.InDelta(t, float32(0.5), bottom.Size().Y, 1e-5)
	assert.InDelta(t, float32(0.5), top.Size().Y, 1e-5)

	// Full width and depth.
	assert.InDelta(t, float32(1), bottom.Size().X, 1e-5)
	assert.InDelta(t, float32(1), bottom.Size().Z, 1e-5)
}
