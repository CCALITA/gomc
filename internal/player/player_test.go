//go:build !ci

package player

import (
	"math"
	"testing"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
	"github.com/fanxiyao/gomc/internal/render"
	"github.com/stretchr/testify/assert"
)

func setupController() (*Controller, *ecs.World) {
	w := ecs.NewWorld()
	e := entity.SpawnPlayer(w, "test", mcmath.Vec3{X: 0, Y: 64, Z: 0})
	cam := &render.Camera{Position: mcmath.Vec3{X: 0, Y: 65.62, Z: 0}, FOV: 70}
	km := input.NewKeyMap()
	ctrl := NewController(e, w, cam, km)
	return ctrl, w
}

func TestNewController(t *testing.T) {
	ctrl, _ := setupController()
	assert.Equal(t, DefaultWalkSpeed, ctrl.WalkSpeed)
	assert.Equal(t, DefaultSprintSpeed, ctrl.SprintSpeed)
	assert.Equal(t, DefaultSneakSpeed, ctrl.SneakSpeed)
	assert.Equal(t, DefaultJumpVelocity, ctrl.JumpVelocity)
	assert.Equal(t, DefaultReach, ctrl.Reach)
	assert.Equal(t, DefaultSensitivity, ctrl.Sensitivity)
	assert.Equal(t, 0, ctrl.SelectedSlot)
}

func TestCurrentSpeed_Walk(t *testing.T) {
	ctrl, _ := setupController()
	mgr := input.NewManager()
	speed := ctrl.currentSpeed(mgr)
	assert.Equal(t, DefaultWalkSpeed, speed)
}

func TestCurrentSpeed_Sprint(t *testing.T) {
	ctrl, _ := setupController()
	mgr := input.NewManager()
	sprintKey := ctrl.KeyMap.GetKey(input.Sprint)
	mgr.KeyCallback(sprintKey, 0, 1, 0) // press
	speed := ctrl.currentSpeed(mgr)
	assert.Equal(t, DefaultSprintSpeed, speed)
}

func TestCurrentSpeed_Sneak(t *testing.T) {
	ctrl, _ := setupController()
	mgr := input.NewManager()
	sneakKey := ctrl.KeyMap.GetKey(input.Sneak)
	mgr.KeyCallback(sneakKey, 0, 1, 0) // press
	speed := ctrl.currentSpeed(mgr)
	assert.Equal(t, DefaultSneakSpeed, speed)
}

func TestCurrentSpeed_SneakOverridesSprint(t *testing.T) {
	ctrl, _ := setupController()
	mgr := input.NewManager()
	sneakKey := ctrl.KeyMap.GetKey(input.Sneak)
	sprintKey := ctrl.KeyMap.GetKey(input.Sprint)
	mgr.KeyCallback(sneakKey, 0, 1, 0)
	mgr.KeyCallback(sprintKey, 0, 1, 0)
	speed := ctrl.currentSpeed(mgr)
	assert.Equal(t, DefaultSneakSpeed, speed)
}

func TestSelectSlot(t *testing.T) {
	ctrl, _ := setupController()

	ctrl.SelectSlot(5)
	assert.Equal(t, 5, ctrl.SelectedSlot)

	ctrl.SelectSlot(-1)
	assert.Equal(t, 0, ctrl.SelectedSlot)

	ctrl.SelectSlot(99)
	assert.Equal(t, 8, ctrl.SelectedSlot)
}

func TestScrollSlot(t *testing.T) {
	ctrl, _ := setupController()

	ctrl.SelectedSlot = 0
	ctrl.ScrollSlot(-1) // wrap to 8
	assert.Equal(t, 8, ctrl.SelectedSlot)

	ctrl.ScrollSlot(1) // wrap to 0
	assert.Equal(t, 0, ctrl.SelectedSlot)

	ctrl.SelectedSlot = 4
	ctrl.ScrollSlot(2)
	assert.Equal(t, 6, ctrl.SelectedSlot)
}

func TestSyncCameraPosition(t *testing.T) {
	ctrl, _ := setupController()
	transform := ctrl.getTransform()
	transform.Position = mcmath.Vec3{X: 10, Y: 70, Z: 20}

	ctrl.syncCameraPosition()
	camPos := ctrl.Camera.GetPosition()
	assert.InDelta(t, 10.0, camPos.X, 0.01)
	assert.InDelta(t, 71.62, camPos.Y, 0.01)
	assert.InDelta(t, 20.0, camPos.Z, 0.01)
}

func TestSyncCameraPosition_NoTransform(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	cam := &render.Camera{}
	km := input.NewKeyMap()
	ctrl := NewController(e, w, cam, km)

	// Should not panic with missing transform
	ctrl.syncCameraPosition()
}

func TestPlayerAABB(t *testing.T) {
	ctrl, _ := setupController()
	aabb := ctrl.PlayerAABB()
	size := aabb.Size()
	assert.Greater(t, size.X, float32(0))
	assert.Greater(t, size.Y, float32(0))
	assert.Greater(t, size.Z, float32(0))
}

