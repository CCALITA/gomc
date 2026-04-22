package chunk

import "github.com/fanxiyao/gomc/internal/mcmath"

// ChunkMesh holds the generated mesh data for a chunk.
type ChunkMesh struct {
	Vertices []float32 // interleaved: x, y, z, nx, ny, nz, u, v, ao (9 floats per vertex)
	Indices  []uint32
}

// Stride is the number of float32 values per vertex.
const Stride = 9

// axis identifies the perpendicular axis for a slicing pass.
type axis int

const (
	axisX axis = iota
	axisY
	axisZ
)

// MeshChunk generates a ChunkMesh using greedy meshing with ambient occlusion.
//
// neighbors are the four horizontal neighbors in order: [North, South, East, West]
// (matching mcmath.North=0, South=1, East=2, West=3). Nil entries mean the
// neighbor chunk is not loaded (faces on that boundary are always emitted).
//
// isSolid reports whether a block ID is solid (opaque) — its faces can be
// culled when adjacent to another solid block.
//
// isTransparent reports whether a block ID is transparent (air, glass, etc.) —
// faces are exposed when a solid block is adjacent to a transparent one.
func MeshChunk(
	chunk *Chunk,
	neighbors [4]*Chunk,
	isSolid func(uint16) bool,
	isTransparent func(uint16) bool,
) *ChunkMesh {
	m := &ChunkMesh{}

	for si := 0; si < numSections; si++ {
		sec := chunk.GetSection(si)
		if sec == nil || sec.IsEmpty() {
			continue
		}
		baseY := si * mcmath.SectionHeight

		meshSection(m, chunk, neighbors, baseY, isSolid, isTransparent)
	}

	return m
}

// faceDir describes one of the six face directions for greedy meshing.
type faceDir struct {
	ax             axis
	sign           int // +1 or -1
	nx, ny, nz     float32
}

// mergedQuad represents a rectangle found by greedy merging.
type mergedQuad struct {
	u, v, w, h int
}

// meshSection performs greedy meshing for a single 16x16x16 section.
func meshSection(
	m *ChunkMesh,
	chunk *Chunk,
	neighbors [4]*Chunk,
	baseY int,
	isSolid func(uint16) bool,
	isTransparent func(uint16) bool,
) {
	const N = mcmath.ChunkSize

	getBlock := makeGetBlock(chunk, neighbors)

	dirs := [6]faceDir{
		{axisX, 1, 1, 0, 0},   // East  (+X)
		{axisX, -1, -1, 0, 0}, // West  (-X)
		{axisY, 1, 0, 1, 0},   // Up    (+Y)
		{axisY, -1, 0, -1, 0}, // Down  (-Y)
		{axisZ, 1, 0, 0, 1},   // South (+Z)
		{axisZ, -1, 0, 0, -1}, // North (-Z)
	}

	var mask [N * N]uint16

	for _, d := range dirs {
		for layer := 0; layer < N; layer++ {
			buildFaceMask(&mask, chunk, getBlock, d, layer, baseY, isSolid, isTransparent)
			quads := greedyMerge(&mask)
			for _, q := range quads {
				emitQuad(m, chunk, neighbors, d.ax, d.sign, layer, q.u, q.v, q.w, q.h, baseY,
					d.nx, d.ny, d.nz, getBlock, isSolid)
			}
		}
	}
}

// makeGetBlock returns a block-fetching closure that crosses into neighbor chunks.
func makeGetBlock(c *Chunk, neighbors [4]*Chunk) func(int, int, int) uint16 {
	const N = mcmath.ChunkSize
	return func(x, y, z int) uint16 {
		if y < 0 || y >= mcmath.ChunkHeight {
			return 0
		}
		if x >= 0 && x < N && z >= 0 && z < N {
			return c.GetBlock(x, y, z)
		}
		if z < 0 && x >= 0 && x < N {
			if neighbors[0] != nil {
				return neighbors[0].GetBlock(x, y, z+N)
			}
			return 0
		}
		if z >= N && x >= 0 && x < N {
			if neighbors[1] != nil {
				return neighbors[1].GetBlock(x, y, z-N)
			}
			return 0
		}
		if x >= N && z >= 0 && z < N {
			if neighbors[2] != nil {
				return neighbors[2].GetBlock(x-N, y, z)
			}
			return 0
		}
		if x < 0 && z >= 0 && z < N {
			if neighbors[3] != nil {
				return neighbors[3].GetBlock(x+N, y, z)
			}
			return 0
		}
		return 0
	}
}

