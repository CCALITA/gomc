//go:build !ci

package render

import (
	"math"
	"testing"
	"unsafe"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// WireframeRenderer set/clear target tests
// ---------------------------------------------------------------------------

func TestWireframeRenderer_SetTarget(t *testing.T) {
	wr := &WireframeRenderer{}

	if wr.IsActive() {
		t.Error("new WireframeRenderer should not be active")
	}

	pos := mcmath.BlockPos{X: 10, Y: 64, Z: -5}
	wr.SetTarget(pos)

	if !wr.IsActive() {
		t.Error("expected active after SetTarget")
	}

	got, active := wr.Target()
	if !active {
		t.Error("expected Target() to report active")
	}
	if got != pos {
		t.Errorf("expected target %v, got %v", pos, got)
	}
}

func TestWireframeRenderer_ClearTarget(t *testing.T) {
	wr := &WireframeRenderer{}

	wr.SetTarget(mcmath.BlockPos{X: 1, Y: 2, Z: 3})
	wr.ClearTarget()

	if wr.IsActive() {
		t.Error("expected inactive after ClearTarget")
	}

	_, active := wr.Target()
	if active {
		t.Error("expected Target() to report inactive after ClearTarget")
	}
}

func TestWireframeRenderer_SetTarget_Overwrite(t *testing.T) {
	wr := &WireframeRenderer{}

	wr.SetTarget(mcmath.BlockPos{X: 1, Y: 2, Z: 3})
	wr.SetTarget(mcmath.BlockPos{X: 10, Y: 20, Z: 30})

	got, active := wr.Target()
	if !active {
		t.Error("expected active after second SetTarget")
	}
	expected := mcmath.BlockPos{X: 10, Y: 20, Z: 30}
	if got != expected {
		t.Errorf("expected target %v, got %v", expected, got)
	}
}

// ---------------------------------------------------------------------------
// Wireframe cube vertex generation tests
// ---------------------------------------------------------------------------

func TestWireframeCubeVertices_Count(t *testing.T) {
	verts := WireframeCubeVertices()

	// 12 lines * 2 endpoints * 3 floats = 72 floats
	expectedFloats := 24 * 3
	if len(verts) != expectedFloats {
		t.Errorf("expected %d floats, got %d", expectedFloats, len(verts))
	}

	// 24 vertices (for 12 lines)
	expectedVertices := 24
	if len(verts)/3 != expectedVertices {
		t.Errorf("expected %d vertices, got %d", expectedVertices, len(verts)/3)
	}
}

func TestWireframeCubeVertices_InUnitCube(t *testing.T) {
	verts := WireframeCubeVertices()

	for i := 0; i < len(verts); i += 3 {
		x, y, z := verts[i], verts[i+1], verts[i+2]
		if x < 0 || x > 1 || y < 0 || y > 1 || z < 0 || z > 1 {
			t.Errorf("vertex %d (%f, %f, %f) outside unit cube [0,1]", i/3, x, y, z)
		}
	}
}

func TestWireframeCubeVertices_12UniqueEdges(t *testing.T) {
	verts := WireframeCubeVertices()

	type edge struct {
		ax, ay, az float32
		bx, by, bz float32
	}

	edges := make(map[edge]bool)
	for i := 0; i < len(verts); i += 6 {
		e := edge{verts[i], verts[i+1], verts[i+2], verts[i+3], verts[i+4], verts[i+5]}
		// Normalise edge direction to avoid counting (a,b) and (b,a) as different.
		reversed := edge{e.bx, e.by, e.bz, e.ax, e.ay, e.az}
		if _, ok := edges[reversed]; ok {
			t.Errorf("duplicate edge found at line %d", i/6)
		}
		edges[e] = true
	}

	if len(edges) != 12 {
		t.Errorf("expected 12 unique edges, got %d", len(edges))
	}
}

// ---------------------------------------------------------------------------
// Wireframe AABB tests
// ---------------------------------------------------------------------------

func TestWireframeBlockAABB_SlightlyLarger(t *testing.T) {
	pos := mcmath.BlockPos{X: 5, Y: 10, Z: 15}
	aabb := WireframeBlockAABB(pos)

	size := aabb.Size()

	if math.Abs(float64(size.X-wireframeScale)) > 1e-5 {
		t.Errorf("expected X size %f, got %f", wireframeScale, size.X)
	}
	if math.Abs(float64(size.Y-wireframeScale)) > 1e-5 {
		t.Errorf("expected Y size %f, got %f", wireframeScale, size.Y)
	}
	if math.Abs(float64(size.Z-wireframeScale)) > 1e-5 {
		t.Errorf("expected Z size %f, got %f", wireframeScale, size.Z)
	}

	// The AABB should contain the block's unit cube.
	blockAABB := mcmath.BlockAABB(pos)
	if !aabb.ContainsPoint(blockAABB.Min) {
		t.Error("wireframe AABB should contain the block's min corner")
	}
	if !aabb.ContainsPoint(blockAABB.Max) {
		t.Error("wireframe AABB should contain the block's max corner")
	}
}

func TestWireframeBlockAABB_CentredOnBlock(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	aabb := WireframeBlockAABB(pos)
	center := aabb.Center()

	// Centre of a block at (0,0,0) is (0.5, 0.5, 0.5).
	if math.Abs(float64(center.X-0.5)) > 1e-5 {
		t.Errorf("expected center X 0.5, got %f", center.X)
	}
	if math.Abs(float64(center.Y-0.5)) > 1e-5 {
		t.Errorf("expected center Y 0.5, got %f", center.Y)
	}
	if math.Abs(float64(center.Z-0.5)) > 1e-5 {
		t.Errorf("expected center Z 0.5, got %f", center.Z)
	}
}

// ---------------------------------------------------------------------------
// WireframePushConstants size test
// ---------------------------------------------------------------------------

func TestWireframePushConstants_Size(t *testing.T) {
	var pc WireframePushConstants
	size := int(unsafe.Sizeof(pc))
	if size != 64 {
		t.Errorf("WireframePushConstants size should be 64 bytes, got %d", size)
	}
}

// ---------------------------------------------------------------------------
// Chunk AABB from key test (frustum culling helper)
// ---------------------------------------------------------------------------

func TestChunkAABBFromKey(t *testing.T) {
	tests := []struct {
		name    string
		key     [2]int32
		wantMin mcmath.Vec3
		wantMax mcmath.Vec3
	}{
		{
			name:    "origin chunk",
			key:     [2]int32{0, 0},
			wantMin: mcmath.Vec3{X: 0, Y: 0, Z: 0},
			wantMax: mcmath.Vec3{X: 16, Y: 256, Z: 16},
		},
		{
			name:    "positive chunk",
			key:     [2]int32{2, 3},
			wantMin: mcmath.Vec3{X: 32, Y: 0, Z: 48},
			wantMax: mcmath.Vec3{X: 48, Y: 256, Z: 64},
		},
		{
			name:    "negative chunk",
			key:     [2]int32{-1, -2},
			wantMin: mcmath.Vec3{X: -16, Y: 0, Z: -32},
			wantMax: mcmath.Vec3{X: 0, Y: 256, Z: -16},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aabb := chunkAABBFromKey(tc.key)
			if aabb.Min != tc.wantMin {
				t.Errorf("min: got %v, want %v", aabb.Min, tc.wantMin)
			}
			if aabb.Max != tc.wantMax {
				t.Errorf("max: got %v, want %v", aabb.Max, tc.wantMax)
			}
		})
	}
}
