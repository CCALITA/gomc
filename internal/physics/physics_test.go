package physics

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
)

// playerBox returns a typical player-sized bounding box (0.6 wide, 1.8 tall).
func playerBox() mcmath.AABB {
	return mcmath.AABB{
		Min: mcmath.Vec3{X: -0.3, Y: 0, Z: -0.3},
		Max: mcmath.Vec3{X: 0.3, Y: 1.8, Z: 0.3},
	}
}

// noBlocks is a getBlockAABBs function that always returns no blocks.
func noBlocks(_ mcmath.AABB) []mcmath.AABB { return nil }

// floorAt returns a getBlockAABBs function that reports a solid floor at the
// given block Y level.
func floorAt(y float32) func(mcmath.AABB) []mcmath.AABB {
	return func(region mcmath.AABB) []mcmath.AABB {
		var out []mcmath.AABB
		minX := int32(region.Min.X) - 1
		maxX := int32(region.Max.X) + 1
		minZ := int32(region.Min.Z) - 1
		maxZ := int32(region.Max.Z) + 1
		blockY := int32(y)
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				block := mcmath.BlockAABB(mcmath.BlockPos{X: x, Y: blockY, Z: z})
				if block.Intersects(region) {
					out = append(out, block)
				}
			}
		}
		return out
	}
}

// wallAtX returns block AABBs forming a wall at x=wallX for y in [0, height).
func wallAtX(wallX int32, height int32) func(mcmath.AABB) []mcmath.AABB {
	return func(region mcmath.AABB) []mcmath.AABB {
		var out []mcmath.AABB
		minZ := int32(region.Min.Z) - 1
		maxZ := int32(region.Max.Z) + 1
		for y := int32(0); y < height; y++ {
			for z := minZ; z <= maxZ; z++ {
				block := mcmath.BlockAABB(mcmath.BlockPos{X: wallX, Y: y, Z: z})
				if block.Intersects(region) {
					out = append(out, block)
				}
			}
		}
		return out
	}
}

// wallAtZ returns block AABBs forming a wall at z=wallZ for y in [0, height).
func wallAtZ(wallZ int32, height int32) func(mcmath.AABB) []mcmath.AABB {
	return func(region mcmath.AABB) []mcmath.AABB {
		var out []mcmath.AABB
		minX := int32(region.Min.X) - 1
		maxX := int32(region.Max.X) + 1
		for y := int32(0); y < height; y++ {
			for x := minX; x <= maxX; x++ {
				block := mcmath.BlockAABB(mcmath.BlockPos{X: x, Y: y, Z: wallZ})
				if block.Intersects(region) {
					out = append(out, block)
				}
			}
		}
		return out
	}
}

// floorAndWall combines floor blocks and a wall.
func floorAndWall(floorY float32, wallX int32, wallHeight int32) func(mcmath.AABB) []mcmath.AABB {
	floorFn := floorAt(floorY)
	wallFn := wallAtX(wallX, wallHeight)
	return func(region mcmath.AABB) []mcmath.AABB {
		blocks := floorFn(region)
		blocks = append(blocks, wallFn(region)...)
		return blocks
	}
}

// ---------------------------------------------------------------------------
// Body tests
// ---------------------------------------------------------------------------

func TestNewBody(t *testing.T) {
	bb := playerBox()
	b := NewBody(bb)

	assert.Equal(t, DefaultGravity, b.Gravity)
	assert.Equal(t, DefaultAirDrag, b.Drag)
	assert.Equal(t, DefaultMaxFallSpeed, b.MaxFallSpeed)
	assert.Equal(t, bb, b.BoundingBox)
	assert.False(t, b.OnGround)
	assert.False(t, b.NoClip)
}

func TestWorldAABB(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{X: 10, Y: 20, Z: 30}

	w := b.WorldAABB()
	assert.InDelta(t, 9.7, float64(w.Min.X), 0.001)
	assert.InDelta(t, 20.0, float64(w.Min.Y), 0.001)
	assert.InDelta(t, 29.7, float64(w.Min.Z), 0.001)
	assert.InDelta(t, 10.3, float64(w.Max.X), 0.001)
	assert.InDelta(t, 21.8, float64(w.Max.Y), 0.001)
	assert.InDelta(t, 30.3, float64(w.Max.Z), 0.001)
}