// buildFaceMask populates the 2D mask of exposed faces for one slice/layer.
func buildFaceMask(
	mask *[mcmath.ChunkSize * mcmath.ChunkSize]uint16,
	c *Chunk,
	getBlock func(int, int, int) uint16,
	d faceDir,
	layer, baseY int,
	isSolid func(uint16) bool,
	isTransparent func(uint16) bool,
) {
	const N = mcmath.ChunkSize
	for v := 0; v < N; v++ {
		for u := 0; u < N; u++ {
			var bx, by, bz int
			var nx, ny, nz int
			switch d.ax {
			case axisX:
				bx, by, bz = layer, baseY+v, u
				nx, ny, nz = d.sign, 0, 0
			case axisY:
				bx, by, bz = u, baseY+layer, v
				nx, ny, nz = 0, d.sign, 0
			case axisZ:
				bx, by, bz = u, baseY+v, layer
				nx, ny, nz = 0, 0, d.sign
			}

			blockID := c.GetBlock(bx, by, bz)
			if blockID == 0 || !isSolid(blockID) {
				mask[v*N+u] = 0
				continue
			}
			neighborID := getBlock(bx+nx, by+ny, bz+nz)
			if isTransparent(neighborID) {
				mask[v*N+u] = blockID
			} else {
				mask[v*N+u] = 0
			}
		}
	}
}

// greedyMerge runs greedy rectangle merging on a 16x16 mask and returns the
// merged quads.
func greedyMerge(mask *[mcmath.ChunkSize * mcmath.ChunkSize]uint16) []mergedQuad {
	const N = mcmath.ChunkSize
	var done [N * N]bool
	var quads []mergedQuad

	for v := 0; v < N; v++ {
		for u := 0; u < N; u++ {
			idx := v*N + u
			if mask[idx] == 0 || done[idx] {
				continue
			}
			bid := mask[idx]

			w := 1
			for u+w < N && mask[v*N+u+w] == bid && !done[v*N+u+w] {
				w++
			}
			h := 1
		outer:
			for v+h < N {
				for du := 0; du < w; du++ {
					ni := (v+h)*N + u + du
					if mask[ni] != bid || done[ni] {
						break outer
					}
				}
				h++
			}

			for dv := 0; dv < h; dv++ {
				for du := 0; du < w; du++ {
					done[(v+dv)*N+u+du] = true
				}
			}

			quads = append(quads, mergedQuad{u: u, v: v, w: w, h: h})
		}
	}

	return quads
}

// vertex3 is a simple 3D position used during mesh generation.
type vertex3 struct{ x, y, z float32 }

// emitQuad adds four vertices and six indices (two triangles) for a merged face.
func emitQuad(
	m *ChunkMesh,
	chunk *Chunk,
	neighbors [4]*Chunk,
	ax axis,
	sign int,
	layer, u, v, w, h, baseY int,
	normalX, normalY, normalZ float32,
	getBlock func(int, int, int) uint16,
	isSolid func(uint16) bool,
) {
	const N = mcmath.ChunkSize

	// Convert (layer, u, v) back to world-local XYZ for the four corners.
	// Offset the face by +1 along the axis when sign is positive.
	offset := 0
	if sign > 0 {
		offset = 1
	}

	var corners [4]vertex3

	switch ax {
	case axisX:
		fx := float32(layer + offset)
		corners[0] = vertex3{fx, float32(baseY + v), float32(u)}
		corners[1] = vertex3{fx, float32(baseY + v), float32(u + w)}
		corners[2] = vertex3{fx, float32(baseY + v + h), float32(u + w)}
		corners[3] = vertex3{fx, float32(baseY + v + h), float32(u)}
	case axisY:
		fy := float32(baseY + layer + offset)
		corners[0] = vertex3{float32(u), fy, float32(v)}
		corners[1] = vertex3{float32(u + w), fy, float32(v)}
		corners[2] = vertex3{float32(u + w), fy, float32(v + h)}
		corners[3] = vertex3{float32(u), fy, float32(v + h)}
	case axisZ:
		fz := float32(layer + offset)
		corners[0] = vertex3{float32(u), float32(baseY + v), fz}
		corners[1] = vertex3{float32(u + w), float32(baseY + v), fz}
		corners[2] = vertex3{float32(u + w), float32(baseY + v + h), fz}
		corners[3] = vertex3{float32(u), float32(baseY + v + h), fz}
	}

	// Fix winding order: swap corners[1] and [3] for faces where the default
	// corner layout produces CW winding (viewed from the outward normal).
	if (ax == axisX && sign > 0) || (ax == axisY && sign > 0) || (ax == axisZ && sign < 0) {
		corners[1], corners[3] = corners[3], corners[1]
	}

	// UV coordinates based on quad dimensions.
	uvs := [4][2]float32{
		{0, 0},
		{float32(w), 0},
		{float32(w), float32(h)},
		{0, float32(h)},
	}

	// Match UV swap to corner swap.
	if (ax == axisX && sign > 0) || (ax == axisY && sign > 0) || (ax == axisZ && sign < 0) {
		uvs[1], uvs[3] = uvs[3], uvs[1]
	}

	// Compute ambient occlusion for each corner.
	ao := computeAO(corners, normalX, normalY, normalZ, getBlock, isSolid)

	base := uint32(len(m.Vertices) / Stride)

	for i := 0; i < 4; i++ {
		aoVal := float32(3-ao[i]) / 3.0 // 0=darkest, 1=brightest
		m.Vertices = append(m.Vertices,
			corners[i].x, corners[i].y, corners[i].z,
			normalX, normalY, normalZ,
			uvs[i][0], uvs[i][1],
			aoVal,
		)
	}

	// Choose triangle winding to avoid AO interpolation artifacts:
	// flip the diagonal when ao[0]+ao[2] > ao[1]+ao[3].
	if ao[0]+ao[2] > ao[1]+ao[3] {
		m.Indices = append(m.Indices,
			base+1, base+2, base+3,
			base+3, base+0, base+1,
		)
	} else {
		m.Indices = append(m.Indices,
			base+0, base+1, base+2,
			base+2, base+3, base+0,
		)
	}
}

