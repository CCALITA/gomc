package block

// BlockState packs a block ID and state data into a single uint32.
// Layout (little-endian bit fields):
//
//	bits  0-15: BlockID (lower 16 bits)
//	bits 16-18: Orientation (3 bits, 0-5 for 6 directions)
//	bit  19:    Waterlogged (1 bit)
//	bit  20:    Powered (1 bit)
//	bits 21-31: Reserved (zero)
type BlockState uint32

const (
	stateIDMask          = 0x0000_FFFF
	stateOrientationMask = 0x0007_0000
	stateOrientationBits = 16
	stateWaterloggedBit  = 19
	statePoweredBit      = 20
)

// Orientation represents the facing direction of a block, corresponding to
// the six cardinal directions (0=North, 1=South, 2=East, 3=West, 4=Up, 5=Down).
type Orientation uint8

const (
	OrientNorth Orientation = 0
	OrientSouth Orientation = 1
	OrientEast  Orientation = 2
	OrientWest  Orientation = 3
	OrientUp    Orientation = 4
	OrientDown  Orientation = 5
)

// MaxOrientation is the highest valid orientation value.
const MaxOrientation Orientation = OrientDown

// NewBlockState creates a BlockState with the given block ID, orientation,
// and waterlogged flag. Orientation is clamped to 0-5.
func NewBlockState(id BlockID, orientation Orientation, waterlogged bool) BlockState {
	s := BlockState(id) & stateIDMask

	o := orientation
	if o > MaxOrientation {
		o = OrientNorth
	}
	s |= BlockState(o) << stateOrientationBits

	if waterlogged {
		s |= 1 << stateWaterloggedBit
	}

	return s
}

// StateBlockID extracts the block ID from a BlockState.
func (s BlockState) StateBlockID() BlockID {
	return BlockID(s & stateIDMask)
}

// StateOrientation extracts the orientation from a BlockState.
func (s BlockState) StateOrientation() Orientation {
	return Orientation((s & stateOrientationMask) >> stateOrientationBits)
}

// StateWaterlogged reports whether the block is waterlogged.
func (s BlockState) StateWaterlogged() bool {
	return (s>>stateWaterloggedBit)&1 == 1
}

// StatePowered reports whether the block is powered.
func (s BlockState) StatePowered() bool {
	return (s>>statePoweredBit)&1 == 1
}

// WithPowered returns a new BlockState with the powered bit set to the given value.
func (s BlockState) WithPowered(powered bool) BlockState {
	cleared := s &^ (1 << statePoweredBit)
	if powered {
		cleared |= 1 << statePoweredBit
	}
	return cleared
}

// WithOrientation returns a new BlockState with the orientation set to the given value.
// Orientation is clamped to 0-5.
func (s BlockState) WithOrientation(o Orientation) BlockState {
	if o > MaxOrientation {
		o = OrientNorth
	}
	cleared := s &^ stateOrientationMask
	return cleared | BlockState(o)<<stateOrientationBits
}

// WithWaterlogged returns a new BlockState with the waterlogged bit set to the given value.
func (s BlockState) WithWaterlogged(waterlogged bool) BlockState {
	cleared := s &^ (1 << stateWaterloggedBit)
	if waterlogged {
		cleared |= 1 << stateWaterloggedBit
	}
	return cleared
}
