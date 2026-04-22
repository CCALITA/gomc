package render

import (
	"math"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/go-gl/mathgl/mgl32"
)

// defaultFOV is the default field of view in radians (~70 degrees).
const defaultFOV = 70.0 * math.Pi / 180.0

// defaultNear is the default near clip plane distance.
const defaultNear = 0.1

// defaultFar is the default far clip plane distance.
const defaultFar = 1000.0

// Camera represents a first-person camera with position, orientation,
// and projection parameters.
type Camera struct {
	Position mcmath.Vec3
	Yaw      float32 // radians, 0 = looking along -Z
	Pitch    float32 // radians, clamped to [-pi/2, pi/2]
	FOV      float32 // vertical field of view in radians
	Near     float32
	Far      float32

	frustum mcmath.Frustum
}

// NewCamera creates a camera with default settings at the given position.
func NewCamera(pos mcmath.Vec3) *Camera {
	return &Camera{
		Position: pos,
		Yaw:      0,
		Pitch:    0,
		FOV:      defaultFOV,
		Near:     defaultNear,
		Far:      defaultFar,
	}
}

// Forward returns the unit direction vector the camera is looking at.
func (c *Camera) Forward() mcmath.Vec3 {
	cosP := float32(math.Cos(float64(c.Pitch)))
	return mcmath.Vec3{
		X: cosP * float32(math.Sin(float64(c.Yaw))),
		Y: float32(math.Sin(float64(c.Pitch))),
		Z: -cosP * float32(math.Cos(float64(c.Yaw))),
	}
}

// Right returns the unit vector pointing to the camera's right.
func (c *Camera) Right() mcmath.Vec3 {
	// Right is always horizontal (no pitch component).
	return mcmath.Vec3{
		X: float32(math.Cos(float64(c.Yaw))),
		Y: 0,
		Z: float32(math.Sin(float64(c.Yaw))),
	}
}

// Up returns the unit vector pointing upward relative to the camera.
func (c *Camera) Up() mcmath.Vec3 {
	fwd := c.Forward()
	right := c.Right()
	return right.Cross(fwd).Normalize()
}

// ViewMatrix returns the view matrix as a column-major mgl32.Mat4.
func (c *Camera) ViewMatrix() mgl32.Mat4 {
	fwd := c.Forward()
	eye := mgl32.Vec3{c.Position.X, c.Position.Y, c.Position.Z}
	center := mgl32.Vec3{
		c.Position.X + fwd.X,
		c.Position.Y + fwd.Y,
		c.Position.Z + fwd.Z,
	}
	up := mgl32.Vec3{0, 1, 0}
	return mgl32.LookAtV(eye, center, up)
}

// ProjectionMatrix returns the perspective projection matrix.
// Aspect is width/height.
func (c *Camera) ProjectionMatrix(aspect float32) mgl32.Mat4 {
	proj := mgl32.Perspective(c.FOV, aspect, c.Near, c.Far)
	proj[5] *= -1 // Vulkan Y-flip
	// Remap depth from OpenGL [-1,1] to Vulkan [0,1].
	proj[10] = proj[10]*0.5 + proj[11]*0.5
	proj[14] = proj[14]*0.5 + proj[15]*0.5
	return proj
}

// ViewProjectionMatrix returns the combined view-projection matrix.
func (c *Camera) ViewProjectionMatrix(aspect float32) mgl32.Mat4 {
	return c.ProjectionMatrix(aspect).Mul4(c.ViewMatrix())
}

// UpdateFrustum recomputes the camera's view frustum from the current
// view-projection matrix for use in culling.
func (c *Camera) UpdateFrustum(aspect float32) {
	vp := c.ViewProjectionMatrix(aspect)
	c.frustum = mcmath.ExtractFromMatrix(vp)
}

// Frustum returns the most recently computed frustum.
func (c *Camera) Frustum() mcmath.Frustum {
	return c.frustum
}

// Rotate adjusts the camera yaw and pitch by the given deltas (in radians).
// Pitch is clamped to avoid flipping.
func (c *Camera) Rotate(deltaYaw, deltaPitch float32) {
	c.Yaw += deltaYaw
	c.Pitch += deltaPitch

	const maxPitch = math.Pi/2 - 0.01
	if c.Pitch > maxPitch {
		c.Pitch = maxPitch
	}
	if c.Pitch < -maxPitch {
		c.Pitch = -maxPitch
	}
}

// GetPosition returns the current camera position.
func (c *Camera) GetPosition() mcmath.Vec3 {
	return c.Position
}

// SetPosition sets the camera position to the given value.
func (c *Camera) SetPosition(pos mcmath.Vec3) {
	c.Position = pos
}

// MoveForward moves the camera forward (positive) or backward (negative)
// by the given distance along the horizontal forward direction.
func (c *Camera) MoveForward(distance float32) {
	fwd := c.Forward()
	// Project to horizontal plane for standard FPS movement.
	horizontal := mcmath.Vec3{X: fwd.X, Y: 0, Z: fwd.Z}.Normalize()
	c.Position = c.Position.Add(horizontal.Scale(distance))
}

// MoveRight moves the camera right (positive) or left (negative) by the
// given distance.
func (c *Camera) MoveRight(distance float32) {
	right := c.Right()
	c.Position = c.Position.Add(right.Scale(distance))
}

// MoveUp moves the camera up (positive) or down (negative) by the given
// distance along the world Y axis.
func (c *Camera) MoveUp(distance float32) {
	c.Position.Y += distance
}
