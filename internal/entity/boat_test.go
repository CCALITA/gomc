package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

func TestSpawnBoat(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 10, Y: 64, Z: 20}
	e := SpawnBoat(w, pos)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	require.True(t, ok)
	assert.Equal(t, TypeBoat, et.Type)

	h, ok := ecs.GetStore[Health](w).Get(e)
	require.True(t, ok)
	assert.Equal(t, boatHealth, h.Current)
	assert.Equal(t, boatHealth, h.Max)

	bd, ok := ecs.GetStore[BoatData](w).Get(e)
	require.True(t, ok)
	assert.Equal(t, boatWaterSpeed, bd.WaterSpeed)
	assert.Equal(t, boatLandSpeed, bd.LandSpeed)
	assert.False(t, bd.HasRider)
}

func TestBoatFloatsOnWater(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 5, Y: 60, Z: 5}
	boat := SpawnBoat(w, pos)

	waterSurface := float32(63.0)
	sys := &BoatSystem{
		IsWaterAt:     func(_ mcmath.Vec3) bool { return true },
		WaterSurfaceY: func(_, _ float32) float32 { return waterSurface },
	}

	sys.Update(w, 0.05)

	tr, ok := ecs.GetStore[Transform](w).Get(boat)
	require.True(t, ok)
	assert.InDelta(t, waterSurface+boatWaterSurfaceOffset, tr.Position.Y, 0.01,
		"boat should float at water surface + offset")

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(boat)
	require.True(t, ok)
	assert.Equal(t, float32(0), pb.Body.Gravity,
		"gravity should be disabled on water")
}

func TestBoatLandSpeed(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 5, Y: 64, Z: 5}
	boat := SpawnBoat(w, pos)

	sys := &BoatSystem{
		IsWaterAt: func(_ mcmath.Vec3) bool { return false },
	}

	sys.Update(w, 0.05)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(boat)
	require.True(t, ok)
	assert.Equal(t, float32(28.0), pb.Body.Gravity,
		"gravity should be enabled on land")
}

func TestBoatMountDismount(t *testing.T) {
	w := ecs.NewWorld()
	boatPos := mcmath.Vec3{X: 5, Y: 64, Z: 5}
	boat := SpawnBoat(w, boatPos)
	rider := SpawnPlayer(w, "Steve", mcmath.Vec3{X: 6, Y: 64, Z: 5})

	// Mount rider.
	MountBoat(w, boat, rider)
	bd, ok := ecs.GetStore[BoatData](w).Get(boat)
	require.True(t, ok)
	assert.True(t, bd.HasRider)
	assert.Equal(t, rider, bd.Rider)

	// Cannot mount a second rider.
	rider2 := SpawnPlayer(w, "Alex", mcmath.Vec3{X: 7, Y: 64, Z: 5})
	MountBoat(w, boat, rider2)
	bd, _ = ecs.GetStore[BoatData](w).Get(boat)
	assert.Equal(t, rider, bd.Rider, "should not replace existing rider")

	// Dismount.
	DismountBoat(w, boat)
	bd, _ = ecs.GetStore[BoatData](w).Get(boat)
	assert.False(t, bd.HasRider)
}

func TestBoatWaterDetection(t *testing.T) {
	w := ecs.NewWorld()
	boat := SpawnBoat(w, mcmath.Vec3{X: 5, Y: 64, Z: 5})

	// Water at exact position.
	sys := &BoatSystem{
		IsWaterAt:     func(pos mcmath.Vec3) bool { return pos.Y >= 63.9 },
		WaterSurfaceY: func(_, _ float32) float32 { return 64 },
	}

	sys.Update(w, 0.05)

	tr, _ := ecs.GetStore[Transform](w).Get(boat)
	assert.InDelta(t, 64.0+boatWaterSurfaceOffset, tr.Position.Y, 0.01)

	// No water at all.
	sys2 := &BoatSystem{
		IsWaterAt: func(_ mcmath.Vec3) bool { return false },
	}

	sys2.Update(w, 0.05)

	pb, _ := ecs.GetStore[PhysicsBody](w).Get(boat)
	assert.Equal(t, float32(28.0), pb.Body.Gravity,
		"should detect no water and enable gravity")
}

func TestBoatBreakDropsItem(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 5, Y: 64, Z: 5}
	boat := SpawnBoat(w, pos)

	BreakBoat(w, boat)

	assert.False(t, w.Alive(boat), "boat entity should be destroyed")

	// Check that a boat item was spawned.
	var foundBoatItem bool
	ecs.GetStore[Inventory](w).Each(func(e ecs.Entity, inv *Inventory) {
		if inv.Slots[0].ItemID == item.Boat && inv.Slots[0].Count == 1 {
			foundBoatItem = true
		}
	})
	assert.True(t, foundBoatItem, "breaking boat should drop a boat item")
}

func TestBoatRiderMovesWithBoat(t *testing.T) {
	w := ecs.NewWorld()
	boatPos := mcmath.Vec3{X: 5, Y: 64, Z: 5}
	boat := SpawnBoat(w, boatPos)
	rider := SpawnPlayer(w, "Steve", mcmath.Vec3{X: 20, Y: 70, Z: 20})

	MountBoat(w, boat, rider)

	sys := &BoatSystem{
		IsWaterAt:     func(_ mcmath.Vec3) bool { return true },
		WaterSurfaceY: func(_, _ float32) float32 { return 64 },
	}

	sys.Update(w, 0.05)

	boatTr, _ := ecs.GetStore[Transform](w).Get(boat)
	riderTr, _ := ecs.GetStore[Transform](w).Get(rider)
	assert.Equal(t, boatTr.Position.X, riderTr.Position.X,
		"rider X should match boat X")
	assert.Equal(t, boatTr.Position.Z, riderTr.Position.Z,
		"rider Z should match boat Z")
	assert.InDelta(t, boatTr.Position.Y+boatBBox.Max.Y, riderTr.Position.Y, 0.01,
		"rider should sit on top of boat")
}
