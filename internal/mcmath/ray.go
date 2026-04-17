package mcmath

import "math"

// Ray represents a ray with an origin and direction in 3D space.
type Ray struct {
	Origin    Vec3
	Direction Vec3
}

// At returns the point along the ray at parameter t.
func (r Ray) At(t float32) Vec3 {
	return r.Origin.Add(r.Direction.Scale(t))
}

// CastBlocks performs a DDA (Digital Differential Analyzer) voxel traversal
// along the ray up to maxDist, testing each visited block with isBlock.
// Returns whether a block was hit, the block position, the face direction
// the ray entered from, and the t parameter of the hit.
func (r Ray) CastBlocks(maxDist float32, isBlock func(BlockPos) bool) (hit bool, pos BlockPos, face Direction, t float32) {
	// Current voxel position
	x := int32(math.Floor(float64(r.Origin.X)))
	y := int32(math.Floor(float64(r.Origin.Y)))
	z := int32(math.Floor(float64(r.Origin.Z)))

	dx := float64(r.Direction.X)
	dy := float64(r.Direction.Y)
	dz := float64(r.Direction.Z)

	// Step direction (+1 or -1)
	var stepX, stepY, stepZ int32
	// tMax: how far along the ray we must move (in t units) to cross the first
	// voxel boundary in each axis.
	var tMaxX, tMaxY, tMaxZ float64
	// tDelta: how far along the ray (in t units) to cross one full voxel in each axis.
	var tDeltaX, tDeltaY, tDeltaZ float64

	const inf = 1e30

	if dx > 0 {
		stepX = 1
		tMaxX = (float64(x+1) - float64(r.Origin.X)) / dx
		tDeltaX = 1.0 / dx
	} else if dx < 0 {
		stepX = -1
		tMaxX = (float64(x) - float64(r.Origin.X)) / dx
		tDeltaX = -1.0 / dx
	} else {
		stepX = 0
		tMaxX = inf
		tDeltaX = inf
	}

	if dy > 0 {
		stepY = 1
		tMaxY = (float64(y+1) - float64(r.Origin.Y)) / dy
		tDeltaY = 1.0 / dy
	} else if dy < 0 {
		stepY = -1
		tMaxY = (float64(y) - float64(r.Origin.Y)) / dy
		tDeltaY = -1.0 / dy
	} else {
		stepY = 0
		tMaxY = inf
		tDeltaY = inf
	}

	if dz > 0 {
		stepZ = 1
		tMaxZ = (float64(z+1) - float64(r.Origin.Z)) / dz
		tDeltaZ = 1.0 / dz
	} else if dz < 0 {
		stepZ = -1
		tMaxZ = (float64(z) - float64(r.Origin.Z)) / dz
		tDeltaZ = -1.0 / dz
	} else {
		stepZ = 0
		tMaxZ = inf
		tDeltaZ = inf
	}

	maxDistF := float64(maxDist)
	face = North // default; will be overwritten on hit

	// Check the starting block
	startPos := BlockPos{x, y, z}
	if isBlock(startPos) {
		return true, startPos, face, 0
	}

	for {
		// Advance in the axis with the smallest tMax
		var currentT float64
		if tMaxX < tMaxY {
			if tMaxX < tMaxZ {
				currentT = tMaxX
				x += stepX
				tMaxX += tDeltaX
				if stepX > 0 {
					face = West // ray entered from -X face
				} else {
					face = East // ray entered from +X face
				}
			} else {
				currentT = tMaxZ
				z += stepZ
				tMaxZ += tDeltaZ
				if stepZ > 0 {
					face = North // ray entered from -Z face
				} else {
					face = South // ray entered from +Z face
				}
			}
		} else {
			if tMaxY < tMaxZ {
				currentT = tMaxY
				y += stepY
				tMaxY += tDeltaY
				if stepY > 0 {
					face = Down // ray entered from -Y face
				} else {
					face = Up // ray entered from +Y face
				}
			} else {
				currentT = tMaxZ
				z += stepZ
				tMaxZ += tDeltaZ
				if stepZ > 0 {
					face = North
				} else {
					face = South
				}
			}
		}

		if currentT > maxDistF {
			return false, BlockPos{}, North, 0
		}

		bp := BlockPos{x, y, z}
		if isBlock(bp) {
			return true, bp, face, float32(currentT)
		}
	}
}
