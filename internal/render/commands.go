package render

import (
	"fmt"
	"runtime"

	vk "github.com/vulkan-go/vulkan"
)

// CommandPool manages a Vulkan command pool and per-frame command buffers,
// along with the synchronisation primitives needed for frame rendering.
type CommandPool struct {
	Pool           vk.CommandPool
	CommandBuffers []vk.CommandBuffer

	// Per-frame synchronisation objects.
	ImageAvailable []vk.Semaphore
	RenderFinished []vk.Semaphore
	InFlight       []vk.Fence

	CurrentFrame uint32

	ctx *VulkanContext
}

// NewCommandPool creates a command pool on the graphics queue family.
func NewCommandPool(ctx *VulkanContext) (*CommandPool, error) {
	poolInfo := &vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: ctx.QueueIndices.GraphicsFamily,
	}

	var pool vk.CommandPool
	if res := vk.CreateCommandPool(ctx.Device, poolInfo, nil, &pool); res != vk.Success {
		return nil, fmt.Errorf("create command pool: vulkan result %d", res)
	}

	cp := &CommandPool{
		Pool: pool,
		ctx:  ctx,
	}

	if err := cp.allocateCommandBuffers(); err != nil {
		cp.Cleanup()
		return nil, err
	}

	if err := cp.createSyncObjects(); err != nil {
		cp.Cleanup()
		return nil, err
	}

	return cp, nil
}

// allocateCommandBuffers allocates one command buffer per frame in flight.
func (cp *CommandPool) allocateCommandBuffers() error {
	cp.CommandBuffers = make([]vk.CommandBuffer, maxFramesInFlight)

	allocInfo := &vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        cp.Pool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: maxFramesInFlight,
	}

	if res := vk.AllocateCommandBuffers(cp.ctx.Device, allocInfo, cp.CommandBuffers); res != vk.Success {
		return fmt.Errorf("allocate command buffers: vulkan result %d", res)
	}
	return nil
}

// createSyncObjects creates the semaphores and fences for each frame in flight.
func (cp *CommandPool) createSyncObjects() error {
	cp.ImageAvailable = make([]vk.Semaphore, maxFramesInFlight)
	cp.RenderFinished = make([]vk.Semaphore, maxFramesInFlight)
	cp.InFlight = make([]vk.Fence, maxFramesInFlight)

	semaphoreInfo := &vk.SemaphoreCreateInfo{
		SType: vk.StructureTypeSemaphoreCreateInfo,
	}
	fenceInfo := &vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}

	for i := 0; i < maxFramesInFlight; i++ {
		if res := vk.CreateSemaphore(cp.ctx.Device, semaphoreInfo, nil, &cp.ImageAvailable[i]); res != vk.Success {
			return fmt.Errorf("create image available semaphore %d: vulkan result %d", i, res)
		}
		if res := vk.CreateSemaphore(cp.ctx.Device, semaphoreInfo, nil, &cp.RenderFinished[i]); res != vk.Success {
			return fmt.Errorf("create render finished semaphore %d: vulkan result %d", i, res)
		}
		if res := vk.CreateFence(cp.ctx.Device, fenceInfo, nil, &cp.InFlight[i]); res != vk.Success {
			return fmt.Errorf("create in-flight fence %d: vulkan result %d", i, res)
		}
	}

	return nil
}

// BeginFrame waits for the current frame fence, acquires the next swapchain
// image, resets and begins the command buffer, and returns the image index.
func (cp *CommandPool) BeginFrame(swapchain *Swapchain) (uint32, vk.CommandBuffer, error) {
	frame := cp.CurrentFrame

	fences := []vk.Fence{cp.InFlight[frame]}
	if res := vk.WaitForFences(cp.ctx.Device, 1, fences, vk.True, vk.MaxUint64); res != vk.Success {
		return 0, nil, fmt.Errorf("wait for fence: vulkan result %d", res)
	}

	var imageIndex uint32
	res := vk.AcquireNextImage(
		cp.ctx.Device, swapchain.Handle, vk.MaxUint64,
		cp.ImageAvailable[frame], nil, &imageIndex,
	)
	if res == vk.ErrorOutOfDate {
		return 0, nil, fmt.Errorf("swapchain out of date")
	}
	if res != vk.Success && res != vk.Suboptimal {
		return 0, nil, fmt.Errorf("acquire swapchain image: vulkan result %d", res)
	}

	if res := vk.ResetFences(cp.ctx.Device, 1, fences); res != vk.Success {
		return 0, nil, fmt.Errorf("reset fence: vulkan result %d", res)
	}

	cmdBuf := cp.CommandBuffers[frame]
	if res := vk.ResetCommandBuffer(cmdBuf, 0); res != vk.Success {
		return 0, nil, fmt.Errorf("reset command buffer: vulkan result %d", res)
	}

	beginInfo := &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
	}
	if res := vk.BeginCommandBuffer(cmdBuf, beginInfo); res != vk.Success {
		return 0, nil, fmt.Errorf("begin command buffer: vulkan result %d", res)
	}

	return imageIndex, cmdBuf, nil
}

