package item

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Properties helpers
// ---------------------------------------------------------------------------

func TestGetProperties_BlockItem(t *testing.T) {
	p := GetProperties(Stone)
	assert.Equal(t, "Stone", p.Name)
	assert.Equal(t, DefaultMaxStackSize, p.MaxStackSize)
	assert.True(t, p.IsBlock)
	assert.Equal(t, uint16(Stone), p.BlockID)
	assert.Equal(t, ToolNone, p.ToolType)
	assert.Equal(t, 0, p.Durability)
}

func TestGetProperties_Tool(t *testing.T) {
	p := GetProperties(DiamondPickaxe)
	assert.Equal(t, "Diamond Pickaxe", p.Name)
	assert.Equal(t, 1, p.MaxStackSize)
	assert.Equal(t, 1562, p.Durability)
	assert.False(t, p.IsBlock)
	assert.Equal(t, ToolPickaxe, p.ToolType)
	assert.Equal(t, LevelDiamond, p.ToolLevel)
}

func TestGetProperties_Material(t *testing.T) {
	p := GetProperties(Stick)
	assert.Equal(t, "Stick", p.Name)
	assert.Equal(t, DefaultMaxStackSize, p.MaxStackSize)
	assert.False(t, p.IsBlock)
	assert.Equal(t, ToolNone, p.ToolType)
}

func TestGetProperties_Bucket(t *testing.T) {
	p := GetProperties(Bucket)
	assert.Equal(t, "Bucket", p.Name)
	assert.Equal(t, 16, p.MaxStackSize)

	pw := GetProperties(WaterBucket)
	assert.Equal(t, "Water Bucket", pw.Name)
	assert.Equal(t, 1, pw.MaxStackSize)

	pl := GetProperties(LavaBucket)
	assert.Equal(t, "Lava Bucket", pl.Name)
	assert.Equal(t, 1, pl.MaxStackSize)
}

func TestGetProperties_UnknownID(t *testing.T) {
	p := GetProperties(9999)
	assert.Equal(t, DefaultMaxStackSize, p.MaxStackSize)
	assert.Equal(t, ToolNone, p.ToolType)
	assert.Equal(t, "", p.Name)
}

func TestIsStackable(t *testing.T) {
	assert.True(t, IsStackable(Dirt))
	assert.True(t, IsStackable(Stick))
	assert.True(t, IsStackable(Bucket))
	assert.False(t, IsStackable(DiamondSword))
	assert.False(t, IsStackable(WaterBucket))
}

func TestIsTool(t *testing.T) {
	assert.True(t, IsTool(WoodenPickaxe))
	assert.True(t, IsTool(IronAxe))
	assert.True(t, IsTool(DiamondHoe))
	assert.False(t, IsTool(Stone))
	assert.False(t, IsTool(Stick))
}

func TestMaxStack(t *testing.T) {
	assert.Equal(t, DefaultMaxStackSize, MaxStack(Dirt))
	assert.Equal(t, 1, MaxStack(DiamondPickaxe))
	assert.Equal(t, 16, MaxStack(Bucket))
}

// ---------------------------------------------------------------------------
// All block items registered
// ---------------------------------------------------------------------------

func TestAllBlockItems(t *testing.T) {
	blocks := []ItemID{
		Stone, Dirt, Grass, Sand, Gravel, OakLog, OakPlanks,
		Cobblestone, Glass, OakLeaves, IronOre, GoldOre,
		DiamondOre, CoalOre, Bedrock, Water, Lava,
		CraftingTable, Furnace, Chest, TNT, Obsidian, Torch,
	}
	for _, id := range blocks {
		p := GetProperties(id)
		assert.True(t, p.IsBlock, "expected IsBlock for id %d (%s)", id, p.Name)
		assert.Equal(t, id, p.BlockID)
	}
}

// ---------------------------------------------------------------------------
// All tools registered
// ---------------------------------------------------------------------------

func TestAllTools(t *testing.T) {
	tools := []struct {
		id       ItemID
		toolType string
		level    int
	}{
		{WoodenPickaxe, ToolPickaxe, LevelWood},
		{StonePickaxe, ToolPickaxe, LevelStone},
		{IronPickaxe, ToolPickaxe, LevelIron},
		{DiamondPickaxe, ToolPickaxe, LevelDiamond},

		{WoodenAxe, ToolAxe, LevelWood},
		{StoneAxe, ToolAxe, LevelStone},
		{IronAxe, ToolAxe, LevelIron},
		{DiamondAxe, ToolAxe, LevelDiamond},

		{WoodenShovel, ToolShovel, LevelWood},
		{StoneShovel, ToolShovel, LevelStone},
		{IronShovel, ToolShovel, LevelIron},
		{DiamondShovel, ToolShovel, LevelDiamond},

		{WoodenSword, ToolSword, LevelWood},
		{StoneSword, ToolSword, LevelStone},
		{IronSword, ToolSword, LevelIron},
		{DiamondSword, ToolSword, LevelDiamond},

		{WoodenHoe, ToolHoe, LevelWood},
		{StoneHoe, ToolHoe, LevelStone},
		{IronHoe, ToolHoe, LevelIron},
		{DiamondHoe, ToolHoe, LevelDiamond},
	}
	for _, tc := range tools {
		p := GetProperties(tc.id)
		assert.Equal(t, tc.toolType, p.ToolType, "tool type for %s", p.Name)
		assert.Equal(t, tc.level, p.ToolLevel, "tool level for %s", p.Name)
		assert.Equal(t, 1, p.MaxStackSize, "max stack for %s", p.Name)
		assert.Greater(t, p.Durability, 0, "durability for %s", p.Name)
	}
}

