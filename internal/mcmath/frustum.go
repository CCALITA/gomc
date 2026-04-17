package mcmath

import "math"

// Plane represents a 3D plane in the form Normal.Dot(P) + D = 0.
type Plane struct {
	Normal Vec3
	D      float32
}

// normalize adjusts the plane so that its normal has unit length.
func (p Plane) normalize() Plane {
	l := p.Normal.Length()
	if l == 0 {
		return p
	}
	return Plane{
		Normal: Vec3{p.Normal.X / l, p.Normal.Y / l, p.Normal.Z / l},
		D:      p.D / l,
	}
}

// distanceToPoint returns the signed distance from the plane to a point.
func (p Plane) distanceToPoint(pt Vec3) float32 {
	return p.Normal.Dot(pt) + p.D
}

// Frustum represents a view frustum defined by six planes.
// The planes are ordered: Left, Right, Bottom, Top, Near, Far.
type Frustum [6]Plane

// ExtractFromMatrix extracts the six frustum planes from a combined
// view-projection matrix stored in column-major order as [16]float32.
func ExtractFromMatrix(vp [16]float32) Frustum {
	var f Frustum

	// Left: row3 + row0
	f[0] = Plane{
		Normal: Vec3{vp[3] + vp[0], vp[7] + vp[4], vp[11] + vp[8]},
		D:      vp[15] + vp[12],
	}
	// Right: row3 - row0
	f[1] = Plane{
		Normal: Vec3{vp[3] - vp[0], vp[7] - vp[4], vp[11] - vp[8]},
		D:      vp[15] - vp[12],
	}
	// Bottom: row3 + row1
	f[2] = Plane{
		Normal: Vec3{vp[3] + vp[1], vp[7] + vp[5], vp[11] + vp[9]},
		D:      vp[15] + vp[13],
	}
	// Top: row3 - row1
	f[3] = Plane{
		Normal: Vec3{vp[3] - vp[1], vp[7] - vp[5], vp[11] - vp[9]},
		D:      vp[15] - vp[13],
	}
	// Near: row3 + row2
	f[4] = Plane{
		Normal: Vec3{vp[3] + vp[2], vp[7] + vp[6], vp[11] + vp[10]},
		D:      vp[15] + vp[14],
	}
	// Far: row3 - row2
	f[5] = Plane{
		Normal: Vec3{vp[3] - vp[2], vp[7] - vp[6], vp[11] - vp[10]},
		D:      vp[15] - vp[14],
	}

	// Normalize all planes
	for i := range f {
		f[i] = f[i].normalize()
	}

	return f
}

// ContainsPoint returns true if the point is inside or on all frustum planes.
func (f Frustum) ContainsPoint(p Vec3) bool {
	for _, plane := range f {
		if plane.distanceToPoint(p) < 0 {
			return false
		}
	}
	return true
}

// IntersectsAABB returns true if the AABB is at least partially inside the frustum.
// Uses the "positive vertex" test for each plane.
func (f Frustum) IntersectsAABB(box AABB) bool {
	for _, plane := range f {
		// Find the positive vertex (the corner of the AABB most in the direction of the plane normal)
		px := box.Min.X
		if plane.Normal.X >= 0 {
			px = box.Max.X
		}
		py := box.Min.Y
		if plane.Normal.Y >= 0 {
			py = box.Max.Y
		}
		pz := box.Min.Z
		if plane.Normal.Z >= 0 {
			pz = box.Max.Z
		}

		if plane.Normal.X*px+plane.Normal.Y*py+plane.Normal.Z*pz+plane.D < 0 {
			return false
		}
	}
	return true
}

// identityMatrix returns a 4x4 identity matrix in column-major order.
func identityMatrix() [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// orthographicMatrix builds an orthographic projection matrix (column-major).
func orthographicMatrix(left, right, bottom, top, near, far float32) [16]float32 {
	rl := right - left
	tb := top - bottom
	fn := far - near
	return [16]float32{
		2 / rl, 0, 0, 0,
		0, 2 / tb, 0, 0,
		0, 0, -2 / fn, 0,
		-(right + left) / rl, -(top + bottom) / tb, -(far + near) / fn, 1,
	}
}

// perspectiveMatrix builds a perspective projection matrix (column-major).
func perspectiveMatrix(fovY, aspect, near, far float32) [16]float32 {
	f := float32(1.0 / math.Tan(float64(fovY)/2.0))
	nf := near - far
	return [16]float32{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) / nf, -1,
		0, 0, 2 * far * near / nf, 0,
	}
}
