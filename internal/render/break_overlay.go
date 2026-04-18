package render

import (
	"sync"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// overlayOffset is the distance each face is pushed outward from the block face
// to prevent z-fighting with the underlying block.
const overlayOffset = 0.001

// BreakOverlay tracks the block-breaking animation state and generates overlay
// geometry for a single targeted block.
type BreakOverlay struct {
	mu        sync.RWMutex
	targetPos *mcmath.BlockPos
	progress  float32
}

// SetTarget sets the block to show the breaking animation on and the current
// break progress (0.0 = just started, 1.0 = fully broken).
func (bo *BreakOverlay) SetTarget(pos mcmath.BlockPos, progress float32) {
	bo.mu.Lock()
	bo.targetPos = &pos
	bo.progress = progress
	bo.mu.Unlock()
}

// ClearTarget removes the breaking overlay.
func (bo *BreakOverlay) ClearTarget() {
	bo.mu.Lock()
	bo.targetPos = nil
	bo.progress = 0
	bo.mu.Unlock()
}

// HasTarget reports whether a block is currently targeted for breaking.
func (bo *BreakOverlay) HasTarget() bool {
	bo.mu.RLock()
	defer bo.mu.RUnlock()
	return bo.targetPos != nil
}

// GetStage returns the animation stage (0-9) derived from the current progress.
// The stage is floor(progress * 10), clamped to [0, 9].
func (bo *BreakOverlay) GetStage() int {
	bo.mu.RLock()
	defer bo.mu.RUnlock()
	return progressToStage(bo.progress)
}

// progressToStage converts a progress value to a stage integer in [0, 9].
func progressToStage(progress float32) int {
	stage := int(progress * 10)
	if stage < 0 {
		return 0
	}
	if stage > 9 {
		return 9
	}
	return stage
}

// GenerateOverlayVertices produces the vertex and index data for a breaking
// overlay around the block at pos for the given animation stage.
//
// Each vertex contains 4 floats: position (x, y, z) and alpha.
// The overlay has 6 faces (one per block face), each a quad of 4 vertices,
// giving 24 vertices total. Indices reference these vertices with 6 indices
// per face (two triangles), giving 36 indices total.
func GenerateOverlayVertices(pos mcmath.BlockPos, stage int) (vertices []float32, indices []uint32) {
	alpha := float32(stage+1) * 0.08

	x := float32(pos.X)
	y := float32(pos.Y)
	z := float32(pos.Z)

	d := float32(overlayOffset)

	type face struct {
		v [4][3]float32
	}

	faces := [6]face{
		// North face (-Z)
		{v: [4][3]float32{
			{x - d, y - d, z - d},
			{x + 1 + d, y - d, z - d},
			{x + 1 + d, y + 1 + d, z - d},
			{x - d, y + 1 + d, z - d},
		}},
		// South face (+Z)
		{v: [4][3]float32{
			{x + 1 + d, y - d, z + 1 + d},
			{x - d, y - d, z + 1 + d},
			{x - d, y + 1 + d, z + 1 + d},
			{x + 1 + d, y + 1 + d, z + 1 + d},
		}},
		// East face (+X)
		{v: [4][3]float32{
			{x + 1 + d, y - d, z - d},
			{x + 1 + d, y - d, z + 1 + d},
			{x + 1 + d, y + 1 + d, z + 1 + d},
			{x + 1 + d, y + 1 + d, z - d},
		}},
		// West face (-X)
		{v: [4][3]float32{
			{x - d, y - d, z + 1 + d},
			{x - d, y - d, z - d},
			{x - d, y + 1 + d, z - d},
			{x - d, y + 1 + d, z + 1 + d},
		}},
		// Up face (+Y)
		{v: [4][3]float32{
			{x - d, y + 1 + d, z - d},
			{x + 1 + d, y + 1 + d, z - d},
			{x + 1 + d, y + 1 + d, z + 1 + d},
			{x - d, y + 1 + d, z + 1 + d},
		}},
		// Down face (-Y)
		{v: [4][3]float32{
			{x - d, y - d, z + 1 + d},
			{x + 1 + d, y - d, z + 1 + d},
			{x + 1 + d, y - d, z - d},
			{x - d, y - d, z - d},
		}},
	}

	vertices = make([]float32, 0, 24*4)
	indices = make([]uint32, 0, 36)

	for i, f := range faces {
		base := uint32(i * 4)
		for _, v := range f.v {
			vertices = append(vertices, v[0], v[1], v[2], alpha)
		}
		// Two triangles per quad: (0,1,2) and (0,2,3).
		indices = append(indices,
			base+0, base+1, base+2,
			base+0, base+2, base+3,
		)
	}

	return vertices, indices
}