// computeAO calculates ambient occlusion values (0-3) for the four corners
// of a quad. For each corner, it checks three neighbors: the two edge-adjacent
// blocks and the corner-diagonal block. AO = side1 + side2 + corner (clamped
// to 3 when both sides are solid, blocking the corner).
func computeAO(
	corners [4]vertex3,
	nx, ny, nz float32,
	getBlock func(int, int, int) uint16,
	isSolid func(uint16) bool,
) [4]int {
	var result [4]int

	// Determine the two tangent directions based on the normal.
	type iv3 struct{ x, y, z int }
	var t1, t2 iv3
	if nx != 0 {
		t1 = iv3{0, 1, 0}
		t2 = iv3{0, 0, 1}
	} else if ny != 0 {
		t1 = iv3{1, 0, 0}
		t2 = iv3{0, 0, 1}
	} else {
		t1 = iv3{1, 0, 0}
		t2 = iv3{0, 1, 0}
	}

	ni := iv3{int(nx), int(ny), int(nz)}

	// For each corner, determine the sign along t1 and t2.
	// Corners are ordered: (0,0), (w,0), (w,h), (0,h) — but for AO we only
	// care about the unit offset directions at the corner, not the quad size.
	// We use the corner position minus the face center to determine direction.
	cx := (corners[0].x + corners[2].x) / 2
	cy := (corners[0].y + corners[2].y) / 2
	cz := (corners[0].z + corners[2].z) / 2

	for i := 0; i < 4; i++ {
		// Direction from face center to corner.
		dx := corners[i].x - cx
		dy := corners[i].y - cy
		dz := corners[i].z - cz

		// Project onto tangent directions to get sign.
		s1 := sign(float32(t1.x)*dx + float32(t1.y)*dy + float32(t1.z)*dz)
		s2 := sign(float32(t2.x)*dx + float32(t2.y)*dy + float32(t2.z)*dz)

		// The block we sample is offset from the corner in the normal direction,
		// shifted inward by 0.5 so we land in the right voxel.
		// Base position: corner - 0.5*normal (to get the block the face sits on)
		// then we offset by t1*s1 and t2*s2 to get edge and corner neighbors.
		bx := int(corners[i].x - 0.5*nx)
		by := int(corners[i].y - 0.5*ny)
		bz := int(corners[i].z - 0.5*nz)

		// Edge neighbors (along the face, offset into the normal direction).
		side1Block := getBlock(bx+ni.x+t1.x*s1, by+ni.y+t1.y*s1, bz+ni.z+t1.z*s1)
		side2Block := getBlock(bx+ni.x+t2.x*s2, by+ni.y+t2.y*s2, bz+ni.z+t2.z*s2)
		cornerBlock := getBlock(bx+ni.x+t1.x*s1+t2.x*s2, by+ni.y+t1.y*s1+t2.y*s2, bz+ni.z+t1.z*s1+t2.z*s2)

		s1v := boolToInt(isSolid(side1Block))
		s2v := boolToInt(isSolid(side2Block))
		cv := boolToInt(isSolid(cornerBlock))

		if s1v == 1 && s2v == 1 {
			result[i] = 3
		} else {
			result[i] = s1v + s2v + cv
		}
	}

	return result
}

func sign(f float32) int {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