// ---------------------------------------------------------------------------
// Free-fall test
// ---------------------------------------------------------------------------

func TestFreeFall(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{Y: 100}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		ResolveCollision(&b, dt, noBlocks)
	}

	assert.True(t, b.Position.Y < 100, "body should fall")
	assert.False(t, b.OnGround)
}

func TestMaxFallSpeedClamp(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{Y: 10000}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 600; i++ {
		ResolveCollision(&b, dt, noBlocks)
	}

	assert.True(t, b.Velocity.Y >= -b.MaxFallSpeed, "velocity should be clamped")
}

// ---------------------------------------------------------------------------
// Ground collision test
// ---------------------------------------------------------------------------

func TestGroundCollision(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{Y: 0.5}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 120; i++ {
		ResolveCollision(&b, dt, floorAt(-1))
	}

	assert.InDelta(t, 0, float64(b.Position.Y), 0.1, "body should rest on floor")
	assert.True(t, b.OnGround, "body should be on ground")
	assert.InDelta(t, 0, float64(b.Velocity.Y), 0.01, "Y velocity should be zero")
}

// ---------------------------------------------------------------------------
// Wall sliding test
// ---------------------------------------------------------------------------

func TestWallSliding(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{X: 0, Y: 0, Z: 0}
	b.Velocity = mcmath.Vec3{X: 10, Y: 0, Z: 5}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		ResolveCollision(&b, dt, floorAndWall(-1, 2, 5))
	}

	assert.True(t, b.Position.X < 2.0, "body should be stopped by wall")
	assert.True(t, b.Position.Z > 0, "body should slide along Z axis")
}

// ---------------------------------------------------------------------------
// X-axis collision
// ---------------------------------------------------------------------------

func TestXAxisCollision(t *testing.T) {
	b := NewBody(playerBox())
	b.Gravity = 0
	b.Position = mcmath.Vec3{X: 0, Y: 0, Z: 0}
	b.Velocity = mcmath.Vec3{X: 5, Y: 0, Z: 0}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		ResolveCollision(&b, dt, wallAtX(2, 5))
	}

	assert.True(t, b.Position.X < float32(2), "body should be blocked by wall on X")
	assert.InDelta(t, 0, float64(b.Velocity.X), 0.01, "X velocity zeroed on collision")
}

// ---------------------------------------------------------------------------
// Z-axis collision
// ---------------------------------------------------------------------------

func TestZAxisCollision(t *testing.T) {
	b := NewBody(playerBox())
	b.Gravity = 0
	b.Position = mcmath.Vec3{X: 0, Y: 0, Z: 0}
	b.Velocity = mcmath.Vec3{X: 0, Y: 0, Z: 5}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 60; i++ {
		ResolveCollision(&b, dt, wallAtZ(2, 5))
	}

	assert.True(t, b.Position.Z < float32(2), "body should be blocked by wall on Z")
	assert.InDelta(t, 0, float64(b.Velocity.Z), 0.01, "Z velocity zeroed on collision")
}

// ---------------------------------------------------------------------------
// No-clip mode test
// ---------------------------------------------------------------------------

func TestNoClip(t *testing.T) {
	b := NewBody(playerBox())
	b.NoClip = true
	b.Position = mcmath.Vec3{Y: 10}
	b.Velocity = mcmath.Vec3{X: 5, Y: -5, Z: 3}

	dt := float32(1.0)
	ResolveCollision(&b, dt, floorAt(0))

	assert.InDelta(t, 5.0, float64(b.Position.X), 0.001)
	assert.InDelta(t, 5.0, float64(b.Position.Y), 0.001)
	assert.InDelta(t, 3.0, float64(b.Position.Z), 0.001)
}

// ---------------------------------------------------------------------------
// Drag application tests
// ---------------------------------------------------------------------------

