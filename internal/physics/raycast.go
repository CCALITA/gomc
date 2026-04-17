package physics

import "github.com/fanxiyao/gomc/internal/mcmath"

// RaycastBlocks casts a ray through the voxel grid using DDA traversal and
// returns the first solid block hit within maxDist.
//
// isSolid reports whether the block at the given position is solid.
func RaycastBlocks(
	origin, direction mcmath.Vec3,
	maxDist float32,
	isSolid func(mcmath.BlockPos) bool,
) (hit bool, pos mcmath.BlockPos, face mcmath.Direction, t float32) {
	ray := mcmath.Ray{
		Origin:    origin,
		Direction: direction.Normalize(),
	}
	return ray.CastBlocks(maxDist, isSolid)
}
