package player

import (
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/physics"
)

const (
	// baseBreakTime is the base time multiplier for breaking blocks.
	// Actual break time = block.Hardness * baseBreakTime seconds.
	baseBreakTime float32 = 1.5
)

// interactionState holds mutable state for block breaking.
type interactionState struct {
	breakingBlockID block.BlockID
}

// UpdateInteraction handles block breaking and placement each frame.
func (c *Controller) UpdateInteraction(inp *input.Manager, w BlockWorld, dt float32) {
	c.updateBreaking(inp, w, dt)
	c.updatePlacement(inp, w)
}

// updateBreaking handles the progressive block-breaking mechanic when the
// attack button is held. Break progress accumulates based on block hardness;
// once it reaches 1.0, the block is removed and progress resets.
// In Creative mode, all breakable blocks break instantly.
func (c *Controller) updateBreaking(inp *input.Manager, w BlockWorld, dt float32) {
	attackBtn := c.KeyMap.GetKey(input.Attack)

	if !inp.IsMouseDown(attackBtn) {
		c.resetBreaking()
		return
	}

	hit, pos, _ := c.GetTargetBlock(w)
	if !hit {
		c.resetBreaking()
		return
	}

	// If the targeted block changed, reset progress.
	if c.BreakingBlock == nil || *c.BreakingBlock != pos {
		c.BreakingBlock = &pos
		c.BreakProgress = 0
		blockID := w.GetBlock(pos)
		c.interaction.breakingBlockID = blockID
	}

	props := block.GetProperties(c.interaction.breakingBlockID)

	// Unbreakable blocks (hardness < 0, e.g. bedrock).
	if props.Hardness < 0 {
		return
	}

	// Creative mode: instant break on all breakable blocks.
	if c.Mode != nil && c.Mode.CanBreakInstantly() {
		w.SetBlock(pos, block.Air)
		c.resetBreaking()
		return
	}

	// Instant-break blocks (hardness == 0).
	if props.Hardness == 0 {
		w.SetBlock(pos, block.Air)
		c.resetBreaking()
		return
	}

	breakTime := props.Hardness * baseBreakTime
	c.BreakProgress += dt / breakTime

	if c.BreakProgress >= 1.0 {
		w.SetBlock(pos, block.Air)
		c.resetBreaking()
	}
}

// updatePlacement handles block placement when the Use button is just pressed.
func (c *Controller) updatePlacement(inp *input.Manager, w BlockWorld) {
	useBtn := c.KeyMap.GetKey(input.Use)

	if !inp.IsMouseJustPressed(useBtn) {
		return
	}

	hit, pos, face := c.GetTargetBlock(w)
	if !hit {
		return
	}

	targetBlockID := w.GetBlock(pos)
	if targetBlockID == block.CraftingTable || targetBlockID == block.Furnace || targetBlockID == block.Chest {
		if c.OnUseBlock != nil {
			c.OnUseBlock(targetBlockID, pos)
		}
		return
	}

	// Jukebox interaction: right-click with disc inserts; empty-hand ejects.
	if block.IsJukebox(targetBlockID) && c.JukeboxMgr != nil {
		c.handleJukeboxInteraction(pos)
		return
	}

	// Calculate the adjacent block position using the face normal.
	normal := face.Normal()
	placePos := mcmath.BlockPos{
		X: pos.X + int32(normal.X),
		Y: pos.Y + int32(normal.Y),
		Z: pos.Z + int32(normal.Z),
	}

	// Do not place if the position already contains a solid block.
	if block.IsSolid(w.GetBlock(placePos)) {
		return
	}

	// Do not place inside the player's bounding box.
	placeAABB := mcmath.BlockAABB(placePos)
	playerAABB := c.PlayerAABB()
	if placeAABB.Intersects(playerAABB) {
		return
	}

	// Check that the selected hotbar item is a placeable block.
	selectedItem := c.getSelectedHotbarItem()
	if selectedItem.IsEmpty() {
		return
	}
	itemProps := item.GetProperties(selectedItem.ItemID)
	if !itemProps.IsBlock {
		return
	}

	w.SetBlock(placePos, itemProps.BlockID)
}

// getSelectedHotbarItem returns the item stack in the currently selected
// hotbar slot. Returns an empty stack if no inventory is available.
func (c *Controller) getSelectedHotbarItem() item.ItemStack {
	if c.Inventory == nil {
		return item.ItemStack{}
	}
	return c.Inventory.GetSlot(c.SelectedSlot)
}