func TestDragAir(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{Y: 100}
	b.Velocity = mcmath.Vec3{X: 10, Z: 10}

	dt := float32(1.0 / 60.0)
	ResolveCollision(&b, dt, noBlocks)

	expected := float32(10 * (1 - DefaultAirDrag))
	assert.InDelta(t, float64(expected), float64(b.Velocity.X), 0.01)
	assert.InDelta(t, float64(expected), float64(b.Velocity.Z), 0.01)
}

func TestDragGround(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{Y: 0}
	b.Velocity = mcmath.Vec3{X: 10, Z: 10}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 10; i++ {
		ResolveCollision(&b, dt, floorAt(-1))
	}

	b.Velocity.X = 10
	b.Velocity.Z = 10
	ResolveCollision(&b, dt, floorAt(-1))

	expected := float32(10 * (1 - DefaultGroundDrag))
	assert.InDelta(t, float64(expected), float64(b.Velocity.X), 0.5)
	assert.InDelta(t, float64(expected), float64(b.Velocity.Z), 0.5)
}

// ---------------------------------------------------------------------------
// Swept AABB tests
// ---------------------------------------------------------------------------

func TestSweepAABBHit(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 3, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 4, Y: 1, Z: 1},
	}
	velocity := mcmath.Vec3{X: 10, Y: 0, Z: 0}

	tHit, normal, hit := SweepAABB(moving, static, velocity)

	assert.True(t, hit, "should hit the static box")
	assert.InDelta(t, 0.2, float64(tHit), 0.01)
	assert.InDelta(t, -1, float64(normal.X), 0.01, "normal should point -X")
}

func TestSweepAABBMiss(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 5, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 6, Z: 1},
	}
	velocity := mcmath.Vec3{X: 10, Y: 0, Z: 0}

	_, _, hit := SweepAABB(moving, static, velocity)
	assert.False(t, hit, "should miss")
}

func TestSweepAABBNoVelocity(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 3, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 4, Y: 1, Z: 1},
	}

	_, _, hit := SweepAABB(moving, static, mcmath.Vec3{})
	assert.False(t, hit, "zero velocity should not hit")
}

func TestSweepAABBBehind(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 5, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 6, Y: 1, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	velocity := mcmath.Vec3{X: 10, Y: 0, Z: 0}

	_, _, hit := SweepAABB(moving, static, velocity)
	assert.False(t, hit, "static is behind the moving box")
}

func TestSweepAABBYAxis(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 5, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 6, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	velocity := mcmath.Vec3{X: 0, Y: -10, Z: 0}

	tHit, normal, hit := SweepAABB(moving, static, velocity)
	assert.True(t, hit)
	assert.InDelta(t, 0.4, float64(tHit), 0.01)
	assert.InDelta(t, 1, float64(normal.Y), 0.01, "normal should point +Y")
}

func TestSweepAABBZAxis(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 5},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 6},
	}
	velocity := mcmath.Vec3{X: 0, Y: 0, Z: 10}

	tHit, normal, hit := SweepAABB(moving, static, velocity)
	assert.True(t, hit)
	assert.InDelta(t, 0.4, float64(tHit), 0.01)
	assert.InDelta(t, -1, float64(normal.Z), 0.01, "normal should point -Z")
}

func TestSweepAABBTNearGreaterTFar(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 10, Y: 10, Z: 0},
		Max: mcmath.Vec3{X: 11, Y: 11, Z: 1},
	}
	velocity := mcmath.Vec3{X: 20, Y: 1, Z: 0}

	_, _, hit := SweepAABB(moving, static, velocity)
	assert.False(t, hit, "slabs should not overlap in time")
}

func TestSweepAABBTooFar(t *testing.T) {
	moving := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1},
	}
	static := mcmath.AABB{
		Min: mcmath.Vec3{X: 100, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 101, Y: 1, Z: 1},
	}
	velocity := mcmath.Vec3{X: 10, Y: 0, Z: 0}

	_, _, hit := SweepAABB(moving, static, velocity)
	assert.False(t, hit, "static is too far to reach within t=1")
}

// ---------------------------------------------------------------------------
// Raycast tests
// ---------------------------------------------------------------------------

