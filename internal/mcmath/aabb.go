package mcmath

import "fmt"

// AABB represents an axis-aligned bounding box defined by minimum and maximum corners.
type AABB struct {
	Min, Max Vec3
}

// Intersects returns true if this AABB overlaps with other.
func (a AABB) Intersects(other AABB) bool {
	return a.Min.X < other.Max.X && a.Max.X > other.Min.X &&
		a.Min.Y < other.Max.Y && a.Max.Y > other.Min.Y &&
		a.Min.Z < other.Max.Z && a.Max.Z > other.Min.Z
}

// Contains returns true if this AABB fully contains other.
func (a AABB) Contains(other AABB) bool {
	return a.Min.X <= other.Min.X && a.Max.X >= other.Max.X &&
		a.Min.Y <= other.Min.Y && a.Max.Y >= other.Max.Y &&
		a.Min.Z <= other.Min.Z && a.Max.Z >= other.Max.Z
}

// ContainsPoint returns true if the point is inside this AABB.
func (a AABB) ContainsPoint(p Vec3) bool {
	return p.X >= a.Min.X && p.X <= a.Max.X &&
		p.Y >= a.Min.Y && p.Y <= a.Max.Y &&
		p.Z >= a.Min.Z && p.Z <= a.Max.Z
}

// Expand returns a new AABB expanded to include the given point.
func (a AABB) Expand(p Vec3) AABB {
	return AABB{
		Min: Vec3{
			X: minf(a.Min.X, p.X),
			Y: minf(a.Min.Y, p.Y),
			Z: minf(a.Min.Z, p.Z),
		},
		Max: Vec3{
			X: maxf(a.Max.X, p.X),
			Y: maxf(a.Max.Y, p.Y),
			Z: maxf(a.Max.Z, p.Z),
		},
	}
}

// Offset returns a new AABB translated by the given vector.
func (a AABB) Offset(v Vec3) AABB {
	return AABB{
		Min: a.Min.Add(v),
		Max: a.Max.Add(v),
	}
}

// Size returns the size of the AABB as a Vec3.
func (a AABB) Size() Vec3 {
	return a.Max.Sub(a.Min)
}

// Center returns the center point of the AABB.
func (a AABB) Center() Vec3 {
	return a.Min.Add(a.Max).Scale(0.5)
}

// Grow returns a new AABB expanded by amount in all directions.
func (a AABB) Grow(amount float32) AABB {
	return AABB{
		Min: Vec3{a.Min.X - amount, a.Min.Y - amount, a.Min.Z - amount},
		Max: Vec3{a.Max.X + amount, a.Max.Y + amount, a.Max.Z + amount},
	}
}

// String returns a string representation of the AABB.
func (a AABB) String() string {
	return fmt.Sprintf("AABB(%s, %s)", a.Min, a.Max)
}

// RayIntersects tests whether a ray with the given origin and direction
// intersects this AABB within maxDist. It returns true and the distance
// parameter t if the ray hits, or false otherwise. Uses the slab method.
func (a AABB) RayIntersects(origin, direction Vec3, maxDist float32) (bool, float32) {
	var tMin, tMax float32

	if direction.X != 0 {
		invD := 1.0 / direction.X
		t0 := (a.Min.X - origin.X) * invD
		t1 := (a.Max.X - origin.X) * invD
		if invD < 0 {
			t0, t1 = t1, t0
		}
		tMin = t0
		tMax = t1
	} else {
		if origin.X < a.Min.X || origin.X > a.Max.X {
			return false, 0
		}
		tMin = -1e30
		tMax = 1e30
	}

	if direction.Y != 0 {
		invD := 1.0 / direction.Y
		t0 := (a.Min.Y - origin.Y) * invD
		t1 := (a.Max.Y - origin.Y) * invD
		if invD < 0 {
			t0, t1 = t1, t0
		}
		if t0 > tMin {
			tMin = t0
		}
		if t1 < tMax {
			tMax = t1
		}
	} else {
		if origin.Y < a.Min.Y || origin.Y > a.Max.Y {
			return false, 0
		}
	}

	if tMin > tMax {
		return false, 0
	}

	if direction.Z != 0 {
		invD := 1.0 / direction.Z
		t0 := (a.Min.Z - origin.Z) * invD
		t1 := (a.Max.Z - origin.Z) * invD
		if invD < 0 {
			t0, t1 = t1, t0
		}
		if t0 > tMin {
			tMin = t0
		}
		if t1 < tMax {
			tMax = t1
		}
	} else {
		if origin.Z < a.Min.Z || origin.Z > a.Max.Z {
			return false, 0
		}
	}

	if tMin > tMax {
		return false, 0
	}

	// The intersection is valid if the nearest hit is within [0, maxDist].
	t := tMin
	if t < 0 {
		t = tMax
	}
	if t < 0 || t > maxDist {
		return false, 0
	}

	return true, t
}

// BlockAABB returns a unit AABB for the given block position.
func BlockAABB(pos BlockPos) AABB {
	return AABB{
		Min: pos.ToVec3(),
		Max: Vec3{float32(pos.X) + 1, float32(pos.Y) + 1, float32(pos.Z) + 1},
	}
}

// StairAABBs returns two AABBs representing a stair block at pos.
// The bottom half is always a full-width, half-height slab.
// The stepped quarter sits on top of the bottom half and is offset
// in the horizontal axis determined by orientation:
//
//	0 = North (+Z side), 1 = South (-Z side),
//	2 = East (-X side), 3 = West (+X side).
//
// Other orientation values default to North.
func StairAABBs(pos BlockPos, orientation int) []AABB {
	fx := float32(pos.X)
	fy := float32(pos.Y)
	fz := float32(pos.Z)

	// Bottom half: full 1x0.5x1 slab
	bottom := AABB{
		Min: Vec3{fx, fy, fz},
		Max: Vec3{fx + 1, fy + 0.5, fz + 1},
	}

	// Top step: half-width, half-height quarter
	var top AABB
	switch orientation {
	case 1: // South: step on -Z side
		top = AABB{
			Min: Vec3{fx, fy + 0.5, fz},
			Max: Vec3{fx + 1, fy + 1, fz + 0.5},
		}
	case 2: // East: step on -X side
		top = AABB{
			Min: Vec3{fx, fy + 0.5, fz},
			Max: Vec3{fx + 0.5, fy + 1, fz + 1},
		}
	case 3: // West: step on +X side
		top = AABB{
			Min: Vec3{fx + 0.5, fy + 0.5, fz},
			Max: Vec3{fx + 1, fy + 1, fz + 1},
		}
	default: // 0 = North: step on +Z side
		top = AABB{
			Min: Vec3{fx, fy + 0.5, fz + 0.5},
			Max: Vec3{fx + 1, fy + 1, fz + 1},
		}
	}

	return []AABB{bottom, top}
}

// SlabAABB returns a half-height AABB for a slab block at pos.
// If top is true the slab occupies the upper half; otherwise the lower half.
func SlabAABB(pos BlockPos, top bool) AABB {
	fx := float32(pos.X)
	fy := float32(pos.Y)
	fz := float32(pos.Z)

	if top {
		return AABB{
			Min: Vec3{fx, fy + 0.5, fz},
			Max: Vec3{fx + 1, fy + 1, fz + 1},
		}
	}
	return AABB{
		Min: Vec3{fx, fy, fz},
		Max: Vec3{fx + 1, fy + 0.5, fz + 1},
	}
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
