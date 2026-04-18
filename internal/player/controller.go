// Package player implements the first-person player controller including
// movement, camera control, block interaction, and hotbar management.
package player

import (
	"math"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

// ModeChecker abstracts game-mode capability queries so that the player
// package does not depend on the game package (avoiding a circular import).
type ModeChecker interface {
	CanFly() bool
	CanBreakInstantly() bool
	HasInfiniteItems() bool
	CanTakeDamage() bool
	IsNoClip() bool
}

const (
	// DefaultWalkSpeed is the player walking speed in blocks per second.
	DefaultWalkSpeed float32 = 4.317
	// DefaultSprintSpeed is the player sprinting speed in blocks per second.
	DefaultSprintSpeed float32 = 5.612
	// DefaultSneakSpeed is the player sneaking speed in blocks per second.
	DefaultSneakSpeed float32 = 1.31
	// DefaultFlySpeed is the player flying speed in blocks per second.
	DefaultFlySpeed float32 = 10.92
	// DefaultJumpVelocity is the initial upward velocity applied when jumping.
	DefaultJumpVelocity float32 = 8.0
	// DefaultReach is the maximum distance the player can interact with blocks.
	DefaultReach float32 = 5.0
	// DefaultSensitivity is the default mouse sensitivity multiplier.
	DefaultSensitivity float32 = 0.003
	// EyeOffset is the vertical offset from the entity position to the camera.
	EyeOffset float32 = 1.62
	// DoubleTapWindow is the maximum interval (in seconds) between two jump
	// presses to toggle flying.
	DoubleTapWindow float32 = 0.3
)

// Controller manages the player entity, camera, and all player-specific
// state such as movement speeds, block interaction, and hotbar selection.
type Controller struct {
	Entity        ecs.Entity
	ECSWorld      *ecs.World
	Camera        PlayerCamera
	KeyMap        *input.KeyMap
	Mode          ModeChecker
	WalkSpeed     float32
	SprintSpeed   float32
	SneakSpeed    float32
	FlySpeed      float32
	JumpVelocity  float32
	Reach         float32
	Sensitivity   float32
	BreakProgress float32
	BreakingBlock *mcmath.BlockPos
	Inventory     *inventory.Inventory
	SelectedSlot  int
	OnUseBlock    func(blockID uint16, pos mcmath.BlockPos)
	Flying        bool

	// interaction holds block interaction state.
	interaction interactionState

	// lastJumpTime tracks elapsed time since last jump press for double-tap
	// fly toggle detection. A negative value means no recent press.
	lastJumpTime float32
}

// NewController creates a Controller for the given player entity with default
// speeds, reach, and sensitivity.
func NewController(e ecs.Entity, ecsWorld *ecs.World, camera PlayerCamera, keyMap *input.KeyMap) *Controller {
	return &Controller{
		Entity:       e,
		ECSWorld:     ecsWorld,
		Camera:       camera,
		KeyMap:       keyMap,
		WalkSpeed:    DefaultWalkSpeed,
		SprintSpeed:  DefaultSprintSpeed,
		SneakSpeed:   DefaultSneakSpeed,
		FlySpeed:     DefaultFlySpeed,
		JumpVelocity: DefaultJumpVelocity,
		Reach:        DefaultReach,
		Sensitivity:  DefaultSensitivity,
		SelectedSlot: 0,
		lastJumpTime: -1,
	}
}

// Update processes input and advances the player state by dt seconds.
// It handles camera rotation, movement direction, speed selection,
// jumping, flying, and camera synchronisation from the entity transform.
func (c *Controller) Update(inp *input.Manager, w BlockWorld, dt float32) {
	c.updateCamera(inp)
	c.updateMovement(inp, dt)
	c.updateNoClip()
	c.syncCameraPosition()
	c.updateHotbar(inp)
	c.UpdateInteraction(inp, w, dt)
}

// updateNoClip synchronises the physics body NoClip flag with the current
// game mode (spectator mode enables noclip).
func (c *Controller) updateNoClip() {
	if c.Mode == nil {
		return
	}
	body := c.getPhysicsBody()
	if body == nil {
		return
	}
	body.Body.NoClip = c.Mode.IsNoClip()
}

// updateCamera reads mouse delta and rotates the camera.
func (c *Controller) updateCamera(inp *input.Manager) {
	dx, dy := inp.MouseDelta()
	deltaYaw := float32(dx) * c.Sensitivity
	deltaPitch := float32(-dy) * c.Sensitivity
	c.Camera.Rotate(deltaYaw, deltaPitch)
}

// updateMovement reads WASD input, applies the correct speed modifier,
// and sets the entity physics body velocity. In fly-capable modes a
// double-tap of the jump key toggles flying; while flying, jump ascends,
// sneak descends, and gravity is disabled.
func (c *Controller) updateMovement(inp *input.Manager, dt float32) {
	body := c.getPhysicsBody()
	if body == nil {
		return
	}

	c.handleFlyToggle(inp, dt)

	moveDir := c.calculateMoveDirection(inp)

	if c.Flying {
		c.updateFlyingMovement(inp, body, moveDir)
	} else {
		c.updateWalkingMovement(inp, body, moveDir)
	}
}

// handleFlyToggle detects a double-tap of the jump key and toggles
// flying when the current mode allows it.
func (c *Controller) handleFlyToggle(inp *input.Manager, dt float32) {
	if c.Mode == nil || !c.Mode.CanFly() {
		c.Flying = false
		c.lastJumpTime = -1
		return
	}

	jumpKey := c.KeyMap.GetKey(input.Jump)

	// Advance the timer; negative means "no recent press".
	if c.lastJumpTime >= 0 {
		c.lastJumpTime += dt
	}

	if inp.IsKeyJustPressed(jumpKey) {
		if c.lastJumpTime >= 0 && c.lastJumpTime <= DoubleTapWindow {
			c.Flying = !c.Flying
			c.lastJumpTime = -1

			// Disable/enable gravity on the physics body.
			body := c.getPhysicsBody()
			if body != nil {
				if c.Flying {
					body.Body.Gravity = 0
					body.Body.Velocity.Y = 0
				} else {
					body.Body.Gravity = physics.DefaultGravity
				}
			}
		} else {
			c.lastJumpTime = 0
		}
	}
}

// updateFlyingMovement applies horizontal and vertical movement while
// the player is flying. Jump goes up, sneak goes down, no gravity.
func (c *Controller) updateFlyingMovement(inp *input.Manager, body *entity.PhysicsBody, moveDir mcmath.Vec3) {
	speed := c.FlySpeed

	body.Body.Velocity.X = moveDir.X * speed
	body.Body.Velocity.Z = moveDir.Z * speed

	jumpKey := c.KeyMap.GetKey(input.Jump)
	sneakKey := c.KeyMap.GetKey(input.Sneak)

	var verticalVel float32
	if inp.IsKeyDown(jumpKey) {
		verticalVel = speed
	}
	if inp.IsKeyDown(sneakKey) {
		verticalVel = -speed
	}
	body.Body.Velocity.Y = verticalVel
}

// updateWalkingMovement applies normal ground-based movement.
func (c *Controller) updateWalkingMovement(inp *input.Manager, body *entity.PhysicsBody, moveDir mcmath.Vec3) {
	speed := c.currentSpeed(inp)
	body.Body.Velocity.X = moveDir.X * speed
	body.Body.Velocity.Z = moveDir.Z * speed

	// Jump: set Y velocity if on ground and jump key pressed.
	jumpKey := c.KeyMap.GetKey(input.Jump)
	if body.Body.OnGround && inp.IsKeyDown(jumpKey) {
		body.Body.Velocity.Y = c.JumpVelocity
	}
}

// calculateMoveDirection returns a normalised horizontal direction vector
// based on which movement keys are held, relative to the camera orientation.
func (c *Controller) calculateMoveDirection(inp *input.Manager) mcmath.Vec3 {
	fwd := c.Camera.Forward()
	right := c.Camera.Right()

	// Project forward onto horizontal plane.
	horizontal := mcmath.Vec3{X: fwd.X, Y: 0, Z: fwd.Z}.Normalize()

	var dir mcmath.Vec3

	fwdKey := c.KeyMap.GetKey(input.MoveForward)
	backKey := c.KeyMap.GetKey(input.MoveBackward)
	leftKey := c.KeyMap.GetKey(input.MoveLeft)
	rightKey := c.KeyMap.GetKey(input.MoveRight)

	if inp.IsKeyDown(fwdKey) {
		dir = dir.Add(horizontal)
	}
	if inp.IsKeyDown(backKey) {
		dir = dir.Sub(horizontal)
	}
	if inp.IsKeyDown(leftKey) {
		dir = dir.Sub(right)
	}
	if inp.IsKeyDown(rightKey) {
		dir = dir.Add(right)
	}

	return dir.Normalize()
}

// currentSpeed returns the appropriate movement speed based on modifier keys.
func (c *Controller) currentSpeed(inp *input.Manager) float32 {
	sneakKey := c.KeyMap.GetKey(input.Sneak)
	sprintKey := c.KeyMap.GetKey(input.Sprint)

	if inp.IsKeyDown(sneakKey) {
		return c.SneakSpeed
	}
	if inp.IsKeyDown(sprintKey) {
		return c.SprintSpeed
	}
	return c.WalkSpeed
}

// syncCameraPosition updates the camera position from the entity's
// transform position plus the eye-height offset.
func (c *Controller) syncCameraPosition() {
	transform := c.getTransform()
	if transform == nil {
		return
	}
	c.Camera.SetPosition(transform.Position.Add(mcmath.Vec3{X: 0, Y: EyeOffset, Z: 0}))
}

// getPhysicsBody returns a pointer to the entity's PhysicsBody component,
// or nil if missing.
func (c *Controller) getPhysicsBody() *entity.PhysicsBody {
	store := ecs.GetStore[entity.PhysicsBody](c.ECSWorld)
	pb, ok := store.Get(c.Entity)
	if !ok {
		return nil
	}
	return pb
}

// getTransform returns a pointer to the entity's Transform component,
// or nil if missing.
func (c *Controller) getTransform() *entity.Transform {
	store := ecs.GetStore[entity.Transform](c.ECSWorld)
	t, ok := store.Get(c.Entity)
	if !ok {
		return nil
	}
	return t
}

// PlayerAABB returns the player's world-space AABB computed from the
// entity's physics body.
func (c *Controller) PlayerAABB() mcmath.AABB {
	pb := c.getPhysicsBody()
	if pb == nil {
		return mcmath.AABB{}
	}
	return pb.Body.WorldAABB()
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
