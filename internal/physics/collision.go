package physics

import (
	"math"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ResolveCollision performs axis-sweep collision resolution for a body over the
// given time step dt.  getBlockAABBs must return all solid block AABBs that
// overlap the supplied region.
//
// The Y axis is resolved first so that OnGround can be determined before
// applying horizontal drag.
func ResolveCollision(body *Body, dt float32, getBlockAABBs func(mcmath.AABB) []mcmath.AABB) {
	if body.NoClip {
		body.Position = body.Position.Add(body.Velocity.Scale(dt))
		return
	}

	// Apply gravity.
	body.Velocity.Y -= body.Gravity * dt

	// Clamp fall speed (velocity.Y is negative when falling).
	if body.Velocity.Y < -body.MaxFallSpeed {
		body.Velocity.Y = -body.MaxFallSpeed
	}

	body.OnGround = false

	// --- Y axis ---
	moveY := body.Velocity.Y * dt
	body.Position.Y += moveY

	worldAABB := body.WorldAABB()
	blocks := getBlockAABBs(worldAABB)
	for _, block := range blocks {
		if !worldAABB.Intersects(block) {
			continue
		}
		pen := penetration(worldAABB, block, 1) // axis Y
		if pen == 0 {
			continue
		}
		body.Position.Y += pen
		worldAABB = body.WorldAABB()
		if pen > 0 {
			// Pushed upward -> standing on ground.
			body.OnGround = true
		}
		body.Velocity.Y = 0
	}

	// --- X axis ---
	moveX := body.Velocity.X * dt
	body.Position.X += moveX

	worldAABB = body.WorldAABB()
	blocks = getBlockAABBs(worldAABB)
	for _, block := range blocks {
		if !worldAABB.Intersects(block) {
			continue
		}
		pen := penetration(worldAABB, block, 0) // axis X
		if pen == 0 {
			continue
		}
		body.Position.X += pen
		worldAABB = body.WorldAABB()
		body.Velocity.X = 0
	}

	// --- Z axis ---
	moveZ := body.Velocity.Z * dt
	body.Position.Z += moveZ

	worldAABB = body.WorldAABB()
	blocks = getBlockAABBs(worldAABB)
	for _, block := range blocks {
		if !worldAABB.Intersects(block) {
			continue
		}
		pen := penetration(worldAABB, block, 2) // axis Z
		if pen == 0 {
			continue
		}
		body.Position.Z += pen
		worldAABB = body.WorldAABB()
		body.Velocity.Z = 0
	}

	// Apply drag.
	drag := body.Drag
	if body.OnGround {
		drag = DefaultGroundDrag
	}
	body.Velocity.X *= 1 - drag
	body.Velocity.Z *= 1 - drag
}

// penetration returns the signed distance needed to push body out of block
// along the given axis (0=X, 1=Y, 2=Z). A positive value means push in the
// positive direction and vice-versa.
func penetration(body, block mcmath.AABB, axis int) float32 {
	var bMin, bMax, oMin, oMax float32
	switch axis {
	case 0:
		bMin, bMax = body.Min.X, body.Max.X
		oMin, oMax = block.Min.X, block.Max.X
	case 1:
		bMin, bMax = body.Min.Y, body.Max.Y
		oMin, oMax = block.Min.Y, block.Max.Y
	case 2:
		bMin, bMax = body.Min.Z, body.Max.Z
		oMin, oMax = block.Min.Z, block.Max.Z
	}

	// Positive direction push (body's min is inside block).
	pushPos := oMax - bMin
	// Negative direction push (body's max is inside block).
	pushNeg := oMin - bMax

	if float32(math.Abs(float64(pushPos))) < float32(math.Abs(float64(pushNeg))) {
		return pushPos
	}
	return pushNeg
}