// BeginRenderPass begins the render pass with the given framebuffer and extent.
func BeginRenderPass(cmdBuf vk.CommandBuffer, renderPass vk.RenderPass, framebuffer vk.Framebuffer, extent vk.Extent2D) {
	clearValues := []vk.ClearValue{
		vk.NewClearValue([]float32{0.53, 0.81, 0.92, 1.0}), // sky blue
		vk.NewClearDepthStencil(1.0, 0),
	}

	renderPassInfo := &vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  renderPass,
		Framebuffer: framebuffer,
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{X: 0, Y: 0},
			Extent: extent,
		},
		ClearValueCount: uint32(len(clearValues)),
		PClearValues:    clearValues,
	}

	vk.CmdBeginRenderPass(cmdBuf, renderPassInfo, vk.SubpassContentsInline)
}

// EndFrame ends the command buffer, submits it, and presents the image.
func (cp *CommandPool) EndFrame(swapchain *Swapchain, imageIndex uint32) error {
	frame := cp.CurrentFrame
	cmdBuf := cp.CommandBuffers[frame]

	vk.CmdEndRenderPass(cmdBuf)

	if res := vk.EndCommandBuffer(cmdBuf); res != vk.Success {
		return fmt.Errorf("end command buffer: vulkan result %d", res)
	}

	waitSemaphores := []vk.Semaphore{cp.ImageAvailable[frame]}
	waitStages := []vk.PipelineStageFlags{
		vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
	}
	signalSemaphores := []vk.Semaphore{cp.RenderFinished[frame]}

	submitInfo := &vk.SubmitInfo{
		SType:                vk.StructureTypeSubmitInfo,
		WaitSemaphoreCount:   1,
		PWaitSemaphores:      waitSemaphores,
		PWaitDstStageMask:    waitStages,
		CommandBufferCount:   1,
		PCommandBuffers:      []vk.CommandBuffer{cmdBuf},
		SignalSemaphoreCount: 1,
		PSignalSemaphores:    signalSemaphores,
	}

	if res := vk.QueueSubmit(cp.ctx.GraphicsQueue, 1, []vk.SubmitInfo{*submitInfo}, cp.InFlight[frame]); res != vk.Success {
		return fmt.Errorf("submit draw command buffer: vulkan result %d", res)
	}

	swapchains := []vk.Swapchain{swapchain.Handle}
	imageIndices := []uint32{imageIndex}

	presentInfo := &vk.PresentInfo{
		SType:              vk.StructureTypePresentInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    signalSemaphores,
		SwapchainCount:     1,
		PSwapchains:        swapchains,
		PImageIndices:      imageIndices,
	}

	res := vk.QueuePresent(cp.ctx.PresentQueue, presentInfo)
	if res == vk.ErrorOutOfDate || res == vk.Suboptimal {
		return fmt.Errorf("swapchain out of date")
	}
	if res != vk.Success {
		return fmt.Errorf("present swapchain image: vulkan result %d", res)
	}

	cp.CurrentFrame = (cp.CurrentFrame + 1) % maxFramesInFlight

	return nil
}

// CopyBuffer copies data from src to dst using a one-time command buffer.
func (cp *CommandPool) CopyBuffer(src, dst vk.Buffer, size vk.DeviceSize) error {
	cmdBuf, err := cp.beginSingleTimeCommands()
	if err != nil {
		return err
	}

	copyRegion := vk.BufferCopy{
		SrcOffset: 0,
		DstOffset: 0,
		Size:      size,
	}
	vk.CmdCopyBuffer(cmdBuf, src, dst, 1, []vk.BufferCopy{copyRegion})

	return cp.endSingleTimeCommands(cmdBuf)
}

// CopyBufferToImage copies buffer data into a VkImage.
func (cp *CommandPool) CopyBufferToImage(buffer vk.Buffer, image vk.Image, width, height uint32) error {
	cmdBuf, err := cp.beginSingleTimeCommands()
	if err != nil {
		return err
	}

	region := vk.BufferImageCopy{
		BufferOffset:      0,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
		ImageExtent: vk.Extent3D{Width: width, Height: height, Depth: 1},
	}

	vk.CmdCopyBufferToImage(cmdBuf, buffer, image, vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{region})

	return cp.endSingleTimeCommands(cmdBuf)
}

