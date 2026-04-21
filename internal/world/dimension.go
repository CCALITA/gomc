package world

// DimensionID identifies a world dimension.
type DimensionID int

// Dimension constants.
const (
	Overworld DimensionID = 0
	Nether    DimensionID = 1
)

// NetherOverworldScale is the coordinate scaling factor between Nether and
// Overworld (1 block in Nether = 8 blocks in Overworld).
const NetherOverworldScale = 8

// NewNetherWorld creates a World configured for Nether terrain generation.
func NewNetherWorld(seed int64) *World {
	w := NewWorld(seed)
	w.Dimension = Nether
	w.netherGen = NewNetherTerrainGenerator(seed)
	return w
}
