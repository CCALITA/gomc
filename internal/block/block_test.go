package block

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/registry"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// block.go -- constant IDs
// ---------------------------------------------------------------------------

func TestBlockIDConstants(t *testing.T) {
	assert.Equal(t, BlockID(0), Air)
	assert.Equal(t, BlockID(1), Stone)
	assert.Equal(t, BlockID(2), Dirt)
	assert.Equal(t, BlockID(3), Grass)
	assert.Equal(t, BlockID(4), Sand)
	assert.Equal(t, BlockID(5), Gravel)
	assert.Equal(t, BlockID(6), OakLog)
	assert.Equal(t, BlockID(7), OakLeaves)
	assert.Equal(t, BlockID(8), OakPlanks)
	assert.Equal(t, BlockID(9), Cobblestone)
	assert.Equal(t, BlockID(10), Glass)
	assert.Equal(t, BlockID(11), Water)
	assert.Equal(t, BlockID(12), Lava)
	assert.Equal(t, BlockID(13), IronOre)
	assert.Equal(t, BlockID(14), CoalOre)
	assert.Equal(t, BlockID(15), DiamondOre)
	assert.Equal(t, BlockID(16), GoldOre)
	assert.Equal(t, BlockID(17), Bedrock)
	assert.Equal(t, BlockID(18), CraftingTable)
	assert.Equal(t, BlockID(19), Furnace)
	assert.Equal(t, BlockID(20), Chest)
	assert.Equal(t, BlockID(21), Torch)
	assert.Equal(t, BlockID(22), Obsidian)
	assert.Equal(t, BlockID(23), Sandstone)
}

// ---------------------------------------------------------------------------
// properties.go -- GetProperties / IsSolid / IsTransparent / IsLightSource
// ---------------------------------------------------------------------------

func TestGetProperties_KnownBlock(t *testing.T) {
	p := GetProperties(Stone)
	assert.Equal(t, "stone", p.Name)
	assert.True(t, p.Solid)
	assert.False(t, p.Transparent)
	assert.Equal(t, float32(1.5), p.Hardness)
	assert.Equal(t, float32(6), p.BlastResistance)
	assert.Equal(t, uint8(0), p.LightEmission)
	assert.Equal(t, uint8(15), p.LightFilter)
}

func TestGetProperties_Air(t *testing.T) {
	p := GetProperties(Air)
	assert.Equal(t, "air", p.Name)
	assert.False(t, p.Solid)
	assert.True(t, p.Transparent)
}

func TestGetProperties_UnknownBlock(t *testing.T) {
	p := GetProperties(9999)
	assert.Equal(t, BlockProperties{}, p)
}

func TestIsSolid(t *testing.T) {
	tests := []struct {
		id   BlockID
		want bool
	}{
		{Air, false},
		{Stone, true},
		{Dirt, true},
		{Water, false},
		{Lava, false},
		{Torch, false},
		{Bedrock, true},
		{Glass, true},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, IsSolid(tc.id), "IsSolid(%d)", tc.id)
	}
}

func TestIsTransparent(t *testing.T) {
	tests := []struct {
		id   BlockID
		want bool
	}{
		{Air, true},
		{Stone, false},
		{Glass, true},
		{Water, true},
		{OakLeaves, true},
		{Torch, true},
		{Lava, true},
		{Dirt, false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, IsTransparent(tc.id), "IsTransparent(%d)", tc.id)
	}
}

func TestIsLightSource(t *testing.T) {
	tests := []struct {
		id   BlockID
		want bool
	}{
		{Stone, false},
		{Air, false},
		{Lava, true},
		{Torch, true},
		{Glass, false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, IsLightSource(tc.id), "IsLightSource(%d)", tc.id)
	}
}

func TestAllBlocksHaveProperties(t *testing.T) {
	for _, id := range blockOrder {
		p := GetProperties(id)
		assert.NotEmpty(t, p.Name, "block %d has empty name", id)
	}
}

func TestBedrockSpecialValues(t *testing.T) {
	p := GetProperties(Bedrock)
	assert.Equal(t, float32(-1), p.Hardness)
	assert.Equal(t, float32(3600000), p.BlastResistance)
}

// ---------------------------------------------------------------------------
// face.go -- BlockFace, FaceUV, FaceVertices
// ---------------------------------------------------------------------------