// GetTargetBlock performs a raycast from the camera along the forward
// direction up to Reach distance and returns the first solid block hit.
func (c *Controller) GetTargetBlock(w BlockWorld) (hit bool, pos mcmath.BlockPos, face mcmath.Direction) {
	origin := c.Camera.GetPosition()
	direction := c.Camera.Forward()

	isSolid := func(bp mcmath.BlockPos) bool {
		return block.IsSolid(w.GetBlock(bp))
	}

	hit, pos, face, _ = physics.RaycastBlocks(origin, direction, c.Reach, isSolid)
	return hit, pos, face
}

// GetBreakProgress returns the current block breaking progress (0.0 to 1.0).
func (c *Controller) GetBreakProgress() float32 {
	return c.BreakProgress
}

// resetBreaking clears all breaking state.
func (c *Controller) resetBreaking() {
	c.BreakProgress = 0
	c.BreakingBlock = nil
	c.interaction.breakingBlockID = block.Air
}

const (
	// combatReach is the maximum distance for melee attacks against entities.
	combatReach float32 = 5.0
	// handDamage is the damage dealt by an empty-hand melee attack.
	handDamage float32 = 1.0
	// knockbackStrength is the horizontal knockback speed applied on hit.
	knockbackStrength float32 = 8.0
	// knockbackUpward is the upward velocity component of knockback.
	knockbackUpward float32 = 4.0
)

// UpdateCombat checks for a melee attack on the current frame. When the
// attack button is just pressed, it raycasts from the camera forward and
// finds the nearest entity within combatReach whose WorldAABB intersects the
// ray. On hit, a Damage component is applied to the target entity.
func (c *Controller) UpdateCombat(inp *input.Manager, ecsWorld *ecs.World, _ float32) {
	attackBtn := c.KeyMap.GetKey(input.Attack)
	if !inp.IsMouseJustPressed(attackBtn) {
		return
	}

	origin := c.Camera.GetPosition()
	direction := c.Camera.Forward()

	// Find the nearest entity whose world AABB intersects the ray.
	type hitResult struct {
		entity ecs.Entity
		dist   float32
		pos    mcmath.Vec3
	}

	var nearest *hitResult

	ecs.Query2[entity.Transform, entity.PhysicsBody](ecsWorld, func(e ecs.Entity, t *entity.Transform, pb *entity.PhysicsBody) {
		// Skip self.
		if e == c.Entity {
			return
		}

		aabb := pb.Body.WorldAABB()
		hit, dist := aabb.RayIntersects(origin, direction, combatReach)
		if !hit {
			return
		}

		if nearest == nil || dist < nearest.dist {
			nearest = &hitResult{entity: e, dist: dist, pos: t.Position}
		}
	})

	if nearest == nil {
		return
	}

	// Compute knockback direction away from the player.
	playerTransform := c.getTransform()
	if playerTransform == nil {
		return
	}

	kb := entity.KnockbackFromTo(playerTransform.Position, nearest.pos, knockbackStrength, knockbackUpward)

	ecs.GetStore[entity.Damage](ecsWorld).Set(nearest.entity, entity.Damage{
		Amount:    handDamage,
		Knockback: kb,
	})
}

// handleJukeboxInteraction handles right-clicking a jukebox.
// If the player holds a music disc, it is inserted and playback starts.
// If the jukebox is playing and the player's hand is empty, the disc is ejected.
func (c *Controller) handleJukeboxInteraction(pos mcmath.BlockPos) {
	selectedItem := c.getSelectedHotbarItem()

	if !selectedItem.IsEmpty() && block.IsDisc(selectedItem.ItemID) {
		if c.JukeboxMgr.InsertDisc(pos, selectedItem.ItemID) {
			// Consume one disc from the hotbar slot.
			if c.Inventory != nil {
				stack := c.Inventory.GetSlot(c.SelectedSlot)
				stack.Count--
				if stack.Count <= 0 {
					stack = item.ItemStack{}
				}
				c.Inventory.SetSlot(c.SelectedSlot, stack)
			}
		}
		return
	}

	if selectedItem.IsEmpty() && c.JukeboxMgr.IsPlaying(pos) {
		_ = c.JukeboxMgr.EjectDisc(pos)
		// The ejected disc would be spawned as an item entity in the world.
		// Item entity spawning is handled by the game layer.
	}
}
