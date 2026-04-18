package render

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/fanxiyao/gomc/internal/mcmath"
	vk "github.com/vulkan-go/vulkan"
)

// chunkMesh holds the GPU buffers and index count for a single chunk's mesh.
type chunkMesh struct {
	VertexBuffer *Buffer
	IndexBuffer  *Buffer
	IndexCount   uint32
}

// ViewProjectionUBO is the uniform buffer object layout sent to the vertex
// shader. It contains the combined view-projection matrix (16 floats).
type ViewProjectionUBO struct {
	ViewProjection [16]float32
}

// TODO(day-night): Add a TimeOfDayUBO struct and per-frame uniform buffers
// for the fragment shader's TimeOfDay uniform block (set=0, binding=1).
// The struct should match the GLSL layout in chunk.frag:
//
//   type TimeOfDayUBO struct {
//       SunDir       [3]float32
//       _pad0        float32
//       AmbientColor [3]float32
//       _pad1        float32
//       SkyColor     [3]float32
//       AmbientLevel float32
//   }
//
// Wire it up by:
// 1. Adding a second descriptor set layout binding in Pipeline.CreateGraphicsPipeline
//    for (set=0, binding=1, UniformBuffer, FragmentBit).
// 2. Creating per-frame TimeOfDay uniform buffers in NewChunkRenderer.
// 3. Updating the descriptor sets to include the TimeOfDay buffer.
// 4. In DrawAll, accept a *game.TimeKeeper parameter, populate TimeOfDayUBO
//    from its SunDirection/AmbientLevel/SkyColor methods, and upload it.
// See internal/game/timekeeper.go for the TimeKeeper API.

// ChunkPushConstants is the push constant block for per-chunk data.
// It carries the chunk's world-space origin.
type ChunkPushConstants struct {
	ChunkWorldX float32
	ChunkWorldY float32
	ChunkWorldZ float32
	_pad        float32 // align to 16 bytes
}

// ChunkRenderer manages uploading and drawing chunk meshes.
type ChunkRenderer struct {
	mu     sync.RWMutex
	meshes map[[2]int32]*chunkMesh

	uniformBuffers [maxFramesInFlight]*Buffer

	ctx     *VulkanContext
	cmdPool *CommandPool
	pipe    *Pipeline
}

// NewChunkRenderer creates a ChunkRenderer.
func NewChunkRenderer(ctx *VulkanContext, cmdPool *CommandPool, pipe *Pipeline) (*ChunkRenderer, error) {
	cr := &ChunkRenderer{
		meshes:  make(map[[2]int32]*chunkMesh),
		ctx:     ctx,
		cmdPool: cmdPool,
		pipe:    pipe,
	}

	// Create per-frame uniform buffers.
	for i := 0; i < maxFramesInFlight; i++ {
		ub, err := CreateUniformBuffer(ctx, int(unsafe.Sizeof(ViewProjectionUBO{})))
		if err != nil {
			cr.Cleanup()
			return nil, fmt.Errorf("failed to create uniform buffer %d: %w", i, err)
		}
		cr.uniformBuffers[i] = ub
	}

	return cr, nil
}

// UploadMesh uploads (or replaces) the mesh for the chunk at chunkPos.
func (cr *ChunkRenderer) UploadMesh(chunkPos mcmath.ChunkPos, vertices []float32, indices []uint32) error {
	key := chunkPos.Key()

	vb, err := CreateVertexBuffer(cr.ctx, cr.cmdPool, vertices)
	if err != nil {
		return fmt.Errorf("failed to create chunk vertex buffer: %w", err)
	}

	ib, err := CreateIndexBuffer(cr.ctx, cr.cmdPool, indices)
	if err != nil {
		vb.Cleanup()
		return fmt.Errorf("failed to create chunk index buffer: %w", err)
	}

	mesh := &chunkMesh{
		VertexBuffer: vb,
		IndexBuffer:  ib,
		IndexCount:   uint32(len(indices)),
	}

	cr.mu.Lock()
	old, exists := cr.meshes[key]
	cr.meshes[key] = mesh
	cr.mu.Unlock()

	if exists {
		old.VertexBuffer.Cleanup()
		old.IndexBuffer.Cleanup()
	}

	return nil
}

// RemoveMesh removes and cleans up the mesh for the chunk at chunkPos.
func (cr *ChunkRenderer) RemoveMesh(chunkPos mcmath.ChunkPos) {
	key := chunkPos.Key()

	cr.mu.Lock()
	mesh, exists := cr.meshes[key]
	if exists {
		delete(cr.meshes, key)
	}
	cr.mu.Unlock()

	if exists {
		mesh.VertexBuffer.Cleanup()
		mesh.IndexBuffer.Cleanup()
	}
}

// DrawAll records draw commands for all uploaded chunk meshes.
// It updates the view-projection uniform buffer and issues indexed draw
// calls with per-chunk push constants.
func (cr *ChunkRenderer) DrawAll(cmdBuf vk.CommandBuffer, camera *Camera, aspect float32, frameIndex uint32) {
	// Update the uniform buffer for this frame.
	vp := camera.ViewProjectionMatrix(aspect)
	ubo := ViewProjectionUBO{ViewProjection: vp}

	ub := cr.uniformBuffers[frameIndex]
	_ = ub.UpdateUniformBuffer(unsafe.Pointer(&ubo), vk.DeviceSize(unsafe.Sizeof(ubo)))

	vk.CmdBindPipeline(cmdBuf, vk.PipelineBindPointGraphics, cr.pipe.GraphicsPipeline)

	cr.mu.RLock()
	defer cr.mu.RUnlock()

	for key, mesh := range cr.meshes {
		if mesh.IndexCount == 0 {
			continue
		}

		offsets := []vk.DeviceSize{0}
		buffers := []vk.Buffer{mesh.VertexBuffer.Handle}
		vk.CmdBindVertexBuffers(cmdBuf, 0, 1, buffers, offsets)
		vk.CmdBindIndexBuffer(cmdBuf, mesh.IndexBuffer.Handle, 0, vk.IndexTypeUint32)

		pc := ChunkPushConstants{
			ChunkWorldX: float32(key[0]) * float32(mcmath.ChunkSize),
			ChunkWorldY: 0,
			ChunkWorldZ: float32(key[1]) * float32(mcmath.ChunkSize),
		}

		vk.CmdPushConstants(
			cmdBuf,
			cr.pipe.PipelineLayout,
			vk.ShaderStageFlags(vk.ShaderStageVertexBit),
			0,
			uint32(unsafe.Sizeof(pc)),
			unsafe.Pointer(&pc),
		)

		vk.CmdDrawIndexed(cmdBuf, mesh.IndexCount, 1, 0, 0, 0)
	}
}

// MeshCount returns the number of currently uploaded chunk meshes.
func (cr *ChunkRenderer) MeshCount() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return len(cr.meshes)
}

// Cleanup destroys all chunk meshes and uniform buffers.
func (cr *ChunkRenderer) Cleanup() {
	cr.mu.Lock()
	for _, mesh := range cr.meshes {
		mesh.VertexBuffer.Cleanup()
		mesh.IndexBuffer.Cleanup()
	}
	cr.meshes = nil
	cr.mu.Unlock()

	for i := range cr.uniformBuffers {
		if cr.uniformBuffers[i] != nil {
			cr.uniformBuffers[i].Cleanup()
			cr.uniformBuffers[i] = nil
		}
	}
}
