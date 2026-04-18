package render

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// EntityVertexStride is the byte stride for an entity vertex:
// position(vec3) + color(vec3) = 6 floats. Exported for pipeline creation.
const EntityVertexStride = 6 * 4 // 24 bytes

// entityFloatsPerVertex is the number of float32 values per entity vertex.
const entityFloatsPerVertex = 6

// entityTypeColors maps entity type constants to RGB colours.
var entityTypeColors = map[uint8][3]float32{
	entity.TypePlayer:   {0.8, 0.6, 0.4},
	entity.TypeZombie:   {0.3, 0.6, 0.2},
	entity.TypeSkeleton: {0.85, 0.82, 0.75},
	entity.TypeCow:      {0.4, 0.25, 0.1},
	entity.TypePig:      {0.9, 0.6, 0.6},
	entity.TypeSheep:    {0.9, 0.9, 0.85},
	entity.TypeChicken:  {0.95, 0.95, 0.9},
	entity.TypeItem:     {1.0, 0.9, 0.2},
}

// defaultEntityColor is used when the entity type has no assigned colour.
var defaultEntityColor = [3]float32{0.5, 0.5, 0.5}

// boxIndices is the shared index template for a cube with 8 vertices.
// 6 faces, 2 triangles each = 36 indices (counter-clockwise winding).
var boxIndices = [36]uint32{
	0, 2, 1, 0, 3, 2, // Front  (z = min)
	4, 5, 6, 4, 6, 7, // Back   (z = max)
	0, 4, 7, 0, 7, 3, // Left   (x = min)
	1, 2, 6, 1, 6, 5, // Right  (x = max)
	0, 1, 5, 0, 5, 4, // Bottom (y = min)
	3, 7, 6, 3, 6, 2, // Top    (y = max)
}

// EntityRenderer manages entity rendering data preparation.
type EntityRenderer struct {
	// Vertices and Indices are populated by DrawEntities and can be read
	// by the caller for uploading to the GPU.
	Vertices []float32
	Indices  []uint32
}

// NewEntityRenderer creates an EntityRenderer.
func NewEntityRenderer() *EntityRenderer {
	return &EntityRenderer{}
}

// EntityTypeColor returns the RGB colour for the given entity type.
func EntityTypeColor(entityType uint8) [3]float32 {
	if c, ok := entityTypeColors[entityType]; ok {
		return c
	}
	return defaultEntityColor
}

// GenerateBoxVertices generates a coloured cube mesh from an AABB.
// Vertex format: position(3) + colour(3) = 6 floats per vertex.
// Returns 8 vertices (48 floats) and 36 indices (12 triangles).
func GenerateBoxVertices(aabb mcmath.AABB, color [3]float32) ([]float32, []uint32) {
	min := aabb.Min
	max := aabb.Max

	// 8 corners of the AABB.
	corners := [8][3]float32{
		{min.X, min.Y, min.Z}, // 0
		{max.X, min.Y, min.Z}, // 1
		{max.X, max.Y, min.Z}, // 2
		{min.X, max.Y, min.Z}, // 3
		{min.X, min.Y, max.Z}, // 4
		{max.X, min.Y, max.Z}, // 5
		{max.X, max.Y, max.Z}, // 6
		{min.X, max.Y, max.Z}, // 7
	}

	vertices := make([]float32, 0, 8*entityFloatsPerVertex)
	for _, c := range corners {
		vertices = append(vertices, c[0], c[1], c[2], color[0], color[1], color[2])
	}

	indices := make([]uint32, len(boxIndices))
	copy(indices, boxIndices[:])

	return vertices, indices
}

// DrawEntities queries all entities with Transform, EntityTypeComp, and
// PhysicsBody components, and generates box vertices at each entity's world
// position using its AABB dimensions and type colour. The combined vertex
// and index data is stored in the EntityRenderer's Vertices and Indices
// fields.
func (er *EntityRenderer) DrawEntities(ecsWorld *ecs.World) {
	var allVertices []float32
	var allIndices []uint32
	var vertexOffset uint32

	ecs.Query3[entity.Transform, entity.EntityTypeComp, entity.PhysicsBody](
		ecsWorld,
		func(_ ecs.Entity, tf *entity.Transform, et *entity.EntityTypeComp, pb *entity.PhysicsBody) {
			color := EntityTypeColor(et.Type)
			worldAABB := pb.Body.BoundingBox.Offset(tf.Position)

			verts, inds := GenerateBoxVertices(worldAABB, color)

			for _, idx := range inds {
				allIndices = append(allIndices, idx+vertexOffset)
			}
			allVertices = append(allVertices, verts...)
			vertexOffset += uint32(len(verts) / entityFloatsPerVertex)
		},
	)

	er.Vertices = allVertices
	er.Indices = allIndices
}

// Cleanup releases the vertex and index data.
func (er *EntityRenderer) Cleanup() {
	er.Vertices = nil
	er.Indices = nil
}
