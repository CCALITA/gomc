package render

import (
	"fmt"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// Buffer wraps a Vulkan buffer with its associated device memory.
type Buffer struct {
	Handle vk.Buffer
	Memory vk.DeviceMemory
	Size   vk.DeviceSize

	ctx *VulkanContext
}

// CreateBuffer creates a Vulkan buffer with the given size, usage, and
// memory property flags.
func CreateBuffer(
	ctx *VulkanContext,
	size vk.DeviceSize,
	usage vk.BufferUsageFlags,
	properties vk.MemoryPropertyFlags,
) (*Buffer, error) {
	bufferInfo := &vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        size,
		Usage:       usage,
		SharingMode: vk.SharingModeExclusive,
	}

	var buffer vk.Buffer
	if res := vk.CreateBuffer(ctx.Device, bufferInfo, nil, &buffer); res != vk.Success {
		return nil, fmt.Errorf("create buffer: vulkan result %d", res)
	}

	var memReqs vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(ctx.Device, buffer, &memReqs)
	memReqs.Deref()

	memTypeIndex, err := findMemoryType(ctx.PhysicalDevice, memReqs.MemoryTypeBits, properties)
	if err != nil {
		vk.DestroyBuffer(ctx.Device, buffer, nil)
		return nil, fmt.Errorf("find buffer memory type: %w", err)
	}

	allocInfo := &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var memory vk.DeviceMemory
	if res := vk.AllocateMemory(ctx.Device, allocInfo, nil, &memory); res != vk.Success {
		vk.DestroyBuffer(ctx.Device, buffer, nil)
		return nil, fmt.Errorf("allocate buffer memory: vulkan result %d", res)
	}

	if res := vk.BindBufferMemory(ctx.Device, buffer, memory, 0); res != vk.Success {
		vk.FreeMemory(ctx.Device, memory, nil)
		vk.DestroyBuffer(ctx.Device, buffer, nil)
		return nil, fmt.Errorf("bind buffer memory: vulkan result %d", res)
	}

	return &Buffer{
		Handle: buffer,
		Memory: memory,
		Size:   size,
		ctx:    ctx,
	}, nil
}

// Cleanup destroys the buffer and frees its memory.
func (b *Buffer) Cleanup() {
	if b.Handle != nil {
		vk.DestroyBuffer(b.ctx.Device, b.Handle, nil)
	}
	if b.Memory != nil {
		vk.FreeMemory(b.ctx.Device, b.Memory, nil)
	}
}

// CreateVertexBuffer creates a device-local vertex buffer by staging
// the given float32 data through a host-visible staging buffer.
func CreateVertexBuffer(ctx *VulkanContext, cmdPool *CommandPool, data []float32) (*Buffer, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("vertex data is empty")
	}

	bufferSize := vk.DeviceSize(len(data) * 4)

	staging, err := CreateBuffer(
		ctx,
		bufferSize,
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
	)
	if err != nil {
		return nil, fmt.Errorf("create staging buffer: %w", err)
	}
	defer staging.Cleanup()

	if err := mapAndCopy(ctx.Device, staging.Memory, bufferSize, unsafe.Pointer(&data[0])); err != nil {
		return nil, fmt.Errorf("map staging buffer: %w", err)
	}

	vertexBuffer, err := CreateBuffer(
		ctx,
		bufferSize,
		vk.BufferUsageFlags(vk.BufferUsageTransferDstBit|vk.BufferUsageVertexBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
	)
	if err != nil {
		return nil, fmt.Errorf("create vertex buffer: %w", err)
	}

	if err := cmdPool.CopyBuffer(staging.Handle, vertexBuffer.Handle, bufferSize); err != nil {
		vertexBuffer.Cleanup()
		return nil, fmt.Errorf("copy to vertex buffer: %w", err)
	}

	return vertexBuffer, nil
}

// CreateIndexBuffer creates a device-local index buffer by staging
// the given uint32 data through a host-visible staging buffer.
func CreateIndexBuffer(ctx *VulkanContext, cmdPool *CommandPool, data []uint32) (*Buffer, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("index data is empty")
	}

	bufferSize := vk.DeviceSize(len(data) * 4)

	staging, err := CreateBuffer(
		ctx,
		bufferSize,
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
	)
	if err != nil {
		return nil, fmt.Errorf("create index staging buffer: %w", err)
	}
	defer staging.Cleanup()

	if err := mapAndCopy(ctx.Device, staging.Memory, bufferSize, unsafe.Pointer(&data[0])); err != nil {
		return nil, fmt.Errorf("map index staging buffer: %w", err)
	}

	indexBuffer, err := CreateBuffer(
		ctx,
		bufferSize,
		vk.BufferUsageFlags(vk.BufferUsageTransferDstBit|vk.BufferUsageIndexBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
	)
	if err != nil {
		return nil, fmt.Errorf("create index buffer: %w", err)
	}

	if err := cmdPool.CopyBuffer(staging.Handle, indexBuffer.Handle, bufferSize); err != nil {
		indexBuffer.Cleanup()
		return nil, fmt.Errorf("copy to index buffer: %w", err)
	}

	return indexBuffer, nil
}

// CreateUniformBuffer creates a host-visible, host-coherent buffer
// suitable for uniform data that is updated each frame.
func CreateUniformBuffer(ctx *VulkanContext, size int) (*Buffer, error) {
	return CreateBuffer(
		ctx,
		vk.DeviceSize(size),
		vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
	)
}

// UpdateUniformBuffer maps the buffer memory and copies the provided data.
func (b *Buffer) UpdateUniformBuffer(data unsafe.Pointer, size vk.DeviceSize) error {
	return mapAndCopy(b.ctx.Device, b.Memory, size, data)
}

// mapAndCopy maps device memory, copies data into it, and unmaps.
func mapAndCopy(device vk.Device, memory vk.DeviceMemory, size vk.DeviceSize, src unsafe.Pointer) error {
	var mapped unsafe.Pointer
	if res := vk.MapMemory(device, memory, 0, size, 0, &mapped); res != vk.Success {
		return fmt.Errorf("map memory: vulkan result %d", res)
	}

	// Convert source pointer to a byte slice for vk.Memcopy.
	srcBytes := unsafe.Slice((*byte)(src), int(size))
	vk.Memcopy(mapped, srcBytes)
	vk.UnmapMemory(device, memory)

	return nil
}