// ---------------------------------------------------------------------------
// ItemStack
// ---------------------------------------------------------------------------

func TestNewItemStack_Block(t *testing.T) {
	s := NewItemStack(Dirt, 32)
	assert.Equal(t, Dirt, s.ItemID)
	assert.Equal(t, 32, s.Count)
	assert.Equal(t, 0, s.Durability)
}

func TestNewItemStack_Tool(t *testing.T) {
	s := NewItemStack(IronPickaxe, 1)
	assert.Equal(t, IronPickaxe, s.ItemID)
	assert.Equal(t, 1, s.Count)
	assert.Equal(t, 251, s.Durability)
}

func TestItemStack_IsEmpty(t *testing.T) {
	assert.True(t, ItemStack{}.IsEmpty())
	assert.True(t, ItemStack{Count: 0}.IsEmpty())
	assert.True(t, ItemStack{Count: -1}.IsEmpty())
	assert.False(t, NewItemStack(Dirt, 1).IsEmpty())
}

func TestItemStack_CanStackWith(t *testing.T) {
	a := NewItemStack(Dirt, 10)
	b := NewItemStack(Dirt, 5)
	assert.True(t, a.CanStackWith(b))

	// Different items
	c := NewItemStack(Stone, 5)
	assert.False(t, a.CanStackWith(c))

	// Tools cannot stack
	t1 := NewItemStack(IronPickaxe, 1)
	t2 := NewItemStack(IronPickaxe, 1)
	assert.False(t, t1.CanStackWith(t2))
}

func TestItemStack_CanStackWith_Bucket(t *testing.T) {
	// Empty buckets can stack (max 16)
	a := NewItemStack(Bucket, 5)
	b := NewItemStack(Bucket, 3)
	assert.True(t, a.CanStackWith(b))

	// Filled buckets cannot stack (max 1)
	wa := NewItemStack(WaterBucket, 1)
	wb := NewItemStack(WaterBucket, 1)
	assert.False(t, wa.CanStackWith(wb))
}

func TestItemStack_Merge_Basic(t *testing.T) {
	a := NewItemStack(Dirt, 50)
	b := NewItemStack(Dirt, 20)
	remaining := a.Merge(b)
	assert.Equal(t, 64, a.Count)
	assert.Equal(t, 6, remaining.Count)
	assert.Equal(t, Dirt, remaining.ItemID)
}

func TestItemStack_Merge_ExactFit(t *testing.T) {
	a := NewItemStack(Dirt, 32)
	b := NewItemStack(Dirt, 32)
	remaining := a.Merge(b)
	assert.Equal(t, 64, a.Count)
	assert.Equal(t, 0, remaining.Count)
}

func TestItemStack_Merge_AlreadyFull(t *testing.T) {
	a := NewItemStack(Dirt, 64)
	b := NewItemStack(Dirt, 10)
	remaining := a.Merge(b)
	assert.Equal(t, 64, a.Count)
	assert.Equal(t, 10, remaining.Count)
}

func TestItemStack_Merge_Incompatible(t *testing.T) {
	a := NewItemStack(Dirt, 10)
	b := NewItemStack(Stone, 10)
	remaining := a.Merge(b)
	assert.Equal(t, 10, a.Count)
	assert.Equal(t, 10, remaining.Count)
	assert.Equal(t, Stone, remaining.ItemID)
}

func TestItemStack_Merge_Tools(t *testing.T) {
	a := NewItemStack(IronPickaxe, 1)
	b := NewItemStack(IronPickaxe, 1)
	remaining := a.Merge(b)
	assert.Equal(t, 1, a.Count)
	assert.Equal(t, 1, remaining.Count)
}

func TestItemStack_Split_Normal(t *testing.T) {
	s := NewItemStack(Dirt, 64)
	taken, remaining := s.Split(32)
	assert.Equal(t, 32, taken.Count)
	assert.Equal(t, 32, remaining.Count)
	assert.Equal(t, Dirt, taken.ItemID)
	assert.Equal(t, Dirt, remaining.ItemID)
}