// TransitionImageLayout transitions an image between layout states.
func (cp *CommandPool) TransitionImageLayout(image vk.Image, oldLayout, newLayout vk.ImageLayout) error {
	cmdBuf, err := cp.beginSingleTimeCommands()
	if err != nil {
		return err
	}

	barrier := vk.ImageMemoryBarrier{
		SType:               vk.StructureTypeImageMemoryBarrier,
		OldLayout:           oldLayout,
		NewLayout:           newLayout,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               image,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}

	var srcStage, dstStage vk.PipelineStageFlags

	if oldLayout == vk.ImageLayoutUndefined && newLayout == vk.ImageLayoutTransferDstOptimal {
		barrier.SrcAccessMask = 0
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		srcStage = vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
	} else if oldLayout == vk.ImageLayoutTransferDstOptimal && newLayout == vk.ImageLayoutShaderReadOnlyOptimal {
		barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
		srcStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit)
	} else {
		return fmt.Errorf("unsupported layout transition: %d -> %d", oldLayout, newLayout)
	}

	vk.CmdPipelineBarrier(
		cmdBuf,
		srcStage, dstStage,
		0,
		0, nil,
		0, nil,
		1, []vk.ImageMemoryBarrier{barrier},
	)

	return cp.endSingleTimeCommands(cmdBuf)
}

// beginSingleTimeCommands allocates and begins a one-shot command buffer.
func (cp *CommandPool) beginSingleTimeCommands() (vk.CommandBuffer, error) {
	allocInfo := &vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		Level:              vk.CommandBufferLevelPrimary,
		CommandPool:        cp.Pool,
		CommandBufferCount: 1,
	}

	cmdBuffers := make([]vk.CommandBuffer, 1)
	if res := vk.AllocateCommandBuffers(cp.ctx.Device, allocInfo, cmdBuffers); res != vk.Success {
		return nil, fmt.Errorf("allocate single-time command buffer: vulkan result %d", res)
	}

	beginInfo := &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	}

	if res := vk.BeginCommandBuffer(cmdBuffers[0], beginInfo); res != vk.Success {
		vk.FreeCommandBuffers(cp.ctx.Device, cp.Pool, 1, cmdBuffers)
		return nil, fmt.Errorf("begin single-time command buffer: vulkan result %d", res)
	}

	return cmdBuffers[0], nil
}

// endSingleTimeCommands ends, submits, and waits for the command buffer,
// then frees it.
func (cp *CommandPool) endSingleTimeCommands(cmdBuf vk.CommandBuffer) error {
	if res := vk.EndCommandBuffer(cmdBuf); res != vk.Success {
		vk.FreeCommandBuffers(cp.ctx.Device, cp.Pool, 1, []vk.CommandBuffer{cmdBuf})
		return fmt.Errorf("end single-time command buffer: vulkan result %d", res)
	}

	submitInfo := &vk.SubmitInfo{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cmdBuf},
	}

	vk.DeviceWaitIdle(cp.ctx.Device)
	res := vk.QueueSubmit(cp.ctx.GraphicsQueue, 1, []vk.SubmitInfo{*submitInfo}, cp.InFlight[0])
	runtime.KeepAlive(submitInfo)

	if res != vk.Success {
		vk.FreeCommandBuffers(cp.ctx.Device, cp.Pool, 1, []vk.CommandBuffer{cmdBuf})
		return fmt.Errorf("submit single-time command buffer: vulkan result %d", res)
	}

	fences := []vk.Fence{cp.InFlight[0]}
	vk.WaitForFences(cp.ctx.Device, 1, fences, vk.True, vk.MaxUint64)
	vk.ResetFences(cp.ctx.Device, 1, fences)
	vk.FreeCommandBuffers(cp.ctx.Device, cp.Pool, 1, []vk.CommandBuffer{cmdBuf})

	return nil
}

// Cleanup destroys sync objects, the command pool, and all associated
// command buffers.
func (cp *CommandPool) Cleanup() {
	for i := 0; i < maxFramesInFlight; i++ {
		if i < len(cp.ImageAvailable) && cp.ImageAvailable[i] != nil {
			vk.DestroySemaphore(cp.ctx.Device, cp.ImageAvailable[i], nil)
		}
		if i < len(cp.RenderFinished) && cp.RenderFinished[i] != nil {
			vk.DestroySemaphore(cp.ctx.Device, cp.RenderFinished[i], nil)
		}
		if i < len(cp.InFlight) && cp.InFlight[i] != nil {
			vk.DestroyFence(cp.ctx.Device, cp.InFlight[i], nil)
		}
	}
	if cp.Pool != nil {
		vk.DestroyCommandPool(cp.ctx.Device, cp.Pool, nil)
	}
}
