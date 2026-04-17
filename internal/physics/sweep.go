package physics

import "github.com/fanxiyao/gomc/internal/mcmath"

// SweepAABB performs a swept (continuous) collision test between a moving AABB
// and a static AABB.  It returns the normalised time of impact t in [0,1],
// the collision normal, and whether a hit occurred.
//
// The approach uses the Minkowski difference: expand the static box by the
// extents of the moving box and then ray-cast from the moving box's centre.
func SweepAABB(moving, static mcmath.AABB, velocity mcmath.Vec3) (t float32, normal mcmath.Vec3, hit bool) {
	// If there is no movement we cannot sweep.
	if velocity.X == 0 && velocity.Y == 0 && velocity.Z == 0 {
		return 0, mcmath.Vec3{}, false
	}

	// Half-extents of moving box.
	halfW := (moving.Max.X - moving.Min.X) * 0.5
	halfH := (moving.Max.Y - moving.Min.Y) * 0.5
	halfD := (moving.Max.Z - moving.Min.Z) * 0.5

	// Minkowski-expanded static box.
	expanded := mcmath.AABB{
		Min: mcmath.Vec3{
			X: static.Min.X - halfW,
			Y: static.Min.Y - halfH,
			Z: static.Min.Z - halfD,
		},
		Max: mcmath.Vec3{
			X: static.Max.X + halfW,
			Y: static.Max.Y + halfH,
			Z: static.Max.Z + halfD,
		},
	}

	// Ray origin = centre of moving box.
	origin := moving.Center()

	// Slab intersection against the expanded box.
	var tNear, tFar float32 = 0, 1
	var hitNormal mcmath.Vec3

	axes := [3][2]float32{
		{origin.X, velocity.X},
		{origin.Y, velocity.Y},
		{origin.Z, velocity.Z},
	}
	mins := [3]float32{expanded.Min.X, expanded.Min.Y, expanded.Min.Z}
	maxs := [3]float32{expanded.Max.X, expanded.Max.Y, expanded.Max.Z}

	for i := 0; i < 3; i++ {
		o := axes[i][0]
		d := axes[i][1]
		mn := mins[i]
		mx := maxs[i]

		if d == 0 {
			// Parallel to slab — check if origin is inside.
			if o < mn || o > mx {
				return 0, mcmath.Vec3{}, false
			}
			continue
		}

		invD := 1.0 / d
		t1 := (mn - o) * invD
		t2 := (mx - o) * invD

		// n is the normal for the near plane.
		var n mcmath.Vec3
		switch i {
		case 0:
			n = mcmath.Vec3{X: -1}
		case 1:
			n = mcmath.Vec3{Y: -1}
		case 2:
			n = mcmath.Vec3{Z: -1}
		}

		if t1 > t2 {
			t1, t2 = t2, t1
			// Flip normal.
			n = n.Scale(-1)
		}

		if t1 > tNear {
			tNear = t1
			hitNormal = n
		}
		if t2 < tFar {
			tFar = t2
		}

		if tNear > tFar {
			return 0, mcmath.Vec3{}, false
		}
	}

	// Must be within the sweep range [0,1].
	if tNear < 0 || tNear > 1 {
		return 0, mcmath.Vec3{}, false
	}

	// Guard against very small / negative numerical margin.
	if tFar < -1e-6 {
		return 0, mcmath.Vec3{}, false
	}

	return tNear, hitNormal, true
}
