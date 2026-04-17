//go:build !ci

package render

import (
	"math"
	"testing"
	"unsafe"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// ---------------------------------------------------------------------------
// Camera tests
// ---------------------------------------------------------------------------

func TestNewCamera(t *testing.T) {
	pos := mcmath.Vec3{X: 10, Y: 20, Z: 30}
	cam := NewCamera(pos)

	if cam.Position != pos {
		t.Errorf("expected position %v, got %v", pos, cam.Position)
	}
	if cam.FOV != defaultFOV {
		t.Errorf("expected FOV %f, got %f", defaultFOV, cam.FOV)
	}
	if cam.Near != defaultNear {
		t.Errorf("expected Near %f, got %f", defaultNear, cam.Near)
	}
	if cam.Far != defaultFar {
		t.Errorf("expected Far %f, got %f", defaultFar, cam.Far)
	}
}

func TestCamera_Forward_DefaultLooksNegZ(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	fwd := cam.Forward()

	// Yaw=0, Pitch=0 => forward is (0, 0, -1)
	assertVec3Near(t, "forward", fwd, mcmath.Vec3{X: 0, Y: 0, Z: -1}, 1e-5)
}

func TestCamera_Forward_Yaw90(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	cam.Yaw = math.Pi / 2 // 90 degrees => looking along +X

	fwd := cam.Forward()
	assertVec3Near(t, "forward yaw90", fwd, mcmath.Vec3{X: 1, Y: 0, Z: 0}, 1e-5)
}

func TestCamera_Forward_PitchUp(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	cam.Pitch = math.Pi / 4 // 45 degrees up

	fwd := cam.Forward()

	expectedY := float32(math.Sin(math.Pi / 4))
	if math.Abs(float64(fwd.Y-expectedY)) > 1e-5 {
		t.Errorf("expected Y component ~%f, got %f", expectedY, fwd.Y)
	}
	// Length should be ~1
	length := fwd.Length()
	if math.Abs(float64(length-1.0)) > 1e-5 {
		t.Errorf("forward vector length should be ~1, got %f", length)
	}
}

func TestCamera_Right(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	// Default yaw=0 => right should be (+1, 0, 0)
	right := cam.Right()
	assertVec3Near(t, "right", right, mcmath.Vec3{X: 1, Y: 0, Z: 0}, 1e-5)
}

func TestCamera_Right_Yaw90(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	cam.Yaw = math.Pi / 2

	right := cam.Right()
	// When looking along +X, right should be (0, 0, +1)
	assertVec3Near(t, "right yaw90", right, mcmath.Vec3{X: 0, Y: 0, Z: 1}, 1e-5)
}

func TestCamera_Up(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	up := cam.Up()

	// Default orientation: up should be (0, 1, 0)
	assertVec3Near(t, "up", up, mcmath.Vec3{X: 0, Y: 1, Z: 0}, 1e-5)
}

func TestCamera_Rotate_ClampPitch(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})

	// Rotate pitch way past 90 degrees
	cam.Rotate(0, math.Pi)
	maxPitch := float32(math.Pi/2 - 0.01)
	if cam.Pitch > maxPitch {
		t.Errorf("pitch %f should be clamped to %f", cam.Pitch, maxPitch)
	}

	// Rotate pitch way past -90 degrees
	cam.Rotate(0, -2*math.Pi)
	if cam.Pitch < -maxPitch {
		t.Errorf("pitch %f should be clamped to %f", cam.Pitch, -maxPitch)
	}
}

func TestCamera_Rotate_AccumulatesYaw(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	cam.Rotate(0.5, 0)
	cam.Rotate(0.3, 0)

	expected := float32(0.8)
	if math.Abs(float64(cam.Yaw-expected)) > 1e-5 {
		t.Errorf("expected yaw %f, got %f", expected, cam.Yaw)
	}
}

func TestCamera_MoveForward(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{X: 0, Y: 10, Z: 0})
	// Default: looking along -Z
	cam.MoveForward(5.0)

	if math.Abs(float64(cam.Position.Z-(-5.0))) > 1e-4 {
		t.Errorf("expected Z ~-5.0 after moving forward, got %f", cam.Position.Z)
	}
	// Y should not change (horizontal movement)
	if cam.Position.Y != 10 {
		t.Errorf("expected Y to stay 10, got %f", cam.Position.Y)
	}
}

