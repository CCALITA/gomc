package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/player"
	"github.com/fanxiyao/gomc/internal/render"
)

// newDropTestGame creates a minimal Game suitable for testing handleItemDrop.
// The player is at (0,64,0) looking in the -Z direction.
func newDropTestGame() *Game {
	ecsWorld := ecs.NewWorld()
	spawnPos := mcmath.Vec3{X: 0, Y: 64, Z: 0}
	e := entity.SpawnPlayer(ecsWorld, "test", spawnPos)
	cam := render.NewCamera(spawnPos.Add(mcmath.Vec3{Y: player.EyeOffset}))

	km := input.NewKeyMap()
	ctrl := player.NewController(e, ecsWorld, cam, km)

	inv := inventory.NewInventory(36)
	ctrl.Inventory = inv

	mgr := input.NewManager()
	state := NewStateManager()
	state.SetState(GameStatePlaying)

	return &Game{
		ECSWorld:  ecsWorld,
		Player:    ctrl,
		Input:     mgr,
		KeyMap:    km,
		Inventory: inv,
		State:     state,
	}
}

func TestHandleItemDrop_DropsItem(t *testing.T) {
	g := newDropTestGame()

	// Place a stone item in hotbar slot 0.
	g.Inventory.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 5})
	g.Player.SelectedSlot = 0

	// Simulate Q key press.
	dropKey := g.KeyMap.GetKey(input.DropItem)
	g.Input.KeyCallback(dropKey, 0, input.ActionPress, 0)

	g.handleItemDrop()

	// Inventory should have 4 remaining.
	remaining := g.Inventory.GetSlot(0)
	assert.Equal(t, 4, remaining.Count)
	assert.Equal(t, item.Stone, remaining.ItemID)

	// An item drop entity should exist in the ECS world.
	var drops []entity.ItemDrop
	ecs.Query2[entity.ItemDrop, entity.Transform](g.ECSWorld, func(e ecs.Entity, d *entity.ItemDrop, tr *entity.Transform) {
		drops = append(drops, *d)
	})
	require.Len(t, drops, 1)
	assert.Equal(t, item.Stone, drops[0].ItemID)
	assert.Equal(t, 1, drops[0].Count)
}

func TestHandleItemDrop_EmptySlot_NoOp(t *testing.T) {
	g := newDropTestGame()

	// Slot 0 is empty by default.
	dropKey := g.KeyMap.GetKey(input.DropItem)
	g.Input.KeyCallback(dropKey, 0, input.ActionPress, 0)

	g.handleItemDrop()

	// No item drop entities should exist.
	var count int
	ecs.Query2[entity.ItemDrop, entity.Transform](g.ECSWorld, func(e ecs.Entity, d *entity.ItemDrop, tr *entity.Transform) {
		count++
	})
	assert.Equal(t, 0, count)
}

func TestHandleItemDrop_NotPlaying_NoOp(t *testing.T) {
	g := newDropTestGame()
	g.State.SetState(GameStatePaused)
	g.Inventory.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 5})

	dropKey := g.KeyMap.GetKey(input.DropItem)
	g.Input.KeyCallback(dropKey, 0, input.ActionPress, 0)

	g.handleItemDrop()

	// Inventory should be unchanged.
	assert.Equal(t, 5, g.Inventory.GetSlot(0).Count)
}

func TestHandleItemDrop_LastItem_ClearsSlot(t *testing.T) {
	g := newDropTestGame()
	g.Inventory.SetSlot(0, item.ItemStack{ItemID: item.Dirt, Count: 1})

	dropKey := g.KeyMap.GetKey(input.DropItem)
	g.Input.KeyCallback(dropKey, 0, input.ActionPress, 0)

	g.handleItemDrop()

	assert.True(t, g.Inventory.GetSlot(0).IsEmpty())
}

func TestHandleItemDrop_SpawnedEntityHasVelocity(t *testing.T) {
	g := newDropTestGame()
	g.Inventory.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 3})

	dropKey := g.KeyMap.GetKey(input.DropItem)
	g.Input.KeyCallback(dropKey, 0, input.ActionPress, 0)

	g.handleItemDrop()

	// Find the spawned entity and check its physics body has non-zero velocity.
	var hasVelocity bool
	ecs.Query2[entity.ItemDrop, entity.PhysicsBody](g.ECSWorld, func(e ecs.Entity, d *entity.ItemDrop, pb *entity.PhysicsBody) {
		vel := pb.Body.Velocity
		if vel.X != 0 || vel.Z != 0 {
			hasVelocity = true
		}
	})
	assert.True(t, hasVelocity, "dropped item should have forward velocity")
}

func TestHandleItemDrop_NoKeyPress_NoOp(t *testing.T) {
	g := newDropTestGame()
	g.Inventory.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 5})

	// Do NOT press Q.
	g.handleItemDrop()

	assert.Equal(t, 5, g.Inventory.GetSlot(0).Count)
}
