package mcmath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlane_Normalize(t *testing.T) {
	p := Plane{Normal: Vec3{0, 0, 2}, D: 4}
	n := p.normalize()
	assert.InDelta(t, float32(1), n.Normal.Length(), 1e-5)
	assert.InDelta(t, float32(2), n.D, 1e-5)
}

func TestPlane_Normalize_Zero(t *testing.T) {
	p := Plane{Normal: Vec3{0, 0, 0}, D: 4}
	n := p.normalize()
	assert.Equal(t, p, n)
}

func TestPlane_DistanceToPoint(t *testing.T) {
	// Plane: y = 5 (normal=(0,1,0), D=-5)
	p := Plane{Normal: Vec3{0, 1, 0}, D: -5}
	assert.InDelta(t, float32(5), p.distanceToPoint(Vec3{0, 10, 0}), 1e-5)
	assert.InDelta(t, float32(-5), p.distanceToPoint(Vec3{0, 0, 0}), 1e-5)
	assert.InDelta(t, float32(0), p.distanceToPoint(Vec3{0, 5, 0}), 1e-5)
}

func TestExtractFromMatrix_Identity(t *testing.T) {
	m := identityMatrix()
	f := ExtractFromMatrix(m)

	// With identity matrix, frustum should contain the origin
	assert.True(t, f.ContainsPoint(Vec3{0, 0, 0}))
}

func TestFrustum_ContainsPoint_Orthographic(t *testing.T) {
	// Create an orthographic projection: a box from -10 to +10 in all axes
	m := orthographicMatrix(-10, 10, -10, 10, -10, 10)
	f := ExtractFromMatrix(m)

	assert.True(t, f.ContainsPoint(Vec3{0, 0, 0}))
	assert.True(t, f.ContainsPoint(Vec3{5, 5, 5}))
	assert.True(t, f.ContainsPoint(Vec3{-5, -5, -5}))

	// Points outside the box
	assert.False(t, f.ContainsPoint(Vec3{15, 0, 0}))
	assert.False(t, f.ContainsPoint(Vec3{0, 15, 0}))
	assert.False(t, f.ContainsPoint(Vec3{0, 0, 15}))
	assert.False(t, f.ContainsPoint(Vec3{-15, 0, 0}))
}

func TestFrustum_IntersectsAABB_Orthographic(t *testing.T) {
	m := orthographicMatrix(-10, 10, -10, 10, -10, 10)
	f := ExtractFromMatrix(m)

	// AABB fully inside
	inside := AABB{Vec3{-1, -1, -1}, Vec3{1, 1, 1}}
	assert.True(t, f.IntersectsAABB(inside))

	// AABB partially inside (straddling a boundary)
	partial := AABB{Vec3{8, 0, 0}, Vec3{12, 2, 2}}
	assert.True(t, f.IntersectsAABB(partial))

	// AABB fully outside
	outside := AABB{Vec3{20, 20, 20}, Vec3{30, 30, 30}}
	assert.False(t, f.IntersectsAABB(outside))
}

func TestFrustum_ContainsPoint_Perspective(t *testing.T) {
	fov := float32(math.Pi / 2) // 90 degrees
	m := perspectiveMatrix(fov, 1.0, 0.1, 100.0)
	f := ExtractFromMatrix(m)

	// Point in front of the camera near the center (in clip space, z is negative for OpenGL)
	assert.True(t, f.ContainsPoint(Vec3{0, 0, -1}))
	assert.True(t, f.ContainsPoint(Vec3{0, 0, -50}))

	// Point behind camera (positive z in OpenGL)
	assert.False(t, f.ContainsPoint(Vec3{0, 0, 1}))
}

func TestFrustum_IntersectsAABB_Perspective(t *testing.T) {
	fov := float32(math.Pi / 2)
	m := perspectiveMatrix(fov, 1.0, 0.1, 100.0)
	f := ExtractFromMatrix(m)

	// AABB right in front of the camera
	front := AABB{Vec3{-1, -1, -5}, Vec3{1, 1, -3}}
	assert.True(t, f.IntersectsAABB(front))

	// AABB behind the camera
	behind := AABB{Vec3{-1, -1, 10}, Vec3{1, 1, 20}}
	assert.False(t, f.IntersectsAABB(behind))
}

func TestFrustum_IntersectsAABB_AllSides(t *testing.T) {
	m := orthographicMatrix(-10, 10, -10, 10, -10, 10)
	f := ExtractFromMatrix(m)

	// Each side outside
	tests := []struct {
		name string
		box  AABB
		want bool
	}{
		{"left outside", AABB{Vec3{-20, 0, 0}, Vec3{-15, 1, 1}}, false},
		{"right outside", AABB{Vec3{15, 0, 0}, Vec3{20, 1, 1}}, false},
		{"bottom outside", AABB{Vec3{0, -20, 0}, Vec3{1, -15, 1}}, false},
		{"top outside", AABB{Vec3{0, 15, 0}, Vec3{1, 20, 1}}, false},
		{"near outside", AABB{Vec3{0, 0, -20}, Vec3{1, 1, -15}}, false},
		{"far outside", AABB{Vec3{0, 0, 15}, Vec3{1, 1, 20}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, f.IntersectsAABB(tc.box))
		})
	}
}

func TestIdentityMatrix(t *testing.T) {
	m := identityMatrix()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			idx := j*4 + i // column-major
			if i == j {
				assert.Equal(t, float32(1), m[idx])
			} else {
				assert.Equal(t, float32(0), m[idx])
			}
		}
	}
}
