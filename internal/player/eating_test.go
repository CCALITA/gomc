package player

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/stretchr/testify/assert"
)

// helpers ----------------------------------------------------------------

// pressUseButton simulates holding the right mouse button.
func pressUseButton(mgr *input.Manager) {
	mgr.MouseButtonCallback(input.MouseButtonRight, input.ActionPress, 0)
}

// releaseUseButton simulates releasing the right mouse button.
func releaseUseButton(mgr *input.Manager) {
	mgr.MouseButtonCallback(input.MouseButtonRight, input.ActionRelease, 0)
}

// pressAttackButton simulates pressing the left mouse button.
func pressAttackButton(mgr *input.Manager) {
	mgr.MouseButtonCallback(input.MouseButtonLeft, input.ActionPress, 0)
}

// makeHungryPlayer returns a Hunger component with depleted food level.
func makeHungryPlayer(foodLevel int) entity.Hunger {
	h := entity.NewHunger()
	h.FoodLevel = foodLevel
	h.Saturation = 0
	return h
}

// setupEatingTest returns common test fixtures: input manager, keymap,
// inventory with an apple in slot 0, and a hungry hunger component.
func setupEatingTest() (*input.Manager, *input.KeyMap, *inventory.Inventory, entity.Hunger) {
	mgr := input.NewManager()
	km := input.NewKeyMap()
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.NewItemStack(item.Apple, 5))
	hunger := makeHungryPlayer(10) // half-hungry
	return mgr, km, inv, hunger
}

// tests ------------------------------------------------------------------

func TestStartEating_RightClickWithFoodAndHungry(t *testing.T) {
	mgr, km, inv, hunger := setupEatingTest()
	es := NewEatingState()

	pressUseButton(mgr)
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)

	assert.True(t, es.Eating, "should start eating")
	assert.Equal(t, item.Apple, es.ItemID)
	assert.Equal(t, 0, es.EatTicks)
}

func TestEatCompletion_RestoresHunger(t *testing.T) {
	mgr, km, inv, hunger := setupEatingTest()
	hunger.FoodLevel = 10
	hunger.Saturation = 0.0

	pressUseButton(mgr)
	es := NewEatingState()

	// Tick through all 32 eating ticks.
	for i := 0; i < MaxEatTicks+1; i++ {
		mgr.Update()
		pressUseButton(mgr)
		es = UpdateEating(es, mgr, km, inv, 0, &hunger)
	}

	assert.False(t, es.Eating, "should finish eating")
	assert.Equal(t, 14, hunger.FoodLevel, "apple restores 4 food points")
	assert.InDelta(t, 2.4, hunger.Saturation, 0.01, "apple restores 2.4 saturation")

	// Stack should be decremented.
	stack := inv.GetSlot(0)
	assert.Equal(t, 4, stack.Count, "one apple consumed")
}

func TestCannotEatAtFullHunger(t *testing.T) {
	mgr, km, inv, _ := setupEatingTest()
	hunger := entity.NewHunger() // full (20/20)

	pressUseButton(mgr)
	es := NewEatingState()
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)

	assert.False(t, es.Eating, "should not eat when hunger is full")
}

func TestEatingCancelled_OnRelease(t *testing.T) {
	mgr, km, inv, hunger := setupEatingTest()

	pressUseButton(mgr)
	es := NewEatingState()
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)
	assert.True(t, es.Eating)

	// Release right-click.
	mgr.Update()
	releaseUseButton(mgr)
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)

	assert.False(t, es.Eating, "eating should cancel on release")
	assert.Equal(t, 10, hunger.FoodLevel, "hunger unchanged on cancel")
}

func TestEatingCancelled_OnLeftClick(t *testing.T) {
	mgr, km, inv, hunger := setupEatingTest()

	pressUseButton(mgr)
	es := NewEatingState()
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)
	assert.True(t, es.Eating)

	// Press attack while still holding use.
	mgr.Update()
	pressUseButton(mgr)
	pressAttackButton(mgr)
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)

	assert.False(t, es.Eating, "eating should cancel on left-click")
}

func TestCorrectFoodValues_Steak(t *testing.T) {
	mgr := input.NewManager()
	km := input.NewKeyMap()
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.NewItemStack(item.Steak, 1))
	hunger := makeHungryPlayer(5)

	pressUseButton(mgr)
	es := NewEatingState()

	for i := 0; i < MaxEatTicks+1; i++ {
		mgr.Update()
		pressUseButton(mgr)
		es = UpdateEating(es, mgr, km, inv, 0, &hunger)
	}

	assert.False(t, es.Eating)
	assert.Equal(t, 13, hunger.FoodLevel, "steak restores 8 food points")
	assert.InDelta(t, 12.8, hunger.Saturation, 0.01, "steak restores 12.8 saturation")

	// Last item consumed - slot should be empty.
	stack := inv.GetSlot(0)
	assert.True(t, stack.IsEmpty(), "last steak consumed")
}

