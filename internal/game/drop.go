package game

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

const (
	// dropForwardOffset is how far in front of the player (in blocks) the
	// dropped item spawns.
	dropForwardOffset float32 = 1.5
	// dropThrowSpeed is the initial forward velocity of a thrown item
	// (blocks per second).
	dropThrowSpeed float32 = 4.0
)

// handleItemDrop checks for the Q key press and drops a single item from
// the selected hotbar slot in front of the player.
func (g *Game) handleItemDrop() {
	if g.State.CurrentState() != GameStatePlaying {
		return
	}
	if g.Player == nil || g.Inventory == nil || g.ECSWorld == nil {
		return
	}

	dropKey := g.KeyMap.GetKey(input.DropItem)
	if !g.Input.IsKeyJustPressed(dropKey) {
		return
	}

	stack := g.Inventory.GetSlot(g.Player.SelectedSlot)
	if stack.IsEmpty() {
		return
	}

	g.Inventory.RemoveItem(g.Player.SelectedSlot, 1)

	// Compute spawn position: player position + forward * offset.
	transform := ecs.GetStore[entity.Transform](g.ECSWorld)
	t, ok := transform.Get(g.Player.Entity)
	if !ok {
		return
	}

	forward := g.Player.Camera.Forward()
	horizontalFwd := mcmath.Vec3{X: forward.X, Y: 0, Z: forward.Z}.Normalize()

	spawnPos := t.Position.Add(horizontalFwd.Scale(dropForwardOffset))
	// Raise to eye level so the item appears in front of the player's view.
	spawnPos.Y += 1.0

	e := entity.SpawnItemDrop(g.ECSWorld, spawnPos, stack.ItemID, 1)

	// Apply forward velocity so the item flies away from the player.
	pbStore := ecs.GetStore[entity.PhysicsBody](g.ECSWorld)
	if pb, ok := pbStore.Get(e); ok {
		pb.Body.Velocity = horizontalFwd.Scale(dropThrowSpeed)
	}
}
