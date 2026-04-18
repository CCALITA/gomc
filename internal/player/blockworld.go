package player

import "github.com/fanxiyao/gomc/internal/mcmath"

// BlockWorld abstracts block-level access to the voxel world so that the
// player package does not depend on the concrete world.World type.
// The world.World struct satisfies this interface implicitly.
type BlockWorld interface {
	GetBlock(pos mcmath.BlockPos) uint16
	SetBlock(pos mcmath.BlockPos, id uint16)
	GetBlockAABBs(region mcmath.AABB) []mcmath.AABB
}
