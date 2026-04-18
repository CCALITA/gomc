//go:build !ci

package render

import (
	"math"
	"testing"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// ---------------------------------------------------------------------------
// EntityTypeColor tests
// ---------------------------------------------------------------------------

func TestEntityTypeColor_AllKnownTypes(t *testing.T) {
	tests := []struct {
		name       string
		entityType uint8
		wantR      float32
		wantG      float32
		wantB      float32
	}{
		{"Player", entity.TypePlayer, 0.8, 0.6, 0.4},
		{"Zombie", entity.TypeZombie, 0.3, 0.6, 0.2},
		{"Skeleton", entity.TypeSkeleton, 0.85, 0.82, 0.75},
		{"Cow", entity.TypeCow, 0.4, 0.25, 0.1},
		{"Pig", entity.TypePig, 0.9, 0.6, 0.6},
		{"Sheep", entity.TypeSheep, 0.9, 0.9, 0.85},
		{"Chicken", entity.TypeChicken, 0.95, 0.95, 0.9},
		{"Item", entity.TypeItem, 1.0, 0.9, 0.2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			color := EntityTypeColor(tc.entityType)
			if color[0] != tc.wantR || color[1] != tc.wantG || color[2] != tc.wantB {
				t.Errorf("EntityTypeColor(%d) = [%f, %f, %f], want [%f, %f, %f]",
					tc.entityType, color[0], color[1], color[2], tc.wantR, tc.wantG, tc.wantB)
			}
		})
	}
}

func TestEntityTypeColor_UnknownType(t *testing.T) {
	color := EntityTypeColor(255)
	expected := defaultEntityColor
	if color != expected {
		t.Errorf("EntityTypeColor(255) = %v, want default %v", color, expected)
	}
}

// ---------------------------------------------------------------------------
// GenerateBoxVertices tests
// ---------------------------------------------------------------------------

func TestGenerateBoxVertices_Count(t *testing.T) {
	aabb := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 2, Z: 1},
	}
	color := [3]float32{1, 0, 0}

	vertices, indices := GenerateBoxVertices(aabb, color)

	// 8 vertices * 6 floats = 48 floats
	if len(vertices) != 48 {
		t.Errorf("expected 48 floats (8 vertices * 6), got %d", len(vertices))
	}

	// 12 triangles * 3 indices = 36 indices
	if len(indices) != 36 {
		t.Errorf("expected 36 indices (12 triangles), got %d", len(indices))
	}
}

func TestGenerateBoxVertices_VertexPositionsMatchAABB(t *testing.T) {
	aabb := mcmath.AABB{
		Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
		Max: mcmath.Vec3{X: 0.3, Y: 1.8, Z: 0.3},
	}
	color := [3]float32{0.8, 0.6, 0.4}

	vertices, _ := GenerateBoxVertices(aabb, color)

	for i := 0; i < len(vertices); i += 6 {
		x, y, z := vertices[i], vertices[i+1], vertices[i+2]
		vertIdx := i / 6

		if math.Abs(float64(x-aabb.Min.X)) > 1e-5 && math.Abs(float64(x-aabb.Max.X)) > 1e-5 {
			t.Errorf("vertex %d: X=%f does not match min.X=%f or max.X=%f",
				vertIdx, x, aabb.Min.X, aabb.Max.X)
		}
		if math.Abs(float64(y-aabb.Min.Y)) > 1e-5 && math.Abs(float64(y-aabb.Max.Y)) > 1e-5 {
			t.Errorf("vertex %d: Y=%f does not match min.Y=%f or max.Y=%f",
				vertIdx, y, aabb.Min.Y, aabb.Max.Y)
		}
		if math.Abs(float64(z-aabb.Min.Z)) > 1e-5 && math.Abs(float64(z-aabb.Max.Z)) > 1e-5 {
			t.Errorf("vertex %d: Z=%f does not match min.Z=%f or max.Z=%f",
				vertIdx, z, aabb.Min.Z, aabb.Max.Z)
		}
	}
}

func TestGenerateBoxVertices_IndicesValid(t *testing.T) {
	aabb := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	color := [3]float32{0.5, 0.5, 0.5}

	_, indices := GenerateBoxVertices(aabb, color)

	for i, idx := range indices {
		if idx >= 8 {
			t.Errorf("index[%d] = %d, which is >= 8 (out of bounds)", i, idx)
		}
	}
}

