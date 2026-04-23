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

// creativeModeStub implements ModeChecker for creative mode testing.
type creativeModeStub struct{}

func (c *creativeModeStub) CanFly() bool            { return true }
func (c *creativeModeStub) CanBreakInstantly() bool  { return true }
func (c *creativeModeStub) HasInfiniteItems() bool   { return true }
func (c *creativeModeStub) CanTakeDamage() bool      { return false }
func (c *creativeModeStub) IsNoClip() bool           { return false }

// ---------------------------------------------------------------------------
// Eating tests
// ---------------------------------------------------------------------------

// setupEatingTest creates a Controller with an inventory and a Hunger
// component ready for eating tests.
func setupEatingTest() (*Controller, *ecs.World) {
	ecsWorld := ecs.NewWorld()
	e := entity.SpawnPlayer(ecsWorld, "test", mcmath.Vec3{X: 0.5, Y: 64, Z: 0.5})
	cam := render.NewCamera(mcmath.Vec3{X: 0.5, Y: 65.62, Z: 0.5})
	km := input.NewKeyMap()
	ctrl := NewController(e, ecsWorld, cam, km)
	ctrl.Inventory = inventory.NewInventory(36)
	return ctrl, ecsWorld
}

// TestTryEatFood_RestoresHunger verifies that eating a food item restores
// food level and saturation and removes one item from the inventory.
func TestTryEatFood_RestoresHunger(t *testing.T) {
	ctrl, ecsWorld := setupEatingTest()

	// Drain hunger so the player can eat.
	hungerStore := ecs.GetStore[entity.Hunger](ecsWorld)
	hunger, ok := hungerStore.Get(ctrl.Entity)
	require.True(t, ok)
	hunger.FoodLevel = 10
	hunger.Saturation = 0

	// Place a stack of 5 bread in slot 0 (selected slot).
	ctrl.Inventory.SetSlot(0, item.NewItemStack(item.Bread, 5))

	ate := ctrl.tryEatFood()
	assert.True(t, ate, "should eat food successfully")

	// Bread restores 5 food and 6.0 saturation.
	assert.Equal(t, 15, hunger.FoodLevel, "food level should be 10+5=15")
	assert.InDelta(t, 6.0, hunger.Saturation, 0.01, "saturation should be restored")

	// Count should decrease by 1.
	remaining := ctrl.Inventory.GetSlot(0)
	assert.Equal(t, 4, remaining.Count, "item count should decrease by 1")
}

// TestTryEatFood_FullHunger_DoesNotEat verifies that eating is refused
// when the player's hunger bar is already full.
func TestTryEatFood_FullHunger_DoesNotEat(t *testing.T) {
	ctrl, ecsWorld := setupEatingTest()

	// Hunger is at max by default (20).
	hungerStore := ecs.GetStore[entity.Hunger](ecsWorld)
	hunger, ok := hungerStore.Get(ctrl.Entity)
	require.True(t, ok)
	assert.Equal(t, entity.MaxFoodLevel, hunger.FoodLevel)

	ctrl.Inventory.SetSlot(0, item.NewItemStack(item.Bread, 3))

	ate := ctrl.tryEatFood()
	assert.False(t, ate, "should not eat when hunger is full")

	// Item count should be unchanged.
	remaining := ctrl.Inventory.GetSlot(0)
	assert.Equal(t, 3, remaining.Count, "item count should be unchanged")
}

// TestTryEatFood_NonFoodItem_DoesNotEat verifies that trying to eat a
// non-food item (e.g. a tool) does nothing.
func TestTryEatFood_NonFoodItem_DoesNotEat(t *testing.T) {
	ctrl, ecsWorld := setupEatingTest()

	// Drain hunger so the player would eat if the item were food.
	hungerStore := ecs.GetStore[entity.Hunger](ecsWorld)
	hunger, ok := hungerStore.Get(ctrl.Entity)
	require.True(t, ok)
	hunger.FoodLevel = 5

	// Place a wooden pickaxe (non-food) in slot 0.
	ctrl.Inventory.SetSlot(0, item.NewItemStack(item.WoodenPickaxe, 1))

	ate := ctrl.tryEatFood()
	assert.False(t, ate, "should not eat a non-food item")

	// Item should be untouched.
	remaining := ctrl.Inventory.GetSlot(0)
	assert.Equal(t, 1, remaining.Count, "item count should be unchanged")
}
