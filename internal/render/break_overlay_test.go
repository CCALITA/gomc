//go:build !ci

package render

import (
	"math"
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// Stage from progress tests
// ---------------------------------------------------------------------------

func TestBreakOverlay_GetStage(t *testing.T) {
	tests := []struct {
		name     string
		progress float32
		want     int
	}{
		{name: "zero progress", progress: 0.0, want: 0},
		{name: "half progress", progress: 0.5, want: 5},
		{name: "near complete", progress: 0.99, want: 9},
		{name: "full progress clamped", progress: 1.0, want: 9},
		{name: "over 1.0 clamped", progress: 1.5, want: 9},
		{name: "negative clamped", progress: -0.1, want: 0},
		{name: "stage boundary 0.1", progress: 0.1, want: 1},
		{name: "stage boundary 0.9", progress: 0.9, want: 9},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bo := &BreakOverlay{}
			bo.SetTarget(mcmath.BlockPos{X: 0, Y: 0, Z: 0}, tc.progress)
			got := bo.GetStage()
			if got != tc.want {
				t.Errorf("progress %.2f: got stage %d, want %d", tc.progress, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Set / Clear / HasTarget lifecycle tests
// ---------------------------------------------------------------------------

func TestBreakOverlay_Lifecycle(t *testing.T) {
	bo := &BreakOverlay{}

	if bo.HasTarget() {
		t.Error("new BreakOverlay should not have a target")
	}

	pos := mcmath.BlockPos{X: 5, Y: 64, Z: -3}
	bo.SetTarget(pos, 0.5)

	if !bo.HasTarget() {
		t.Error("expected HasTarget true after SetTarget")
	}

	if stage := bo.GetStage(); stage != 5 {
		t.Errorf("expected stage 5, got %d", stage)
	}

	bo.ClearTarget()

	if bo.HasTarget() {
		t.Error("expected HasTarget false after ClearTarget")
	}

	if stage := bo.GetStage(); stage != 0 {
		t.Errorf("expected stage 0 after clear, got %d", stage)
	}
}

func TestBreakOverlay_SetTarget_OverwritesPrevious(t *testing.T) {
	bo := &BreakOverlay{}

	bo.SetTarget(mcmath.BlockPos{X: 1, Y: 2, Z: 3}, 0.2)
	bo.SetTarget(mcmath.BlockPos{X: 10, Y: 20, Z: 30}, 0.7)

	if !bo.HasTarget() {
		t.Error("expected HasTarget true after second SetTarget")
	}

	if stage := bo.GetStage(); stage != 7 {
		t.Errorf("expected stage 7, got %d", stage)
	}
}

// ---------------------------------------------------------------------------
// Vertex and index generation tests
// ---------------------------------------------------------------------------

func TestGenerateOverlayVertices_VertexCount(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	verts, _ := GenerateOverlayVertices(pos, 0)

	// 24 vertices * 4 floats each = 96 floats
	wantFloats := 24 * 4
	if len(verts) != wantFloats {
		t.Errorf("expected %d floats, got %d", wantFloats, len(verts))
	}

	// Verify vertex count
	wantVertices := 24
	if len(verts)/4 != wantVertices {
		t.Errorf("expected %d vertices, got %d", wantVertices, len(verts)/4)
	}
}

func TestGenerateOverlayVertices_IndexCount(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	_, indices := GenerateOverlayVertices(pos, 0)

	// 6 faces * 6 indices = 36
	wantIndices := 36
	if len(indices) != wantIndices {
		t.Errorf("expected %d indices, got %d", wantIndices, len(indices))
	}
}

func TestGenerateOverlayVertices_AlphaIncreasesWithStage(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}

	prevAlpha := float32(0.0)
	for stage := 0; stage <= 9; stage++ {
		verts, _ := GenerateOverlayVertices(pos, stage)

		// Alpha is the 4th float of the first vertex.
		alpha := verts[3]
		expectedAlpha := float32(stage+1) * 0.08

		if alpha != expectedAlpha {
			t.Errorf("stage %d: expected alpha %.4f, got %.4f", stage, expectedAlpha, alpha)
		}

		if stage > 0 && alpha <= prevAlpha {
			t.Errorf("stage %d: alpha %.4f should be greater than previous %.4f", stage, alpha, prevAlpha)
		}

		// Verify all vertices in this mesh share the same alpha.
		for i := 3; i < len(verts); i += 4 {
			if verts[i] != alpha {
				t.Errorf("stage %d: vertex %d alpha %.4f differs from expected %.4f", stage, i/4, verts[i], alpha)
			}
		}

		prevAlpha = alpha
	}
}

func TestGenerateOverlayVertices_IndicesInRange(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	verts, indices := GenerateOverlayVertices(pos, 5)

	vertexCount := uint32(len(verts) / 4)
	for i, idx := range indices {
		if idx >= vertexCount {
			t.Errorf("index[%d] = %d is out of range (vertex count = %d)", i, idx, vertexCount)
		}
	}
}

func TestGenerateOverlayVertices_OffsetFromBlockFace(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	verts, _ := GenerateOverlayVertices(pos, 0)

	// The overlay should extend slightly beyond [0,1] in each axis due to the
	// 0.001 offset. Check that the min coordinate is -0.001 and max is 1.001.
	minX, maxX := verts[0], verts[0]
	minY, maxY := verts[1], verts[1]
	minZ, maxZ := verts[2], verts[2]

	for i := 0; i < len(verts); i += 4 {
		x, y, z := verts[i], verts[i+1], verts[i+2]
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
		if z < minZ {
			minZ = z
		}
		if z > maxZ {
			maxZ = z
		}
	}

	wantMin := float64(-overlayOffset)
	wantMax := float64(1 + overlayOffset)
	eps := 1e-6

	if math.Abs(float64(minX)-wantMin) > eps {
		t.Errorf("min X: got %f, want %f", minX, wantMin)
	}
	if math.Abs(float64(maxX)-wantMax) > eps {
		t.Errorf("max X: got %f, want %f", maxX, wantMax)
	}
	if math.Abs(float64(minY)-wantMin) > eps {
		t.Errorf("min Y: got %f, want %f", minY, wantMin)
	}
	if math.Abs(float64(maxY)-wantMax) > eps {
		t.Errorf("max Y: got %f, want %f", maxY, wantMax)
	}
	if math.Abs(float64(minZ)-wantMin) > eps {
		t.Errorf("min Z: got %f, want %f", minZ, wantMin)
	}
	if math.Abs(float64(maxZ)-wantMax) > eps {
		t.Errorf("max Z: got %f, want %f", maxZ, wantMax)
	}
}