func TestItemStack_Split_All(t *testing.T) {
	s := NewItemStack(Dirt, 10)
	taken, remaining := s.Split(10)
	assert.Equal(t, 10, taken.Count)
	assert.Equal(t, 0, remaining.Count)
}

func TestItemStack_Split_MoreThanAvailable(t *testing.T) {
	s := NewItemStack(Dirt, 5)
	taken, remaining := s.Split(100)
	assert.Equal(t, 5, taken.Count)
	assert.Equal(t, 0, remaining.Count)
}

func TestItemStack_Split_Zero(t *testing.T) {
	s := NewItemStack(Dirt, 10)
	taken, remaining := s.Split(0)
	assert.Equal(t, 0, taken.Count)
	assert.Equal(t, 10, remaining.Count)
}

func TestItemStack_Split_Negative(t *testing.T) {
	s := NewItemStack(Dirt, 10)
	taken, remaining := s.Split(-5)
	assert.Equal(t, 0, taken.Count)
	assert.Equal(t, 10, remaining.Count)
}

func TestItemStack_Split_Tool(t *testing.T) {
	s := NewItemStack(DiamondSword, 1)
	taken, remaining := s.Split(1)
	assert.Equal(t, 1, taken.Count)
	assert.Equal(t, 1562, taken.Durability)
	assert.Equal(t, 0, remaining.Count)
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

func TestInitRegistry(t *testing.T) {
	InitRegistry()
	assert.NotNil(t, Items)
	assert.Greater(t, Items.Len(), 0)
}

func TestInitRegistry_LookupByName(t *testing.T) {
	InitRegistry()

	props, _, ok := Items.GetByName("Stone")
	assert.True(t, ok)
	assert.Equal(t, "Stone", props.Name)
	assert.True(t, props.IsBlock)

	props, _, ok = Items.GetByName("Diamond Pickaxe")
	assert.True(t, ok)
	assert.Equal(t, ToolPickaxe, props.ToolType)
}

func TestInitRegistry_FreezeRejectsNew(t *testing.T) {
	InitRegistry()
	_, err := Items.Register("Fake Item", ItemProperties{Name: "Fake Item"})
	assert.Error(t, err)
}

func TestInitRegistry_AllPropertiesRegistered(t *testing.T) {
	InitRegistry()
	// Every entry in our properties map should be findable by name.
	for _, props := range properties {
		_, _, ok := Items.GetByName(props.Name)
		assert.True(t, ok, "expected %q in registry", props.Name)
	}
}

// ---------------------------------------------------------------------------
// Edge-case coverage boosters
// ---------------------------------------------------------------------------

func TestItemConstants_BlockRange(t *testing.T) {
	// Block IDs should start at 0 (Air) and go up to Torch (23).
	assert.Equal(t, ItemID(0), Air)
	assert.Equal(t, ItemID(23), Torch)
}

func TestItemConstants_ToolRange(t *testing.T) {
	assert.Equal(t, ItemID(100), WoodenPickaxe)
	assert.Equal(t, ItemID(127), LavaBucket)
}

func TestToolLevelConstants(t *testing.T) {
	assert.Equal(t, 0, LevelHand)
	assert.Equal(t, 1, LevelWood)
	assert.Equal(t, 2, LevelStone)
	assert.Equal(t, 3, LevelIron)
	assert.Equal(t, 4, LevelDiamond)
}

func TestToolTypeConstants(t *testing.T) {
	assert.Equal(t, "none", ToolNone)
	assert.Equal(t, "pickaxe", ToolPickaxe)
	assert.Equal(t, "axe", ToolAxe)
	assert.Equal(t, "shovel", ToolShovel)
	assert.Equal(t, "sword", ToolSword)
	assert.Equal(t, "hoe", ToolHoe)
}

func TestAllMaterialsRegistered(t *testing.T) {
	materials := []ItemID{Stick, Coal, IronIngot, GoldIngot, Diamond}
	for _, id := range materials {
		p := GetProperties(id)
		assert.NotEmpty(t, p.Name, "expected name for material id %d", id)
		assert.Equal(t, DefaultMaxStackSize, p.MaxStackSize)
		assert.False(t, p.IsBlock)
		assert.Equal(t, ToolNone, p.ToolType)
	}
}

func TestMerge_PartialTransfer(t *testing.T) {
	a := NewItemStack(Cobblestone, 60)
	b := NewItemStack(Cobblestone, 10)
	remaining := a.Merge(b)
	assert.Equal(t, 64, a.Count)
	assert.Equal(t, 6, remaining.Count)
}

func TestMerge_BucketStack(t *testing.T) {
	a := NewItemStack(Bucket, 10)
	b := NewItemStack(Bucket, 8)
	remaining := a.Merge(b)
	assert.Equal(t, 16, a.Count)
	assert.Equal(t, 2, remaining.Count)
}
