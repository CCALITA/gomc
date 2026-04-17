package mcmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRay_At(t *testing.T) {
	r := Ray{Origin: Vec3{0, 0, 0}, Direction: Vec3{1, 0, 0}}
	p := r.At(5)
	assert.Equal(t, Vec3{5, 0, 0}, p)
}

func TestRay_At_Diagonal(t *testing.T) {
	r := Ray{Origin: Vec3{1, 2, 3}, Direction: Vec3{1, 1, 1}}
	p := r.At(2)
	assert.InDelta(t, float32(3), p.X, 1e-5)
	assert.InDelta(t, float32(4), p.Y, 1e-5)
	assert.InDelta(t, float32(5), p.Z, 1e-5)
}

func TestRay_CastBlocks_HitFirstBlock(t *testing.T) {
	// Origin is inside a solid block at (0,0,0)
	r := Ray{Origin: Vec3{0.5, 0.5, 0.5}, Direction: Vec3{1, 0, 0}}
	hit, pos, _, tVal := r.CastBlocks(10, func(bp BlockPos) bool {
		return bp.X == 0 && bp.Y == 0 && bp.Z == 0
	})
	assert.True(t, hit)
	assert.Equal(t, BlockPos{0, 0, 0}, pos)
	assert.InDelta(t, float32(0), tVal, 1e-5)
}

func TestRay_CastBlocks_HitBlock(t *testing.T) {
	r := Ray{Origin: Vec3{0.5, 0.5, 0.5}, Direction: Vec3{1, 0, 0}}
	targetX := int32(5)
	hit, pos, face, tVal := r.CastBlocks(10, func(bp BlockPos) bool {
		return bp.X == targetX && bp.Y == 0 && bp.Z == 0
	})
	assert.True(t, hit)
	assert.Equal(t, targetX, pos.X)
	assert.Equal(t, int32(0), pos.Y)
	assert.Equal(t, int32(0), pos.Z)
	assert.Equal(t, West, face) // Ray enters from -X side
	assert.True(t, tVal > 0)
}

func TestRay_CastBlocks_Miss(t *testing.T) {
	r := Ray{Origin: Vec3{0.5, 0.5, 0.5}, Direction: Vec3{1, 0, 0}}
	hit, _, _, _ := r.CastBlocks(10, func(bp BlockPos) bool {
		return false // no solid blocks
	})
	assert.False(t, hit)
}

func TestRay_CastBlocks_NegativeDirection(t *testing.T) {
	r := Ray{Origin: Vec3{5.5, 0.5, 0.5}, Direction: Vec3{-1, 0, 0}}
	hit, pos, face, _ := r.CastBlocks(10, func(bp BlockPos) bool {
		return bp.X == 0 && bp.Y == 0 && bp.Z == 0
	})
	assert.True(t, hit)
	assert.Equal(t, BlockPos{0, 0, 0}, pos)
	assert.Equal(t, East, face) // Ray enters from +X side
}

func TestRay_CastBlocks_VerticalUp(t *testing.T) {
	r := Ray{Origin: Vec3{0.5, 0.5, 0.5}, Direction: Vec3{0, 1, 0}}
	hit, pos, face, _ := r.CastBlocks(20, func(bp BlockPos) bool {
		return bp.Y == 10
	})
	assert.True(t, hit)
	assert.Equal(t, int32(10), pos.Y)
	assert.Equal(t, Down, face) // Ray enters from -Y side
}

func TestRay_CastBlocks_VerticalDown(t *testing.T) {
	r := Ray{Origin: Vec3{0.5, 10.5, 0.5}, Direction: Vec3{0, -1, 0}}
	hit, pos, face, _ := r.CastBlocks(20, func(bp BlockPos) bool {
		return bp.Y == 0
	})
	assert.True(t, hit)
	assert.Equal(t, int32(0), pos.Y)
	assert.Equal(t, Up, face) // Ray enters from +Y side
}

func TestRay_CastBlocks_DiagonalZ(t *testing.T) {
	r := Ray{Origin: Vec3{0.5, 0.5, 0.5}, Direction: Vec3{0, 0, 1}}
	hit, pos, face, _ := r.CastBlocks(10, func(bp BlockPos) bool {
		return bp.Z == 5 && bp.X == 0 && bp.Y == 0
	})
	assert.True(t, hit)
	assert.Equal(t, int32(5), pos.Z)
	assert.Equal(t, North, face) // Ray enters from -Z side
}

func TestRay_CastBlocks_NegativeZ(t *testing.T) {
	r := Ray{Origin: Vec3{0.5, 0.5, 5.5}, Direction: Vec3{0, 0, -1}}
	hit, pos, face, _ := r.CastBlocks(10, func(bp BlockPos) bool {
		return bp.Z == 0 && bp.X == 0 && bp.Y == 0
	})
	assert.True(t, hit)
	assert.Equal(t, int32(0), pos.Z)
	assert.Equal(t, South, face) // Ray enters from +Z side
}

func TestRay_CastBlocks_MaxDistance(t *testing.T) {
	r := Ray{Origin: Vec3{0.5, 0.5, 0.5}, Direction: Vec3{1, 0, 0}}
	hit, _, _, _ := r.CastBlocks(3, func(bp BlockPos) bool {
		return bp.X == 10 // too far
	})
	assert.False(t, hit)
}
