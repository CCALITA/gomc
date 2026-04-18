package block

import (
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// BlockFace identifies one of the six faces of a cube-shaped block.
// The values mirror mcmath.Direction so they can be converted freely.
type BlockFace int

// Block face constants for each of the six cube faces.
const (
	FaceNorth  BlockFace = BlockFace(mcmath.North) // -Z
	FaceSouth  BlockFace = BlockFace(mcmath.South) // +Z
	FaceEast   BlockFace = BlockFace(mcmath.East)  // +X
	FaceWest   BlockFace = BlockFace(mcmath.West)  // -X
	FaceTop    BlockFace = BlockFace(mcmath.Up)     // +Y
	FaceBottom BlockFace = BlockFace(mcmath.Down)   // -Y
)

// AllFaces returns all six block faces.
func AllFaces() []BlockFace {
	return []BlockFace{FaceNorth, FaceSouth, FaceEast, FaceWest, FaceTop, FaceBottom}
}

// Direction converts the BlockFace to its corresponding mcmath.Direction.
func (f BlockFace) Direction() mcmath.Direction {
	return mcmath.Direction(f)
}

// atlasSize is the number of tiles per row/column in the texture atlas.
const atlasSize = 16

// tileSize is the UV width and height of a single tile in normalized [0,1] space.
const tileSize = 1.0 / float32(atlasSize)

// FaceUV returns the texture atlas UV coordinates for the given block and
// face. The atlas is a 16x16 grid; each block ID maps to one tile per face.
// The returned values are (u, v, w, h) where (u,v) is the top-left corner
// and (w,h) is the width and height in UV space.
func FaceUV(blockID BlockID, face BlockFace) (u, v, w, h float32) {
	_ = face // all faces share the same tile for now
	col := int(blockID) % atlasSize
	row := int(blockID) / atlasSize
	u = float32(col) * tileSize
	v = float32(row) * tileSize
	w = tileSize
	h = tileSize
	return
}

// FaceVertices returns the four corner vertices (in counter-clockwise winding
// order when viewed from outside the cube) for the specified face of a unit
// block whose minimum corner is at (x, y, z).
func FaceVertices(face BlockFace, x, y, z float32) [4]mcmath.Vec3 {
	x0, y0, z0 := x, y, z
	x1, y1, z1 := x+1, y+1, z+1

	switch face {
	case FaceNorth: // -Z face
		return [4]mcmath.Vec3{
			{X: x1, Y: y0, Z: z0},
			{X: x0, Y: y0, Z: z0},
			{X: x0, Y: y1, Z: z0},
			{X: x1, Y: y1, Z: z0},
		}
	case FaceSouth: // +Z face
		return [4]mcmath.Vec3{
			{X: x0, Y: y0, Z: z1},
			{X: x1, Y: y0, Z: z1},
			{X: x1, Y: y1, Z: z1},
			{X: x0, Y: y1, Z: z1},
		}
	case FaceEast: // +X face
		return [4]mcmath.Vec3{
			{X: x1, Y: y0, Z: z1},
			{X: x1, Y: y0, Z: z0},
			{X: x1, Y: y1, Z: z0},
			{X: x1, Y: y1, Z: z1},
		}
	case FaceWest: // -X face
		return [4]mcmath.Vec3{
			{X: x0, Y: y0, Z: z0},
			{X: x0, Y: y0, Z: z1},
			{X: x0, Y: y1, Z: z1},
			{X: x0, Y: y1, Z: z0},
		}
	case FaceTop: // +Y face
		return [4]mcmath.Vec3{
			{X: x0, Y: y1, Z: z0},
			{X: x0, Y: y1, Z: z1},
			{X: x1, Y: y1, Z: z1},
			{X: x1, Y: y1, Z: z0},
		}
	case FaceBottom: // -Y face
		return [4]mcmath.Vec3{
			{X: x0, Y: y0, Z: z1},
			{X: x0, Y: y0, Z: z0},
			{X: x1, Y: y0, Z: z0},
			{X: x1, Y: y0, Z: z1},
		}
	default:
		return [4]mcmath.Vec3{}
	}
}
