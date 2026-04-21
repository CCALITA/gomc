package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

// blockMap is a test helper mapping (x,y,z) to a block ID.
type blockMap map[[3]int32]uint16

func (m blockMap) get(x, y, z int32) uint16 {
	return m[[3]int32{x, y, z}]
}

func setupMinecartWorld(blocks blockMap) (*ecs.World, *MinecartSystem) {
	w := ecs.NewWorld()
	sys := &MinecartSystem{
		GetBlock: blocks.get,
	}
	return w, sys
}

func TestMinecartFollowsRail(t *testing.T) {
	blocks := blockMap{
		{0, 0, 0}: block.Rail,
		{0, 0, -1}: block.Rail,
		{0, 0, -2}: block.Rail,
	}
	w, sys := setupMinecartWorld(blocks)
	e := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	mc, ok := ecs.GetStore[MinecartData](w).Get(e)
	assert.True(t, ok)
	mc.Speed = 4.0
	mc.Direction = mcmath.North // -Z

	tr, _ := ecs.GetStore[Transform](w).Get(e)
	startZ := tr.Position.Z

	sys.Update(w, 0.05) // dt=50ms

	tr, _ = ecs.GetStore[Transform](w).Get(e)
	assert.Less(t, tr.Position.Z, startZ, "minecart should move in -Z (North)")
	assert.True(t, mc.OnRail)
}

func TestMinecartSpeedOnPoweredRail(t *testing.T) {
	blocks := blockMap{
		{0, 0, 0}: block.PoweredRail,
		{0, 0, -1}: block.PoweredRail,
	}
	w, sys := setupMinecartWorld(blocks)
	e := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	mc, _ := ecs.GetStore[MinecartData](w).Get(e)
	mc.Speed = 2.0
	mc.Direction = mcmath.North

	sys.Update(w, 0.05)

	assert.Equal(t, minecartPoweredSpeed, mc.Speed,
		"speed should be set to powered rail speed")
}

func TestMinecartDecelerationOnUnpoweredRail(t *testing.T) {
	blocks := blockMap{
		{0, 0, 0}:  block.Rail,
		{0, 0, -1}: block.Rail,
		{0, 0, -2}: block.Rail,
		{0, 0, -3}: block.Rail,
		{0, 0, -4}: block.Rail,
	}
	w, sys := setupMinecartWorld(blocks)
	e := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	mc, _ := ecs.GetStore[MinecartData](w).Get(e)
	mc.Speed = 4.0
	mc.Direction = mcmath.North

	sys.Update(w, 0.05) // small dt so cart stays on rail

	expected := float32(4.0) - minecartDeceleration*0.05
	assert.InDelta(t, expected, mc.Speed, 0.01,
		"speed should decrease by deceleration rate")
}

func TestMinecartSlopeAcceleration(t *testing.T) {
	blocks := blockMap{
		{0, 1, 0}:  block.Rail, // minecart is here
		{0, 0, 0}:  block.Rail, // rail below = slope
		{0, 1, -1}: block.Rail,
		{0, 0, -1}: block.Rail,
	}
	w, sys := setupMinecartWorld(blocks)
	e := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 1.5, Z: 0.5})

	mc, _ := ecs.GetStore[MinecartData](w).Get(e)
	mc.Speed = 2.0
	mc.Direction = mcmath.North

	sys.Update(w, 0.05)

	expected := float32(2.0) + minecartSlopeAccel*0.05
	if expected > minecartMaxSpeed {
		expected = minecartMaxSpeed
	}
	assert.InDelta(t, expected, mc.Speed, 0.01,
		"speed should increase on slope")
}

func TestMinecartStopsOffRail(t *testing.T) {
	blocks := blockMap{
		{0, 0, 0}: block.Air,
	}
	w, sys := setupMinecartWorld(blocks)
	e := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	mc, _ := ecs.GetStore[MinecartData](w).Get(e)
	mc.Speed = 4.0

	sys.Update(w, 0.05)

	assert.False(t, mc.OnRail)
	assert.Equal(t, float32(0), mc.Speed)
}

func TestMinecartMountAndDismount(t *testing.T) {
	blocks := blockMap{
		{0, 0, 0}: block.Rail,
	}
	w, _ := setupMinecartWorld(blocks)
	minecart := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})
	rider := SpawnPlayer(w, "TestPlayer", mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	// Mount.
	MountMinecart(w, minecart, rider)
	mc, _ := ecs.GetStore[MinecartData](w).Get(minecart)
	assert.Equal(t, rider, mc.Rider)

	// Dismount.
	DismountMinecart(w, minecart)
	mc, _ = ecs.GetStore[MinecartData](w).Get(minecart)
	assert.Equal(t, ecs.Entity(0), mc.Rider)
}

func TestMinecartRiderMovesWithCart(t *testing.T) {
	blocks := blockMap{
		{0, 0, 0}: block.Rail,
		{0, 0, -1}: block.Rail,
		{0, 0, -2}: block.Rail,
	}
	w, sys := setupMinecartWorld(blocks)
	minecart := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})
	rider := SpawnPlayer(w, "TestRider", mcmath.Vec3{X: 0.5, Y: 1.2, Z: 0.5})

	MountMinecart(w, minecart, rider)
	mc, _ := ecs.GetStore[MinecartData](w).Get(minecart)
	mc.Speed = 4.0
	mc.Direction = mcmath.North

	sys.Update(w, 0.05)

	cartT, _ := ecs.GetStore[Transform](w).Get(minecart)
	riderT, _ := ecs.GetStore[Transform](w).Get(rider)

	assert.InDelta(t, cartT.Position.X, riderT.Position.X, 0.01)
	assert.InDelta(t, cartT.Position.Y+0.7, riderT.Position.Y, 0.01)
	assert.InDelta(t, cartT.Position.Z, riderT.Position.Z, 0.01)
}

func TestMinecartAABB(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnMinecart(w, mcmath.Vec3{X: 5, Y: 10, Z: 5})

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)

	width := pb.Body.BoundingBox.Max.X - pb.Body.BoundingBox.Min.X
	height := pb.Body.BoundingBox.Max.Y - pb.Body.BoundingBox.Min.Y
	depth := pb.Body.BoundingBox.Max.Z - pb.Body.BoundingBox.Min.Z

	assert.InDelta(t, 0.98, width, 0.01)
	assert.InDelta(t, 0.7, height, 0.01)
	assert.InDelta(t, 0.98, depth, 0.01)
}

func TestMinecartSpeedClampedAtZero(t *testing.T) {
	blocks := blockMap{
		{0, 0, 0}: block.Rail,
	}
	w, sys := setupMinecartWorld(blocks)
	e := SpawnMinecart(w, mcmath.Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	mc, _ := ecs.GetStore[MinecartData](w).Get(e)
	mc.Speed = 0.1
	mc.Direction = mcmath.North

	// Decelerate for long enough to go negative.
	sys.Update(w, 10.0)

	assert.GreaterOrEqual(t, mc.Speed, float32(0))
}

func TestPerpendicularDirections(t *testing.T) {
	ns := perpendicularDirections(mcmath.North)
	assert.Equal(t, mcmath.East, ns[0])
	assert.Equal(t, mcmath.West, ns[1])

	ew := perpendicularDirections(mcmath.East)
	assert.Equal(t, mcmath.North, ew[0])
	assert.Equal(t, mcmath.South, ew[1])
}
