//go:build !ci

package player

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupBreakingTest creates a Controller with camera at (0.5, 65.62, 0.5)
// looking in -Z direction (yaw=0, pitch=0), and a mock world with a block
// placed at stonePos directly in front of the player.
func setupBreakingTest(blockID uint16) (*Controller, *ecs.World, *mockBlockWorld, *input.Manager) {
	ecsWorld := ecs.NewWorld()
	e := entity.SpawnPlayer(ecsWorld, "test", mcmath.Vec3{X: 0.5, Y: 64, Z: 0.5})
	cam := render.NewCamera(mcmath.Vec3{X: 0.5, Y: 65.62, Z: 0.5})
	km := input.NewKeyMap()
	ctrl := NewController(e, ecsWorld, cam, km)

	bw := newMockBlockWorld()
	// Place the target block 2 blocks in front (-Z) at eye level.
	stonePos := mcmath.BlockPos{X: 0, Y: 65, Z: -1}
	bw.SetBlock(stonePos, blockID)

	mgr := input.NewManager()
	return ctrl, ecsWorld, bw, mgr
}

// TestBreakStone_SpawnsCobblestoneItemDrop verifies that breaking a stone
// block spawns a cobblestone item drop entity in the ECS world.
func TestBreakStone_SpawnsCobblestoneItemDrop(t *testing.T) {
	ctrl, ecsWorld, bw, mgr := setupBreakingTest(block.Stone)

	// Press and hold the attack (left mouse) button.
	attackBtn := ctrl.KeyMap.GetKey(input.Attack)
	mgr.MouseButtonCallback(attackBtn, input.ActionPress, 0)

	// First call: registers the target block and starts breaking.
	ctrl.updateBreaking(mgr, bw, 0.01)

	// Verify the block is still there (stone hardness=1.5, breakTime=2.25s).
	stonePos := mcmath.BlockPos{X: 0, Y: 65, Z: -1}
	assert.Equal(t, block.Stone, bw.GetBlock(stonePos), "block should still be stone mid-break")

	// Second call with enough dt to finish breaking (dt >= breakTime).
	ctrl.updateBreaking(mgr, bw, 3.0)

	// Block should now be air.
	assert.Equal(t, block.Air, bw.GetBlock(stonePos), "block should be air after breaking")

	// Verify a cobblestone item drop was spawned.
	var drops []entity.ItemDrop
	ecs.Query2[entity.ItemDrop, entity.Transform](ecsWorld, func(e ecs.Entity, d *entity.ItemDrop, tr *entity.Transform) {
		drops = append(drops, *d)
	})

	require.Len(t, drops, 1, "expected exactly one item drop entity")
	assert.Equal(t, item.Cobblestone, drops[0].ItemID, "stone should drop cobblestone")
	assert.Equal(t, 1, drops[0].Count)
}

// TestBreakBlock_CreativeMode_NoDrops verifies that breaking a block in
// creative mode does NOT spawn item drops.
func TestBreakBlock_CreativeMode_NoDrops(t *testing.T) {
	ctrl, ecsWorld, bw, mgr := setupBreakingTest(block.Stone)

	// Enable creative mode.
	ctrl.Mode = &creativeModeStub{}

	attackBtn := ctrl.KeyMap.GetKey(input.Attack)
	mgr.MouseButtonCallback(attackBtn, input.ActionPress, 0)

	// Creative mode: instant break on first call.
	ctrl.updateBreaking(mgr, bw, 0.01)

	stonePos := mcmath.BlockPos{X: 0, Y: 65, Z: -1}
	assert.Equal(t, block.Air, bw.GetBlock(stonePos), "creative mode should break instantly")

	// Verify no drops were spawned.
	var drops []entity.ItemDrop
	ecs.Query2[entity.ItemDrop, entity.Transform](ecsWorld, func(e ecs.Entity, d *entity.ItemDrop, tr *entity.Transform) {
		drops = append(drops, *d)
	})
	assert.Empty(t, drops, "creative mode should not spawn drops")
}