func TestBlockFaceDirectionMapping(t *testing.T) {
	assert.Equal(t, mcmath.North, FaceNorth.Direction())
	assert.Equal(t, mcmath.South, FaceSouth.Direction())
	assert.Equal(t, mcmath.East, FaceEast.Direction())
	assert.Equal(t, mcmath.West, FaceWest.Direction())
	assert.Equal(t, mcmath.Up, FaceTop.Direction())
	assert.Equal(t, mcmath.Down, FaceBottom.Direction())
}

func TestAllFaces(t *testing.T) {
	faces := AllFaces()
	assert.Len(t, faces, 6)
	assert.Contains(t, faces, FaceNorth)
	assert.Contains(t, faces, FaceSouth)
	assert.Contains(t, faces, FaceEast)
	assert.Contains(t, faces, FaceWest)
	assert.Contains(t, faces, FaceTop)
	assert.Contains(t, faces, FaceBottom)
}

func TestFaceUV_FirstTile(t *testing.T) {
	u, v, w, h := FaceUV(Air, FaceNorth)
	assert.InDelta(t, 0.0, u, 1e-6)
	assert.InDelta(t, 0.0, v, 1e-6)
	assert.InDelta(t, 1.0/16.0, w, 1e-6)
	assert.InDelta(t, 1.0/16.0, h, 1e-6)
}

func TestFaceUV_SecondRow(t *testing.T) {
	u, v, w, h := FaceUV(GoldOre, FaceTop)
	assert.InDelta(t, 0.0, u, 1e-6)
	assert.InDelta(t, 1.0/16.0, v, 1e-6)
	assert.InDelta(t, 1.0/16.0, w, 1e-6)
	assert.InDelta(t, 1.0/16.0, h, 1e-6)
}

func TestFaceUV_MidAtlas(t *testing.T) {
	u, v, _, _ := FaceUV(Glass, FaceSouth)
	assert.InDelta(t, 10.0/16.0, u, 1e-6)
	assert.InDelta(t, 0.0, v, 1e-6)
}

func TestFaceUV_AllFacesSameTile(t *testing.T) {
	var firstU, firstV float32
	for i, face := range AllFaces() {
		u, v, _, _ := FaceUV(Stone, face)
		if i == 0 {
			firstU, firstV = u, v
		} else {
			assert.InDelta(t, firstU, u, 1e-6)
			assert.InDelta(t, firstV, v, 1e-6)
		}
	}
}

func TestFaceVertices_North(t *testing.T) {
	verts := FaceVertices(FaceNorth, 0, 0, 0)
	for _, v := range verts {
		assert.Equal(t, float32(0), v.Z, "north face Z must be 0")
	}
	assert.Equal(t, float32(0), verts[0].Y)
	assert.Equal(t, float32(0), verts[1].Y)
	assert.Equal(t, float32(1), verts[2].Y)
	assert.Equal(t, float32(1), verts[3].Y)
}

func TestFaceVertices_South(t *testing.T) {
	verts := FaceVertices(FaceSouth, 0, 0, 0)
	for _, v := range verts {
		assert.Equal(t, float32(1), v.Z, "south face Z must be 1")
	}
}

func TestFaceVertices_East(t *testing.T) {
	verts := FaceVertices(FaceEast, 0, 0, 0)
	for _, v := range verts {
		assert.Equal(t, float32(1), v.X, "east face X must be 1")
	}
}

func TestFaceVertices_West(t *testing.T) {
	verts := FaceVertices(FaceWest, 0, 0, 0)
	for _, v := range verts {
		assert.Equal(t, float32(0), v.X, "west face X must be 0")
	}
}

func TestFaceVertices_Top(t *testing.T) {
	verts := FaceVertices(FaceTop, 0, 0, 0)
	for _, v := range verts {
		assert.Equal(t, float32(1), v.Y, "top face Y must be 1")
	}
}

func TestFaceVertices_Bottom(t *testing.T) {
	verts := FaceVertices(FaceBottom, 0, 0, 0)
	for _, v := range verts {
		assert.Equal(t, float32(0), v.Y, "bottom face Y must be 0")
	}
}