func TestCamera_MoveRight(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	cam.MoveRight(3.0)

	if math.Abs(float64(cam.Position.X-3.0)) > 1e-4 {
		t.Errorf("expected X ~3.0 after moving right, got %f", cam.Position.X)
	}
}

func TestCamera_MoveUp(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{Y: 5})
	cam.MoveUp(2.0)

	if cam.Position.Y != 7.0 {
		t.Errorf("expected Y=7.0, got %f", cam.Position.Y)
	}
}

func TestCamera_ViewMatrix_NotZero(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	view := cam.ViewMatrix()

	// The view matrix should not be all zeros.
	allZero := true
	for i := 0; i < 16; i++ {
		if view[i] != 0 {
			allZero = false
		}
	}
	if allZero {
		t.Error("view matrix is all zeros")
	}
}

func TestCamera_ProjectionMatrix_NotZero(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	proj := cam.ProjectionMatrix(16.0 / 9.0)

	allZero := true
	for i := 0; i < 16; i++ {
		if proj[i] != 0 {
			allZero = false
		}
	}
	if allZero {
		t.Error("projection matrix is all zeros")
	}
}

func TestCamera_ProjectionMatrix_VulkanYFlip(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})
	proj := cam.ProjectionMatrix(1.0)

	// In Vulkan, Y is flipped. proj[5] (col=1, row=1) should be negative.
	if proj[5] >= 0 {
		t.Errorf("expected proj[5] < 0 for Vulkan Y flip, got %f", proj[5])
	}
}

func TestCamera_ViewProjectionMatrix(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	vp := cam.ViewProjectionMatrix(16.0 / 9.0)

	// The VP matrix should not be all zeros.
	allZero := true
	for i := 0; i < 16; i++ {
		if vp[i] != 0 {
			allZero = false
		}
	}
	if allZero {
		t.Error("VP matrix is all zeros")
	}
}

func TestCamera_UpdateFrustum(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{X: 0, Y: 64, Z: 0})
	cam.UpdateFrustum(16.0 / 9.0)

	frustum := cam.Frustum()

	// A point directly in front should be inside.
	fwd := cam.Forward()
	testPoint := cam.Position.Add(fwd.Scale(10))
	if !frustum.ContainsPoint(testPoint) {
		t.Error("point directly in front of camera should be inside frustum")
	}

	// A point far behind should be outside.
	behind := cam.Position.Add(fwd.Scale(-100))
	if frustum.ContainsPoint(behind) {
		t.Error("point far behind camera should be outside frustum")
	}
}

// ---------------------------------------------------------------------------
// TextureAtlas tests (pure logic, no GPU)
// ---------------------------------------------------------------------------

func TestTextureAtlas_TileUV(t *testing.T) {
	atlas := &TextureAtlas{
		TileWidth:  16,
		TileHeight: 16,
		Columns:    16,
		Rows:       16,
	}

	tests := []struct {
		col, row       int
		wantU0, wantV0 float32
		wantU1, wantV1 float32
	}{
		{0, 0, 0.0, 0.0, 1.0 / 16, 1.0 / 16},
		{1, 0, 1.0 / 16, 0.0, 2.0 / 16, 1.0 / 16},
		{15, 15, 15.0 / 16, 15.0 / 16, 1.0, 1.0},
	}

	for _, tc := range tests {
		u0, v0, u1, v1 := atlas.TileUV(tc.col, tc.row)
		if !nearEqual(u0, tc.wantU0, 1e-5) || !nearEqual(v0, tc.wantV0, 1e-5) ||
			!nearEqual(u1, tc.wantU1, 1e-5) || !nearEqual(v1, tc.wantV1, 1e-5) {
			t.Errorf("TileUV(%d,%d) = (%f,%f,%f,%f), want (%f,%f,%f,%f)",
				tc.col, tc.row, u0, v0, u1, v1,
				tc.wantU0, tc.wantV0, tc.wantU1, tc.wantV1)
		}
	}
}

// ---------------------------------------------------------------------------
// QueueFamilyIndices tests
// ---------------------------------------------------------------------------