// TestBreakBlock_InstantBreak_SpawnsDrops verifies that breaking a solid block
// with hardness 0 (instant break) attempts to spawn item drops. TNT is solid
// with hardness 0; it breaks immediately and calls the drop path without error.
func TestBreakBlock_InstantBreak_SpawnsDrops(t *testing.T) {
	ctrl, _, bw, mgr := setupBreakingTest(block.TNT)

	attackBtn := ctrl.KeyMap.GetKey(input.Attack)
	mgr.MouseButtonCallback(attackBtn, input.ActionPress, 0)

	// TNT has hardness 0 and is solid; should break instantly on first call.
	ctrl.updateBreaking(mgr, bw, 0.01)

	tntPos := mcmath.BlockPos{X: 0, Y: 65, Z: -1}
	assert.Equal(t, block.Air, bw.GetBlock(tntPos), "hardness-0 block should break instantly")

	// Confirm breaking state was reset.
	assert.Nil(t, ctrl.BreakingBlock, "breaking state should be reset after instant break")
	assert.Equal(t, float32(0), ctrl.BreakProgress)
}

// TestPlacement_DecrementsItemCount verifies that placing a block in survival
// mode consumes 1 item from the selected hotbar slot.
func TestPlacement_DecrementsItemCount(t *testing.T) {
	ctrl, _, bw, mgr := setupBreakingTest(block.Air)

	// Give the player 10 dirt in hotbar slot 0.
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 10})
	ctrl.Inventory = inv
	ctrl.SelectedSlot = 0

	// Clear the block in front so placement succeeds.
	placeTarget := mcmath.BlockPos{X: 0, Y: 65, Z: -1}
	bw.SetBlock(placeTarget, block.Air)

	// Place a solid block next to the target so the raycast hits something.
	// The player looks at -Z; put a solid block at Z=-2 so the adjacent face
	// places at Z=-1.
	bw.SetBlock(mcmath.BlockPos{X: 0, Y: 65, Z: -2}, block.Stone)

	useBtn := ctrl.KeyMap.GetKey(input.Use)
	mgr.MouseButtonCallback(useBtn, input.ActionPress, 0)

	ctrl.updatePlacement(mgr, bw)

	slot := ctrl.Inventory.GetSlot(0)
	assert.Equal(t, 9, slot.Count, "placing a block should decrement item count by 1")
}

// TestPlacement_CreativeMode_DoesNotDecrement verifies that placing a block in
// creative mode does NOT consume items from the hotbar.
func TestPlacement_CreativeMode_DoesNotDecrement(t *testing.T) {
	ctrl, _, bw, mgr := setupBreakingTest(block.Air)

	// Enable creative mode.
	ctrl.Mode = &creativeModeStub{}

	// Give the player 10 dirt in hotbar slot 0.
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 10})
	ctrl.Inventory = inv
	ctrl.SelectedSlot = 0

	// Clear the block in front so placement succeeds.
	placeTarget := mcmath.BlockPos{X: 0, Y: 65, Z: -1}
	bw.SetBlock(placeTarget, block.Air)

	// Place a solid block at Z=-2 so raycast hits it and places at Z=-1.
	bw.SetBlock(mcmath.BlockPos{X: 0, Y: 65, Z: -2}, block.Stone)

	useBtn := ctrl.KeyMap.GetKey(input.Use)
	mgr.MouseButtonCallback(useBtn, input.ActionPress, 0)

	ctrl.updatePlacement(mgr, bw)

	slot := ctrl.Inventory.GetSlot(0)
	assert.Equal(t, 10, slot.Count, "creative mode should not consume items")
}

// creativeModeStub implements ModeChecker for creative mode testing.
type creativeModeStub struct{}

func (c *creativeModeStub) CanFly() bool            { return true }
func (c *creativeModeStub) CanBreakInstantly() bool  { return true }
func (c *creativeModeStub) HasInfiniteItems() bool   { return true }
func (c *creativeModeStub) CanTakeDamage() bool      { return false }
func (c *creativeModeStub) IsNoClip() bool           { return false }
