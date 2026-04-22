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
// shader. It contains separate view and projection matrices.
type ViewProjectionUBO struct {
	View       [16]float32
	Projection [16]float32
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
// It carries the model matrix (translation to chunk world position).
type ChunkPushConstants struct {
	Model [16]float32
}

func chunkModelMatrix(worldX, worldY, worldZ float32) [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		worldX, worldY, worldZ, 1,
	}
}

// ChunkRenderer manages uploading and drawing chunk meshes.
type ChunkRenderer struct {
	mu     sync.RWMutex
	meshes map[[2]int32]*chunkMesh

	uniformBuffers  [maxFramesInFlight]*Buffer
	descriptorPool  vk.DescriptorPool
	descriptorSets  []vk.DescriptorSet

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

	// Create descriptor pool.
	poolSizes := []vk.DescriptorPoolSize{
		{Type: vk.DescriptorTypeUniformBuffer, DescriptorCount: maxFramesInFlight},
		{Type: vk.DescriptorTypeCombinedImageSampler, DescriptorCount: maxFramesInFlight},
	}
	poolInfo := &vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		PoolSizeCount: uint32(len(poolSizes)),
		PPoolSizes:    poolSizes,
		MaxSets:       maxFramesInFlight,
	}
	if res := vk.CreateDescriptorPool(ctx.Device, poolInfo, nil, &cr.descriptorPool); res != vk.Success {
		cr.Cleanup()
		return nil, fmt.Errorf("create descriptor pool: vulkan result %d", res)
	}

	// Allocate descriptor sets.
	layouts := make([]vk.DescriptorSetLayout, maxFramesInFlight)
	for i := range layouts {
		layouts[i] = pipe.DescriptorLayout
	}
	allocInfo := &vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     cr.descriptorPool,
		DescriptorSetCount: maxFramesInFlight,
		PSetLayouts:        layouts,
	}
	cr.descriptorSets = make([]vk.DescriptorSet, maxFramesInFlight)
	if res := vk.AllocateDescriptorSets(ctx.Device, allocInfo, &cr.descriptorSets[0]); res != vk.Success {
		cr.Cleanup()
		return nil, fmt.Errorf("allocate descriptor sets: vulkan result %d", res)
	}

	// Update descriptor sets to point to uniform buffers.
	for i := 0; i < maxFramesInFlight; i++ {
		bufferInfo := vk.DescriptorBufferInfo{
			Buffer: cr.uniformBuffers[i].Handle,
			Offset: 0,
			Range:  vk.DeviceSize(unsafe.Sizeof(ViewProjectionUBO{})),
		}
		write := vk.WriteDescriptorSet{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          cr.descriptorSets[i],
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorType:  vk.DescriptorTypeUniformBuffer,
			DescriptorCount: 1,
			PBufferInfo:     []vk.DescriptorBufferInfo{bufferInfo},
		}
		vk.UpdateDescriptorSets(ctx.Device, 1, []vk.WriteDescriptorSet{write}, 0, nil)
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
// calls with per-chunk push constants. Chunks whose AABBs fall entirely
// outside the camera frustum are skipped (frustum culling).
func (cr *ChunkRenderer) DrawAll(cmdBuf vk.CommandBuffer, camera *Camera, aspect float32, frameIndex uint32) {
	// Update the uniform buffer for this frame.
	view := camera.ViewMatrix()
	proj := camera.ProjectionMatrix(aspect)
	ubo := ViewProjectionUBO{
		View:       view,
		Projection: proj,
	}

	ub := cr.uniformBuffers[frameIndex]
	_ = ub.UpdateUniformBuffer(unsafe.Pointer(&ubo), vk.DeviceSize(unsafe.Sizeof(ubo)))

	vk.CmdBindPipeline(cmdBuf, vk.PipelineBindPointGraphics, cr.pipe.GraphicsPipeline)
	vk.CmdBindDescriptorSets(cmdBuf, vk.PipelineBindPointGraphics, cr.pipe.PipelineLayout, 0, 1, []vk.DescriptorSet{cr.descriptorSets[frameIndex]}, 0, nil)

	frustum := camera.Frustum()

	cr.mu.RLock()
	defer cr.mu.RUnlock()

	for key, mesh := range cr.meshes {
		if mesh.IndexCount == 0 {
			continue
		}

		// Frustum culling: compute the chunk AABB and skip if outside.
		chunkAABB := chunkAABBFromKey(key)
		if !frustum.IntersectsAABB(chunkAABB) {
			continue
		}

		offsets := []vk.DeviceSize{0}
		buffers := []vk.Buffer{mesh.VertexBuffer.Handle}
		vk.CmdBindVertexBuffers(cmdBuf, 0, 1, buffers, offsets)
		vk.CmdBindIndexBuffer(cmdBuf, mesh.IndexBuffer.Handle, 0, vk.IndexTypeUint32)

		pc := ChunkPushConstants{
			Model: chunkModelMatrix(
				float32(key[0])*float32(mcmath.ChunkSize),
				0,
				float32(key[1])*float32(mcmath.ChunkSize),
			),
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

// chunkAABBFromKey computes the world-space axis-aligned bounding box for a
// chunk identified by its [chunkX, chunkZ] map key.
func chunkAABBFromKey(key [2]int32) mcmath.AABB {
	minX := float32(key[0]) * float32(mcmath.ChunkSize)
	minZ := float32(key[1]) * float32(mcmath.ChunkSize)
	return mcmath.AABB{
		Min: mcmath.Vec3{X: minX, Y: 0, Z: minZ},
		Max: mcmath.Vec3{
			X: minX + float32(mcmath.ChunkSize),
			Y: float32(mcmath.ChunkHeight),
			Z: minZ + float32(mcmath.ChunkSize),
		},
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