func TestFaceVertices_Offset(t *testing.T) {
	verts := FaceVertices(FaceTop, 5, 10, 15)
	for _, v := range verts {
		assert.Equal(t, float32(11), v.Y, "top face Y at offset 10")
		assert.GreaterOrEqual(t, v.X, float32(5))
		assert.LessOrEqual(t, v.X, float32(6))
		assert.GreaterOrEqual(t, v.Z, float32(15))
		assert.LessOrEqual(t, v.Z, float32(16))
	}
}

func TestFaceVertices_DefaultFace(t *testing.T) {
	verts := FaceVertices(BlockFace(99), 0, 0, 0)
	for _, v := range verts {
		assert.Equal(t, mcmath.Vec3{}, v, "unknown face should return zero vertices")
	}
}

func TestFaceVertices_FourDistinctPoints(t *testing.T) {
	for _, face := range AllFaces() {
		verts := FaceVertices(face, 0, 0, 0)
		for i := 0; i < 4; i++ {
			for j := i + 1; j < 4; j++ {
				assert.NotEqual(t, verts[i], verts[j],
					"face %d: vertex %d and %d should differ", face, i, j)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// registry_init.go -- InitRegistry
// ---------------------------------------------------------------------------

func TestInitRegistry(t *testing.T) {
	original := Blocks
	Blocks = registry.New[BlockProperties]()
	defer func() { Blocks = original }()

	InitRegistry()

	assert.Equal(t, len(blockOrder), Blocks.Len())

	props, ok := Blocks.Get(Stone)
	assert.True(t, ok)
	assert.Equal(t, "stone", props.Name)

	props, ok = Blocks.Get(Torch)
	assert.True(t, ok)
	assert.Equal(t, "torch", props.Name)

	props, id, ok := Blocks.GetByName("diamond_ore")
	assert.True(t, ok)
	assert.Equal(t, DiamondOre, id)
	assert.Equal(t, "diamond_ore", props.Name)

	_, err := Blocks.Register("new_block", BlockProperties{Name: "new_block"})
	assert.Error(t, err)
}

func TestInitRegistry_IDsMatchConstants(t *testing.T) {
	original := Blocks
	Blocks = registry.New[BlockProperties]()
	defer func() { Blocks = original }()

	InitRegistry()

	tests := []struct {
		id   BlockID
		name string
	}{
		{Air, "air"},
		{Stone, "stone"},
		{Dirt, "dirt"},
		{Grass, "grass"},
		{Sandstone, "sandstone"},
		{Bedrock, "bedrock"},
		{CraftingTable, "crafting_table"},
	}
	for _, tc := range tests {
		name, ok := Blocks.Name(tc.id)
		assert.True(t, ok, "ID %d should exist", tc.id)
		assert.Equal(t, tc.name, name)
	}
}

func TestInitRegistry_PanicsOnDuplicate(t *testing.T) {
	original := Blocks
	Blocks = registry.New[BlockProperties]()
	defer func() { Blocks = original }()

	_, err := Blocks.Register("air", BlockProperties{Name: "air"})
	assert.NoError(t, err)

	assert.Panics(t, func() {
		InitRegistry()
	})
}

// ---------------------------------------------------------------------------
// Stair and slab block IDs and properties
// ---------------------------------------------------------------------------

func TestStairBlockIDConstants(t *testing.T) {
	assert.Equal(t, BlockID(81), OakStairs)
	assert.Equal(t, BlockID(82), CobblestoneStairs)
	assert.Equal(t, BlockID(83), StoneStairs)
	assert.Equal(t, BlockID(84), BirchStairs)
	assert.Equal(t, BlockID(85), SpruceStairs)
	assert.Equal(t, BlockID(86), SandstoneStairs)
}

func TestSlabBlockIDConstants(t *testing.T) {
	assert.Equal(t, BlockID(87), OakSlab)
	assert.Equal(t, BlockID(88), CobblestoneSlab)
	assert.Equal(t, BlockID(89), StoneSlab)
	assert.Equal(t, BlockID(90), BirchSlab)
	assert.Equal(t, BlockID(91), SpruceSlab)
	assert.Equal(t, BlockID(92), SandstoneSlab)
}

func TestStairProperties(t *testing.T) {
	stairs := []struct {
		id       BlockID
		name     string
		hardness float32
	}{
		{OakStairs, "oak_stairs", 2},
		{CobblestoneStairs, "cobblestone_stairs", 2},
		{StoneStairs, "stone_stairs", 1.5},
		{BirchStairs, "birch_stairs", 2},
		{SpruceStairs, "spruce_stairs", 2},
		{SandstoneStairs, "sandstone_stairs", 0.8},
	}
	for _, tc := range stairs {
		p := GetProperties(tc.id)
		assert.Equal(t, tc.name, p.Name)
		assert.True(t, p.Solid, "%s should be solid", tc.name)
		assert.True(t, p.IsStair, "%s should be a stair", tc.name)
		assert.False(t, p.IsSlab, "%s should not be a slab", tc.name)
		assert.Equal(t, tc.hardness, p.Hardness, "%s hardness", tc.name)
	}
}

func TestSlabProperties(t *testing.T) {
	slabs := []struct {
		id       BlockID
		name     string
		hardness float32
	}{
		{OakSlab, "oak_slab", 2},
		{CobblestoneSlab, "cobblestone_slab", 2},
		{StoneSlab, "stone_slab", 1.5},
		{BirchSlab, "birch_slab", 2},
		{SpruceSlab, "spruce_slab", 2},
		{SandstoneSlab, "sandstone_slab", 0.8},
	}
	for _, tc := range slabs {
		p := GetProperties(tc.id)
		assert.Equal(t, tc.name, p.Name)
		assert.True(t, p.Solid, "%s should be solid", tc.name)
		assert.True(t, p.IsSlab, "%s should be a slab", tc.name)
		assert.False(t, p.IsStair, "%s should not be a stair", tc.name)
		assert.Equal(t, tc.hardness, p.Hardness, "%s hardness", tc.name)
	}
}

func TestIsStairBlock(t *testing.T) {
	assert.True(t, IsStairBlock(OakStairs))
	assert.True(t, IsStairBlock(CobblestoneStairs))
	assert.False(t, IsStairBlock(Stone))
	assert.False(t, IsStairBlock(OakSlab))
}

func TestIsSlabBlock(t *testing.T) {
	assert.True(t, IsSlabBlock(OakSlab))
	assert.True(t, IsSlabBlock(CobblestoneSlab))
	assert.False(t, IsSlabBlock(Stone))
	assert.False(t, IsSlabBlock(OakStairs))
}

func TestGetBlockAABBs_FullBlock(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	aabbs := GetBlockAABBs(Stone, pos, OrientNorth)
	assert.Len(t, aabbs, 1)
	assert.Equal(t, mcmath.BlockAABB(pos), aabbs[0])
}

func TestGetBlockAABBs_Air(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	aabbs := GetBlockAABBs(Air, pos, OrientNorth)
	assert.Nil(t, aabbs)
}

func TestGetBlockAABBs_Stair(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	aabbs := GetBlockAABBs(OakStairs, pos, OrientNorth)
	assert.Len(t, aabbs, 2)
}

func TestGetBlockAABBs_SlabBottom(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	aabbs := GetBlockAABBs(OakSlab, pos, OrientNorth)
	assert.Len(t, aabbs, 1)
	assert.Equal(t, mcmath.SlabAABB(pos, false), aabbs[0])
}

func TestGetBlockAABBs_SlabTop(t *testing.T) {
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	aabbs := GetBlockAABBs(OakSlab, pos, OrientUp)
	assert.Len(t, aabbs, 1)
	assert.Equal(t, mcmath.SlabAABB(pos, true), aabbs[0])
}

func TestInitRegistry_StairsAndSlabs(t *testing.T) {
	original := Blocks
	Blocks = registry.New[BlockProperties]()
	defer func() { Blocks = original }()

	InitRegistry()

	tests := []struct {
		id   BlockID
		name string
	}{
		{OakStairs, "oak_stairs"},
		{CobblestoneStairs, "cobblestone_stairs"},
		{StoneStairs, "stone_stairs"},
		{BirchStairs, "birch_stairs"},
		{SpruceStairs, "spruce_stairs"},
		{SandstoneStairs, "sandstone_stairs"},
		{OakSlab, "oak_slab"},
		{CobblestoneSlab, "cobblestone_slab"},
		{StoneSlab, "stone_slab"},
		{BirchSlab, "birch_slab"},
		{SpruceSlab, "spruce_slab"},
		{SandstoneSlab, "sandstone_slab"},
	}
	for _, tc := range tests {
		name, ok := Blocks.Name(tc.id)
		assert.True(t, ok, "ID %d should exist", tc.id)
		assert.Equal(t, tc.name, name)
	}
}
