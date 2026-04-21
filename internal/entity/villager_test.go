package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
)

func TestExecuteTrade_Success(t *testing.T) {
	data := &VillagerData{
		Prof:   Farmer,
		Trades: DefaultTrades(Farmer),
	}
	inv := inventory.NewInventory(36)
	inv.AddItem(item.ItemStack{ItemID: item.Wheat, Count: 20})

	ok := ExecuteTrade(data, 0, inv)
	assert.True(t, ok)
	assert.Equal(t, 1, data.Trades[0].Uses)

	// Player should have 1 emerald and 0 wheat.
	_, found := inv.FindItem(item.Emerald)
	assert.True(t, found)
	_, found = inv.FindItem(item.Wheat)
	assert.False(t, found)
}

func TestExecuteTrade_MaxUses(t *testing.T) {
	data := &VillagerData{
		Prof: Farmer,
		Trades: []TradeOffer{
			{
				InputItem1: item.ItemStack{ItemID: item.Wheat, Count: 1},
				Output:     item.ItemStack{ItemID: item.Emerald, Count: 1},
				MaxUses:    2,
			},
		},
	}
	inv := inventory.NewInventory(36)
	inv.AddItem(item.ItemStack{ItemID: item.Wheat, Count: 10})

	assert.True(t, ExecuteTrade(data, 0, inv))
	assert.True(t, ExecuteTrade(data, 0, inv))
	assert.False(t, ExecuteTrade(data, 0, inv)) // exhausted
	assert.Equal(t, 2, data.Trades[0].Uses)
}

func TestExecuteTrade_InsufficientItems(t *testing.T) {
	data := &VillagerData{
		Prof:   Farmer,
		Trades: DefaultTrades(Farmer),
	}
	inv := inventory.NewInventory(36)
	inv.AddItem(item.ItemStack{ItemID: item.Wheat, Count: 5}) // need 20

	ok := ExecuteTrade(data, 0, inv)
	assert.False(t, ok)
	assert.Equal(t, 0, data.Trades[0].Uses)
	// Wheat should not have been consumed.
	slot := inv.GetSlot(0)
	assert.Equal(t, 5, slot.Count)
}

func TestExecuteTrade_InvalidIndex(t *testing.T) {
	data := &VillagerData{
		Prof:   Farmer,
		Trades: DefaultTrades(Farmer),
	}
	inv := inventory.NewInventory(36)
	assert.False(t, ExecuteTrade(data, -1, inv))
	assert.False(t, ExecuteTrade(data, 99, inv))
}

func TestDefaultTrades_AllProfessions(t *testing.T) {
	profs := []Profession{Farmer, Librarian, Blacksmith, Cleric, Butcher}
	for _, prof := range profs {
		trades := DefaultTrades(prof)
		assert.NotEmpty(t, trades, "profession %d should have trades", prof)
		for _, tr := range trades {
			assert.Greater(t, tr.MaxUses, 0)
			assert.Greater(t, tr.Output.Count, 0)
			assert.Greater(t, tr.InputItem1.Count, 0)
		}
	}
}

func TestDefaultTrades_UnknownProfession(t *testing.T) {
	trades := DefaultTrades(Profession(255))
	assert.Nil(t, trades)
}

func TestSpawnVillager_Components(t *testing.T) {
	w := ecs.NewWorld()
	pos := mcmath.Vec3{X: 10, Y: 64, Z: 20}
	e := SpawnVillager(w, pos, Librarian)

	assert.True(t, w.Alive(e))

	tr, ok := ecs.GetStore[Transform](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, tr.Position)

	pb, ok := ecs.GetStore[PhysicsBody](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, pos, pb.Body.Position)

	h, ok := ecs.GetStore[Health](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, float32(20), h.Current)
	assert.Equal(t, float32(20), h.Max)

	et, ok := ecs.GetStore[EntityTypeComp](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, TypeVillager, et.Type)

	ai, ok := ecs.GetStore[AI](w).Get(e)
	assert.True(t, ok)
	assert.True(t, ai.Passive)
	assert.Equal(t, AIIdle, ai.State)

	vd, ok := ecs.GetStore[VillagerData](w).Get(e)
	assert.True(t, ok)
	assert.Equal(t, Librarian, vd.Prof)
	assert.Len(t, vd.Trades, 2)
}

func TestTradeOffer_IsExhausted(t *testing.T) {
	offer := TradeOffer{MaxUses: 3, Uses: 2}
	assert.False(t, offer.IsExhausted())
	offer.Uses = 3
	assert.True(t, offer.IsExhausted())
}
