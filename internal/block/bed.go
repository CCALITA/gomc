package block

import "github.com/fanxiyao/gomc/internal/mcmath"

// Bed state encoding uses the upper bits of the block ID (above the base
// 8-bit ID). The lower 8 bits are the Bed block constant (96).
//
// State bits (encoded in the upper byte of uint16):
//   bit 8  (0x0100): IsHead — 1 for the head half, 0 for the foot half.
//   bits 9-10 (0x0600): Orientation — 0=North, 1=South, 2=East, 3=West.
//
// The facing direction indicates where the head is relative to the foot.

const (
	bedHeadBit       = 8
	bedOrientShift   = 9
	bedOrientMask    = 0x03
	bedStateMask     = 0xFF00
	bedPartialHeight = 0.5625 // 9/16 of a full block
)

// BedOrientation represents the horizontal facing of a bed.
type BedOrientation uint8

const (
	BedNorth BedOrientation = 0
	BedSouth BedOrientation = 1
	BedEast  BedOrientation = 2
	BedWest  BedOrientation = 3
)

// IsBed reports whether the given block ID is any bed variant
// (head or foot, any orientation).
func IsBed(id uint16) bool {
	return BaseID(id) == Bed
}

// EncodeBedFoot returns a bed block ID for the foot half with the given orientation.
func EncodeBedFoot(orient BedOrientation) uint16 {
	return Bed | (uint16(orient&bedOrientMask) << bedOrientShift)
}

// EncodeBedHead returns a bed block ID for the head half with the given orientation.
func EncodeBedHead(orient BedOrientation) uint16 {
	return Bed | (1 << bedHeadBit) | (uint16(orient&bedOrientMask) << bedOrientShift)
}

// BedIsHead reports whether the bed block ID represents the head half.
func BedIsHead(id uint16) bool {
	return (id>>bedHeadBit)&1 == 1
}

// BedOrientation extracts the bed orientation from a bed block ID.
func BedOrient(id uint16) BedOrientation {
	return BedOrientation((id >> bedOrientShift) & bedOrientMask)
}

// BedOtherHalf returns the position of the other half of the bed.
// If the given block is the foot, it returns the head position, and vice versa.
func BedOtherHalf(pos mcmath.BlockPos, id uint16) mcmath.BlockPos {
	orient := BedOrient(id)
	dx, dz := bedOrientOffset(orient)
	if BedIsHead(id) {
		// Head -> foot: go opposite direction.
		return pos.Offset(-dx, 0, -dz)
	}
	// Foot -> head: go in facing direction.
	return pos.Offset(dx, 0, dz)
}

// bedOrientOffset returns the (dx, dz) offset from foot to head for the
// given bed orientation.
func bedOrientOffset(orient BedOrientation) (int32, int32) {
	switch orient {
	case BedNorth:
		return 0, -1
	case BedSouth:
		return 0, 1
	case BedEast:
		return 1, 0
	case BedWest:
		return -1, 0
	default:
		return 0, -1
	}
}

// BedAABB returns the AABB for a bed block at the given position.
// The bed is 9/16 of a block tall.
func BedAABB(pos mcmath.BlockPos) mcmath.AABB {
	return mcmath.AABB{
		Min: pos.ToVec3(),
		Max: mcmath.Vec3{
			X: float32(pos.X) + 1,
			Y: float32(pos.Y) + bedPartialHeight,
			Z: float32(pos.Z) + 1,
		},
	}
}

// FacingToBedOrientation converts a horizontal camera forward vector to a
// BedOrientation by selecting the dominant horizontal axis.
func FacingToBedOrientation(forward mcmath.Vec3) BedOrientation {
	absX := forward.X
	absZ := forward.Z
	if absX < 0 {
		absX = -absX
	}
	if absZ < 0 {
		absZ = -absZ
	}

	if absX > absZ {
		if forward.X > 0 {
			return BedEast
		}
		return BedWest
	}
	if forward.Z > 0 {
		return BedSouth
	}
	return BedNorth
}