func TestRaycastHit(t *testing.T) {
	isSolid := func(pos mcmath.BlockPos) bool {
		return pos.X == 5 && pos.Y == 0 && pos.Z == 0
	}

	hit, pos, face, tVal := RaycastBlocks(
		mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5},
		mcmath.Vec3{X: 1, Y: 0, Z: 0},
		10,
		isSolid,
	)

	assert.True(t, hit, "should hit the block")
	assert.Equal(t, mcmath.BlockPos{X: 5, Y: 0, Z: 0}, pos)
	assert.Equal(t, mcmath.West, face, "ray enters from -X face")
	assert.True(t, tVal > 0)
}

func TestRaycastMiss(t *testing.T) {
	isSolid := func(_ mcmath.BlockPos) bool { return false }

	hit, _, _, _ := RaycastBlocks(
		mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5},
		mcmath.Vec3{X: 1, Y: 0, Z: 0},
		10,
		isSolid,
	)

	assert.False(t, hit, "nothing solid")
}

func TestRaycastDown(t *testing.T) {
	isSolid := func(pos mcmath.BlockPos) bool {
		return pos.X == 0 && pos.Y == -3 && pos.Z == 0
	}

	hit, pos, face, _ := RaycastBlocks(
		mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5},
		mcmath.Vec3{X: 0, Y: -1, Z: 0},
		10,
		isSolid,
	)

	assert.True(t, hit)
	assert.Equal(t, mcmath.BlockPos{X: 0, Y: -3, Z: 0}, pos)
	assert.Equal(t, mcmath.Up, face, "ray enters from +Y face")
}

func TestRaycastStartInsideSolid(t *testing.T) {
	isSolid := func(pos mcmath.BlockPos) bool {
		return pos.X == 0 && pos.Y == 0 && pos.Z == 0
	}

	hit, pos, _, tVal := RaycastBlocks(
		mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5},
		mcmath.Vec3{X: 1, Y: 0, Z: 0},
		10,
		isSolid,
	)

	assert.True(t, hit)
	assert.Equal(t, mcmath.BlockPos{X: 0, Y: 0, Z: 0}, pos)
	assert.InDelta(t, 0, float64(tVal), 0.001, "t should be 0 for starting block")
}

// ---------------------------------------------------------------------------
// Penetration helper tests
// ---------------------------------------------------------------------------

func TestPenetration(t *testing.T) {
	body := mcmath.AABB{
		Min: mcmath.Vec3{X: 0.8, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 1.2, Y: 1, Z: 1},
	}
	block := mcmath.AABB{
		Min: mcmath.Vec3{X: 1, Y: 0, Z: 0},
		Max: mcmath.Vec3{X: 2, Y: 1, Z: 1},
	}

	pen := penetration(body, block, 0)
	assert.InDelta(t, -0.2, float64(pen), 0.001)
}

func TestPenetrationZAxis(t *testing.T) {
	body := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 0.8},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 1.2},
	}
	block := mcmath.AABB{
		Min: mcmath.Vec3{X: 0, Y: 0, Z: 1},
		Max: mcmath.Vec3{X: 1, Y: 1, Z: 2},
	}

	pen := penetration(body, block, 2)
	assert.InDelta(t, -0.2, float64(pen), 0.001)
}

// ---------------------------------------------------------------------------
// Combined scenario: body lands and slides
// ---------------------------------------------------------------------------

func TestLandAndSlide(t *testing.T) {
	b := NewBody(playerBox())
	b.Position = mcmath.Vec3{X: 0, Y: 5, Z: 0}
	b.Velocity = mcmath.Vec3{X: 8, Y: 0, Z: 0}

	dt := float32(1.0 / 60.0)
	for i := 0; i < 300; i++ {
		ResolveCollision(&b, dt, floorAt(-1))
	}

	assert.InDelta(t, 0, float64(b.Position.Y), 0.15, "should land on floor")
	assert.True(t, b.Position.X > 0, "should have moved in X")
	assert.True(t, b.OnGround)
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestConstants(t *testing.T) {
	assert.Equal(t, float32(28.0), DefaultGravity)
	assert.Equal(t, float32(0.02), DefaultAirDrag)
	assert.Equal(t, float32(0.6), DefaultGroundDrag)
	assert.Equal(t, float32(78.4), DefaultMaxFallSpeed)
}