func TestGenerateBoxVertices_ColorsApplied(t *testing.T) {
	aabb := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	color := [3]float32{0.3, 0.6, 0.2}

	vertices, _ := GenerateBoxVertices(aabb, color)

	for i := 0; i < len(vertices); i += 6 {
		r, g, b := vertices[i+3], vertices[i+4], vertices[i+5]
		vertIdx := i / 6
		if r != color[0] || g != color[1] || b != color[2] {
			t.Errorf("vertex %d colour = [%f, %f, %f], want [%f, %f, %f]",
				vertIdx, r, g, b, color[0], color[1], color[2])
		}
	}
}

// ---------------------------------------------------------------------------
// DrawEntities integration test
// ---------------------------------------------------------------------------

func TestDrawEntities_GeneratesVertexData(t *testing.T) {
	w := ecs.NewWorld()

	pos := mcmath.Vec3{X: 10, Y: 64, Z: 20}
	e := w.NewEntity()

	ecs.GetStore[entity.Transform](w).Set(e, entity.Transform{Position: pos})
	ecs.GetStore[entity.EntityTypeComp](w).Set(e, entity.EntityTypeComp{Type: entity.TypeZombie})

	bbox := mcmath.AABB{
		Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
		Max: mcmath.Vec3{X: 0.3, Y: 1.95, Z: 0.3},
	}
	body := physics.NewBody(bbox)
	body.Position = pos
	ecs.GetStore[entity.PhysicsBody](w).Set(e, entity.PhysicsBody{Body: &body})

	er := NewEntityRenderer()
	er.DrawEntities(w)

	if len(er.Vertices) != 48 {
		t.Errorf("expected 48 vertex floats, got %d", len(er.Vertices))
	}
	if len(er.Indices) != 36 {
		t.Errorf("expected 36 indices, got %d", len(er.Indices))
	}
}

func TestDrawEntities_MultipleEntities(t *testing.T) {
	w := ecs.NewWorld()

	bbox := mcmath.AABB{
		Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
		Max: mcmath.Vec3{X: 0.3, Y: 1.8, Z: 0.3},
	}

	for i := 0; i < 3; i++ {
		pos := mcmath.Vec3{X: float32(i * 10), Y: 64, Z: 0}
		e := w.NewEntity()

		ecs.GetStore[entity.Transform](w).Set(e, entity.Transform{Position: pos})
		ecs.GetStore[entity.EntityTypeComp](w).Set(e, entity.EntityTypeComp{Type: entity.TypePlayer})

		body := physics.NewBody(bbox)
		body.Position = pos
		ecs.GetStore[entity.PhysicsBody](w).Set(e, entity.PhysicsBody{Body: &body})
	}

	er := NewEntityRenderer()
	er.DrawEntities(w)

	// 3 entities: 3 * 48 = 144 floats, 3 * 36 = 108 indices.
	if len(er.Vertices) != 144 {
		t.Errorf("expected 144 vertex floats for 3 entities, got %d", len(er.Vertices))
	}
	if len(er.Indices) != 108 {
		t.Errorf("expected 108 indices for 3 entities, got %d", len(er.Indices))
	}

	// Verify index offsets: second entity indices should reference vertices 8..15.
	foundOffset := false
	for _, idx := range er.Indices {
		if idx >= 8 && idx < 16 {
			foundOffset = true
			break
		}
	}
	if !foundOffset {
		t.Error("expected indices for second entity to be offset by 8, none found in range [8,16)")
	}
}

func TestDrawEntities_EmptyWorld(t *testing.T) {
	w := ecs.NewWorld()

	er := NewEntityRenderer()
	er.DrawEntities(w)

	if len(er.Vertices) != 0 {
		t.Errorf("expected 0 vertices for empty world, got %d", len(er.Vertices))
	}
	if len(er.Indices) != 0 {
		t.Errorf("expected 0 indices for empty world, got %d", len(er.Indices))
	}
}

func TestEntityRenderer_Cleanup(t *testing.T) {
	er := NewEntityRenderer()
	er.Vertices = make([]float32, 100)
	er.Indices = make([]uint32, 50)

	er.Cleanup()

	if er.Vertices != nil {
		t.Error("expected Vertices to be nil after Cleanup")
	}
	if er.Indices != nil {
		t.Error("expected Indices to be nil after Cleanup")
	}
}