func TestCorrectFoodValues_Cookie(t *testing.T) {
	mgr := input.NewManager()
	km := input.NewKeyMap()
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.NewItemStack(item.Cookie, 3))
	hunger := makeHungryPlayer(15)

	pressUseButton(mgr)
	es := NewEatingState()

	for i := 0; i < MaxEatTicks+1; i++ {
		mgr.Update()
		pressUseButton(mgr)
		es = UpdateEating(es, mgr, km, inv, 0, &hunger)
	}

	assert.False(t, es.Eating)
	assert.Equal(t, 17, hunger.FoodLevel, "cookie restores 2 food points")
	assert.InDelta(t, 0.4, hunger.Saturation, 0.01, "cookie restores 0.4 saturation")
}

func TestCorrectFoodValues_GoldenApple(t *testing.T) {
	mgr := input.NewManager()
	km := input.NewKeyMap()
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.NewItemStack(item.GoldenApple, 1))
	hunger := makeHungryPlayer(12)

	pressUseButton(mgr)
	es := NewEatingState()

	for i := 0; i < MaxEatTicks+1; i++ {
		mgr.Update()
		pressUseButton(mgr)
		es = UpdateEating(es, mgr, km, inv, 0, &hunger)
	}

	assert.False(t, es.Eating)
	assert.Equal(t, 16, hunger.FoodLevel, "golden apple restores 4 food points")
	assert.InDelta(t, 9.6, hunger.Saturation, 0.01, "golden apple restores 9.6 saturation")
}

func TestCannotEatNonFoodItem(t *testing.T) {
	mgr := input.NewManager()
	km := input.NewKeyMap()
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.NewItemStack(item.Dirt, 10))
	hunger := makeHungryPlayer(10)

	pressUseButton(mgr)
	es := NewEatingState()
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)

	assert.False(t, es.Eating, "should not eat non-food item")
}

func TestCannotEatEmptySlot(t *testing.T) {
	mgr := input.NewManager()
	km := input.NewKeyMap()
	inv := inventory.NewInventory(9)
	hunger := makeHungryPlayer(10)

	pressUseButton(mgr)
	es := NewEatingState()
	es = UpdateEating(es, mgr, km, inv, 0, &hunger)

	assert.False(t, es.Eating, "should not eat from empty slot")
}

func TestEatingState_Tick(t *testing.T) {
	es := StartEating(item.Apple)

	next, completed := es.Tick()
	assert.True(t, next.Eating)
	assert.Equal(t, 1, next.EatTicks)
	assert.False(t, completed)

	// Tick to completion.
	for i := 1; i < MaxEatTicks-1; i++ {
		next, completed = next.Tick()
		assert.False(t, completed)
	}
	next, completed = next.Tick()
	assert.True(t, completed, "should complete after MaxEatTicks")
	assert.False(t, next.Eating, "state should reset after completion")
}

func TestEatingState_Cancel(t *testing.T) {
	es := StartEating(item.Bread)
	es, _ = es.Tick()
	es, _ = es.Tick()

	cancelled := es.Cancel()
	assert.False(t, cancelled.Eating)
	assert.Equal(t, 0, cancelled.EatTicks)
	assert.Equal(t, MaxEatTicks, cancelled.MaxEatTicks)
}

func TestEatingState_TickWhenNotEating(t *testing.T) {
	es := NewEatingState()
	next, completed := es.Tick()
	assert.False(t, completed)
	assert.False(t, next.Eating)
}

func TestNewFoodItems_RawBeef(t *testing.T) {
	p := item.GetProperties(item.RawBeef)
	assert.Equal(t, "Raw Beef", p.Name)
	assert.Equal(t, 3, p.FoodRestore)
	assert.InDelta(t, 1.8, p.FoodSaturation, 0.01)
	assert.True(t, item.IsFood(item.RawBeef))
}

func TestNewFoodItems_RawPorkchop(t *testing.T) {
	p := item.GetProperties(item.RawPorkchop)
	assert.Equal(t, "Raw Porkchop", p.Name)
	assert.Equal(t, 3, p.FoodRestore)
	assert.InDelta(t, 1.8, p.FoodSaturation, 0.01)
}

func TestNewFoodItems_RawChicken(t *testing.T) {
	p := item.GetProperties(item.RawChicken)
	assert.Equal(t, "Raw Chicken", p.Name)
	assert.Equal(t, 2, p.FoodRestore)
	assert.InDelta(t, 1.2, p.FoodSaturation, 0.01)
}

func TestNewFoodItems_CookedChicken(t *testing.T) {
	p := item.GetProperties(item.CookedChicken)
	assert.Equal(t, "Cooked Chicken", p.Name)
	assert.Equal(t, 6, p.FoodRestore)
	assert.InDelta(t, 7.2, p.FoodSaturation, 0.01)
}

func TestHungerClampedAtMax(t *testing.T) {
	mgr := input.NewManager()
	km := input.NewKeyMap()
	inv := inventory.NewInventory(9)
	inv.SetSlot(0, item.NewItemStack(item.Steak, 1))
	hunger := makeHungryPlayer(18) // nearly full

	pressUseButton(mgr)
	es := NewEatingState()

	for i := 0; i < MaxEatTicks+1; i++ {
		mgr.Update()
		pressUseButton(mgr)
		es = UpdateEating(es, mgr, km, inv, 0, &hunger)
	}

	assert.Equal(t, entity.MaxFoodLevel, hunger.FoodLevel, "food level clamped at max")
}