func TestQueueFamilyIndices_IsComplete(t *testing.T) {
	tests := []struct {
		name   string
		q      QueueFamilyIndices
		expect bool
	}{
		{"both set", QueueFamilyIndices{HasGraphics: true, HasPresent: true}, true},
		{"graphics only", QueueFamilyIndices{HasGraphics: true}, false},
		{"present only", QueueFamilyIndices{HasPresent: true}, false},
		{"neither", QueueFamilyIndices{}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.q.IsComplete(); got != tc.expect {
				t.Errorf("IsComplete() = %v, want %v", got, tc.expect)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Utility function tests
// ---------------------------------------------------------------------------

func TestClampUint32(t *testing.T) {
	tests := []struct {
		val, lo, hi uint32
		want        uint32
	}{
		{5, 1, 10, 5},
		{0, 1, 10, 1},
		{15, 1, 10, 10},
		{1, 1, 1, 1},
	}

	for _, tc := range tests {
		got := clampUint32(tc.val, tc.lo, tc.hi)
		if got != tc.want {
			t.Errorf("clampUint32(%d, %d, %d) = %d, want %d", tc.val, tc.lo, tc.hi, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// sliceUint32 tests
// ---------------------------------------------------------------------------

func TestSliceUint32_Empty(t *testing.T) {
	result := sliceUint32(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}

	result = sliceUint32([]byte{})
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestSliceUint32_ValidData(t *testing.T) {
	// 8 bytes = 2 uint32 values
	data := []byte{1, 0, 0, 0, 2, 0, 0, 0}
	result := sliceUint32(data)

	if len(result) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result))
	}
	if result[0] != 1 {
		t.Errorf("expected result[0]=1, got %d", result[0])
	}
	if result[1] != 2 {
		t.Errorf("expected result[1]=2, got %d", result[1])
	}
}

// ---------------------------------------------------------------------------
// ChunkPushConstants size test
// ---------------------------------------------------------------------------

func TestChunkPushConstants_Size(t *testing.T) {
	var pc ChunkPushConstants
	size := int(unsafe.Sizeof(pc))
	if size != 16 {
		t.Errorf("ChunkPushConstants size should be 16 bytes, got %d", size)
	}
}

// ---------------------------------------------------------------------------
// Renderer Aspect ratio test
// ---------------------------------------------------------------------------

func TestRenderer_Aspect(t *testing.T) {
	r := &Renderer{width: 1920, height: 1080}
	aspect := r.Aspect()
	expected := float32(1920.0) / float32(1080.0)
	if math.Abs(float64(aspect-expected)) > 1e-5 {
		t.Errorf("expected aspect %f, got %f", expected, aspect)
	}
}

func TestRenderer_Aspect_ZeroHeight(t *testing.T) {
	r := &Renderer{width: 800, height: 0}
	aspect := r.Aspect()
	if aspect != 1.0 {
		t.Errorf("expected aspect 1.0 for zero height, got %f", aspect)
	}
}

// ---------------------------------------------------------------------------
// Camera forward/right orthogonality test
// ---------------------------------------------------------------------------

func TestCamera_ForwardRightOrthogonal(t *testing.T) {
	cam := NewCamera(mcmath.Vec3{})

	testCases := []struct {
		name  string
		yaw   float32
		pitch float32
	}{
		{"default", 0, 0},
		{"yaw45", math.Pi / 4, 0},
		{"yaw90", math.Pi / 2, 0},
		{"pitch30", 0, math.Pi / 6},
		{"yaw45_pitch30", math.Pi / 4, math.Pi / 6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cam.Yaw = tc.yaw
			cam.Pitch = tc.pitch

			fwd := cam.Forward()
			right := cam.Right()
			dot := fwd.Dot(right)

			if math.Abs(float64(dot)) > 1e-4 {
				t.Errorf("forward and right should be orthogonal, dot=%f", dot)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertVec3Near(t *testing.T, name string, got, want mcmath.Vec3, eps float64) {
	t.Helper()
	if math.Abs(float64(got.X-want.X)) > eps ||
		math.Abs(float64(got.Y-want.Y)) > eps ||
		math.Abs(float64(got.Z-want.Z)) > eps {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}

func nearEqual(a, b float32, eps float64) bool {
	return math.Abs(float64(a-b)) < eps
}
