package mcmath

import (
	"fmt"
	"math"
)

// Vec3 represents a 3D vector with float32 components.
type Vec3 struct {
	X, Y, Z float32
}

// Add returns the component-wise sum of v and other.
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

// Sub returns the component-wise difference of v and other.
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

// Mul returns the component-wise product of v and other.
func (v Vec3) Mul(other Vec3) Vec3 {
	return Vec3{v.X * other.X, v.Y * other.Y, v.Z * other.Z}
}

// Scale returns the vector scaled by s.
func (v Vec3) Scale(s float32) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

// Dot returns the dot product of v and other.
func (v Vec3) Dot(other Vec3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross returns the cross product of v and other.
func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		v.Y*other.Z - v.Z*other.Y,
		v.Z*other.X - v.X*other.Z,
		v.X*other.Y - v.Y*other.X,
	}
}

// LengthSq returns the squared length of the vector.
func (v Vec3) LengthSq() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// Length returns the length (magnitude) of the vector.
func (v Vec3) Length() float32 {
	return float32(math.Sqrt(float64(v.LengthSq())))
}

// Normalize returns a unit vector in the same direction as v.
// Returns the zero vector if v has zero length.
func (v Vec3) Normalize() Vec3 {
	l := v.Length()
	if l == 0 {
		return Vec3{}
	}
	return Vec3{v.X / l, v.Y / l, v.Z / l}
}

// Lerp returns the linear interpolation between v and other by t.
func (v Vec3) Lerp(other Vec3, t float32) Vec3 {
	return Vec3{
		v.X + (other.X-v.X)*t,
		v.Y + (other.Y-v.Y)*t,
		v.Z + (other.Z-v.Z)*t,
	}
}

// DistanceSq returns the squared distance between v and other.
func (v Vec3) DistanceSq(other Vec3) float32 {
	return v.Sub(other).LengthSq()
}

// Distance returns the distance between v and other.
func (v Vec3) Distance(other Vec3) float32 {
	return v.Sub(other).Length()
}

// Floor returns the BlockPos obtained by flooring each component.
func (v Vec3) Floor() BlockPos {
	return BlockPos{
		X: int32(math.Floor(float64(v.X))),
		Y: int32(math.Floor(float64(v.Y))),
		Z: int32(math.Floor(float64(v.Z))),
	}
}

// String returns a string representation of the vector.
func (v Vec3) String() string {
	return fmt.Sprintf("Vec3(%.4f, %.4f, %.4f)", v.X, v.Y, v.Z)
}
