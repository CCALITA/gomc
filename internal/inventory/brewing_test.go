package inventory

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/item"
)

func TestBrewingStand_CompletesAfter400Ticks(t *testing.T) {
	bs := NewBrewingStand()
	bs.Fuel = 1
	bs.Ingredient = item.NewItemStack(item.NetherWart, 1)
	bs.PotionSlots[0] = item.NewItemStack(item.WaterBottle, 1)

	// 400 ticks * 0.05s = 20s
	bs.Update(20.0)

	if bs.PotionSlots[0].ItemID != item.AwkwardPotion {
		t.Errorf("expected AwkwardPotion (%d), got %d", item.AwkwardPotion, bs.PotionSlots[0].ItemID)
	}
}

func TestBrewingStand_CorrectRecipeOutputs(t *testing.T) {
	tests := []struct {
		name       string
		ingredient uint16
		input      uint16
		output     uint16
	}{
		{"WaterBottle+NetherWart=AwkwardPotion", item.NetherWart, item.WaterBottle, item.AwkwardPotion},
		{"AwkwardPotion+SpiderEye=PoisonPotion", item.SpiderEye, item.AwkwardPotion, item.PoisonPotion},
		{"AwkwardPotion+BlazePowder=StrengthPotion", item.BlazePowder, item.AwkwardPotion, item.StrengthPotion},
		{"AwkwardPotion+GhastTear=RegenPotion", item.GhastTear, item.AwkwardPotion, item.RegenPotion},
		{"AwkwardPotion+Sugar=SpeedPotion", item.Sugar, item.AwkwardPotion, item.SpeedPotion},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bs := NewBrewingStand()
			bs.Fuel = 1
			bs.Ingredient = item.NewItemStack(tc.ingredient, 1)
			bs.PotionSlots[0] = item.NewItemStack(tc.input, 1)

			bs.Update(20.0)

			if bs.PotionSlots[0].ItemID != tc.output {
				t.Errorf("expected %d, got %d", tc.output, bs.PotionSlots[0].ItemID)
			}
		})
	}
}

func TestBrewingStand_NoBrewWithoutFuel(t *testing.T) {
	bs := NewBrewingStand()
	bs.Fuel = 0
	bs.Ingredient = item.NewItemStack(item.NetherWart, 1)
	bs.PotionSlots[0] = item.NewItemStack(item.WaterBottle, 1)

	bs.Update(20.0)

	if bs.PotionSlots[0].ItemID != item.WaterBottle {
		t.Errorf("expected WaterBottle (no fuel), got %d", bs.PotionSlots[0].ItemID)
	}
	if bs.BrewTicks != 0 {
		t.Errorf("expected 0 brew ticks without fuel, got %d", bs.BrewTicks)
	}
}

func TestBrewingStand_FuelConsumption(t *testing.T) {
	bs := NewBrewingStand()
	bs.Fuel = 3
	bs.Ingredient = item.NewItemStack(item.NetherWart, 1)
	bs.PotionSlots[0] = item.NewItemStack(item.WaterBottle, 1)

	bs.Update(20.0)

	if bs.Fuel != 2 {
		t.Errorf("expected fuel=2 after one brew, got %d", bs.Fuel)
	}
}

func TestBrewingStand_InvalidIngredientRejected(t *testing.T) {
	bs := NewBrewingStand()
	bs.Fuel = 1
	// Use an item that is not a valid brewing ingredient for WaterBottle.
	bs.Ingredient = item.NewItemStack(item.Stick, 1)
	bs.PotionSlots[0] = item.NewItemStack(item.WaterBottle, 1)

	bs.Update(20.0)

	if bs.PotionSlots[0].ItemID != item.WaterBottle {
		t.Errorf("expected WaterBottle (invalid ingredient), got %d", bs.PotionSlots[0].ItemID)
	}
	if bs.BrewTicks != 0 {
		t.Errorf("expected 0 brew ticks for invalid ingredient, got %d", bs.BrewTicks)
	}
}
