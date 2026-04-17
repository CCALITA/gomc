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

// BlockAABB returns a unit AABB for the given block position.
func BlockAABB(pos BlockPos) AABB {
	return AABB{
		Min: pos.ToVec3(),
		Max: Vec3{float32(pos.X) + 1, float32(pos.Y) + 1, float32(pos.Z) + 1},
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