func TestPlayerAABB_NoPhysicsBody(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	cam := &render.Camera{}
	km := input.NewKeyMap()
	ctrl := NewController(e, w, cam, km)

	aabb := ctrl.PlayerAABB()
	assert.Equal(t, mcmath.AABB{}, aabb)
}

func TestGetBreakProgress(t *testing.T) {
	ctrl, _ := setupController()
	assert.Equal(t, float32(0), ctrl.GetBreakProgress())
	ctrl.BreakProgress = 0.5
	assert.Equal(t, float32(0.5), ctrl.GetBreakProgress())
}

func TestResetBreaking(t *testing.T) {
	ctrl, _ := setupController()
	pos := mcmath.BlockPos{X: 1, Y: 2, Z: 3}
	ctrl.BreakingBlock = &pos
	ctrl.BreakProgress = 0.7

	ctrl.resetBreaking()
	assert.Nil(t, ctrl.BreakingBlock)
	assert.Equal(t, float32(0), ctrl.BreakProgress)
}

func TestCalculateMoveDirection_NoInput(t *testing.T) {
	ctrl, _ := setupController()
	mgr := input.NewManager()
	dir := ctrl.calculateMoveDirection(mgr)
	assert.InDelta(t, 0, dir.X, 0.01)
	assert.InDelta(t, 0, dir.Y, 0.01)
	assert.InDelta(t, 0, dir.Z, 0.01)
}

func TestCalculateMoveDirection_Forward(t *testing.T) {
	ctrl, _ := setupController()
	mgr := input.NewManager()
	fwdKey := ctrl.KeyMap.GetKey(input.MoveForward)
	mgr.KeyCallback(fwdKey, 0, 1, 0)

	dir := ctrl.calculateMoveDirection(mgr)
	length := dir.Length()
	if length > 0 {
		assert.InDelta(t, 1.0, length, 0.01)
	}
}

func TestJump_WhenGrounded(t *testing.T) {
	ctrl, _ := setupController()
	pb := ctrl.getPhysicsBody()
	pb.Body.OnGround = true
	pb.Body.Velocity.Y = 0

	mgr := input.NewManager()
	jumpKey := ctrl.KeyMap.GetKey(input.Jump)
	mgr.KeyCallback(jumpKey, 0, 1, 0)

	ctrl.updateMovement(mgr, 0.05)
	assert.Equal(t, DefaultJumpVelocity, pb.Body.Velocity.Y)
}

func TestJump_WhenAirborne(t *testing.T) {
	ctrl, _ := setupController()
	pb := ctrl.getPhysicsBody()
	pb.Body.OnGround = false
	pb.Body.Velocity.Y = -5.0

	mgr := input.NewManager()
	jumpKey := ctrl.KeyMap.GetKey(input.Jump)
	mgr.KeyCallback(jumpKey, 0, 1, 0)

	ctrl.updateMovement(mgr, 0.05)
	assert.Equal(t, float32(-5.0), pb.Body.Velocity.Y)
}

func TestUpdateMovement_NoPhysicsBody(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	cam := &render.Camera{}
	km := input.NewKeyMap()
	ctrl := NewController(e, w, cam, km)
	mgr := input.NewManager()

	// Should not panic
	ctrl.updateMovement(mgr, 0.05)
}

// horizontalForward returns the camera forward vector projected onto
// the horizontal plane with the given yaw, for use in calculating
// the move direction angle.
func horizontalForward(yaw float32) mcmath.Vec3 {
	return mcmath.Vec3{
		X: float32(math.Sin(float64(yaw))),
		Y: 0,
		Z: float32(-math.Cos(float64(yaw))),
	}
}

func TestHorizontalForward(t *testing.T) {
	v := horizontalForward(0)
	assert.InDelta(t, 0, v.X, 0.01)
	assert.Equal(t, float32(0), v.Y)
	assert.InDelta(t, -1.0, v.Z, 0.01)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 9, HotbarSlots)
	assert.Equal(t, float32(1.62), EyeOffset)
	assert.Greater(t, DefaultReach, float32(0))
}

func TestGetPhysicsBody_Missing(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	cam := &render.Camera{}
	km := input.NewKeyMap()
	ctrl := NewController(e, w, cam, km)
	assert.Nil(t, ctrl.getPhysicsBody())
}

func TestGetTransform_Present(t *testing.T) {
	ctrl, _ := setupController()
	transform := ctrl.getTransform()
	assert.NotNil(t, transform)
}

func TestGetTransform_Missing(t *testing.T) {
	w := ecs.NewWorld()
	e := w.NewEntity()
	cam := &render.Camera{}
	km := input.NewKeyMap()
	ctrl := NewController(e, w, cam, km)
	assert.Nil(t, ctrl.getTransform())
}

// Ensure physics.Body is properly set up by factory.
func TestPhysicsBodyFromFactory(t *testing.T) {
	ctrl, _ := setupController()
	pb := ctrl.getPhysicsBody()
	assert.NotNil(t, pb)
	assert.NotNil(t, pb.Body)
	assert.Equal(t, float32(physics.DefaultGravity), pb.Body.Gravity)
}
