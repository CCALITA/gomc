package render

import (
	"fmt"

	vk "github.com/vulkan-go/vulkan"
)

// maxFramesInFlight is the number of frames that can be processed concurrently.
const maxFramesInFlight = 2

// Swapchain manages the Vulkan swapchain, its image views, and framebuffers.
type Swapchain struct {
	Handle       vk.Swapchain
	Images       []vk.Image
	ImageViews   []vk.ImageView
	Framebuffers []vk.Framebuffer
	ImageFormat  vk.Format
	Extent       vk.Extent2D

	// Depth resources
	DepthImage       vk.Image
	DepthImageView   vk.ImageView
	DepthImageMemory vk.DeviceMemory
	DepthFormat      vk.Format

	ctx *VulkanContext
}

// NewSwapchain creates a Swapchain tied to the given VulkanContext.
func NewSwapchain(ctx *VulkanContext) *Swapchain {
	return &Swapchain{ctx: ctx}
}

// Create builds the swapchain, image views, depth resources, and framebuffers.
func (s *Swapchain) Create(width, height uint32, renderPass vk.RenderPass) error {
	if err := s.createSwapchain(width, height); err != nil {
		return err
	}
	if err := s.createImageViews(); err != nil {
		return err
	}
	if err := s.createDepthResources(); err != nil {
		return err
	}
	if err := s.createFramebuffers(renderPass); err != nil {
		return err
	}
	return nil
}

// Recreate destroys the old swapchain resources and creates new ones,
// typically called after a window resize.
func (s *Swapchain) Recreate(width, height uint32, renderPass vk.RenderPass) error {
	vk.DeviceWaitIdle(s.ctx.Device)
	s.cleanupSwapchain()
	return s.Create(width, height, renderPass)
}

// createSwapchain builds the actual VkSwapchain object.
func (s *Swapchain) createSwapchain(width, height uint32) error {
	details := querySwapchainSupport(s.ctx.PhysicalDevice, s.ctx.Surface)

	surfaceFormat := chooseSwapSurfaceFormat(details.Formats)
	presentMode := chooseSwapPresentMode(details.PresentModes)
	extent := chooseSwapExtent(details.Capabilities, width, height)

	imageCount := details.Capabilities.MinImageCount + 1
	if details.Capabilities.MaxImageCount > 0 && imageCount > details.Capabilities.MaxImageCount {
		imageCount = details.Capabilities.MaxImageCount
	}

	createInfo := &vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          s.ctx.Surface,
		MinImageCount:    imageCount,
		ImageFormat:      surfaceFormat.Format,
		ImageColorSpace:  surfaceFormat.ColorSpace,
		ImageExtent:      extent,
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:     details.Capabilities.CurrentTransform,
		CompositeAlpha:   vk.CompositeAlphaOpaqueBit,
		PresentMode:      presentMode,
		Clipped:          vk.True,
	}

	indices := s.ctx.QueueIndices
	if indices.GraphicsFamily != indices.PresentFamily {
		createInfo.ImageSharingMode = vk.SharingModeConcurrent
		createInfo.QueueFamilyIndexCount = 2
		createInfo.PQueueFamilyIndices = []uint32{indices.GraphicsFamily, indices.PresentFamily}
	} else {
		createInfo.ImageSharingMode = vk.SharingModeExclusive
	}

	var swapchain vk.Swapchain
	if res := vk.CreateSwapchain(s.ctx.Device, createInfo, nil, &swapchain); res != vk.Success {
		return fmt.Errorf("failed to create swapchain: %d", res)
	}
	s.Handle = swapchain
	s.ImageFormat = surfaceFormat.Format
	s.Extent = extent

	var count uint32
	vk.GetSwapchainImages(s.ctx.Device, s.Handle, &count, nil)
	s.Images = make([]vk.Image, count)
	vk.GetSwapchainImages(s.ctx.Device, s.Handle, &count, s.Images)

	return nil
}

// createImageViews creates a VkImageView for each swapchain image.
func (s *Swapchain) createImageViews() error {
	s.ImageViews = make([]vk.ImageView, len(s.Images))

	for i, image := range s.Images {
		createInfo := &vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    image,
			ViewType: vk.ImageViewType2d,
			Format:   s.ImageFormat,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleIdentity,
				G: vk.ComponentSwizzleIdentity,
				B: vk.ComponentSwizzleIdentity,
				A: vk.ComponentSwizzleIdentity,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		}

		var imageView vk.ImageView
		if res := vk.CreateImageView(s.ctx.Device, createInfo, nil, &imageView); res != vk.Success {
			return fmt.Errorf("failed to create image view %d: %d", i, res)
		}
		s.ImageViews[i] = imageView
	}
	return nil
}

// createDepthResources creates the depth image, image view, and memory.
func (s *Swapchain) createDepthResources() error {
	s.DepthFormat = vk.FormatD32Sfloat

	imageInfo := &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    s.DepthFormat,
		Extent: vk.Extent3D{
			Width:  s.Extent.Width,
			Height: s.Extent.Height,
			Depth:  1,
		},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage:       vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
	}

	var depthImage vk.Image
	if res := vk.CreateImage(s.ctx.Device, imageInfo, nil, &depthImage); res != vk.Success {
		return fmt.Errorf("failed to create depth image: %d", res)
	}
	s.DepthImage = depthImage

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(s.ctx.Device, s.DepthImage, &memReqs)
	memReqs.Deref()

	memTypeIndex, err := findMemoryType(
		s.ctx.PhysicalDevice,
		memReqs.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
	)
	if err != nil {
		return fmt.Errorf("failed to find depth image memory type: %w", err)
	}

	allocInfo := &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var memory vk.DeviceMemory
	if res := vk.AllocateMemory(s.ctx.Device, allocInfo, nil, &memory); res != vk.Success {
		return fmt.Errorf("failed to allocate depth image memory: %d", res)
	}
	s.DepthImageMemory = memory

	vk.BindImageMemory(s.ctx.Device, s.DepthImage, s.DepthImageMemory, 0)

	viewInfo := &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    s.DepthImage,
		ViewType: vk.ImageViewType2d,
		Format:   s.DepthFormat,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}

	var depthImageView vk.ImageView
	if res := vk.CreateImageView(s.ctx.Device, viewInfo, nil, &depthImageView); res != vk.Success {
		return fmt.Errorf("failed to create depth image view: %d", res)
	}
	s.DepthImageView = depthImageView

	return nil
}

// createFramebuffers creates one framebuffer per swapchain image.
func (s *Swapchain) createFramebuffers(renderPass vk.RenderPass) error {
	s.Framebuffers = make([]vk.Framebuffer, len(s.ImageViews))

	for i, imageView := range s.ImageViews {
		attachments := []vk.ImageView{imageView, s.DepthImageView}

		fbInfo := &vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      renderPass,
			AttachmentCount: uint32(len(attachments)),
			PAttachments:    attachments,
			Width:           s.Extent.Width,
			Height:          s.Extent.Height,
			Layers:          1,
		}

		var framebuffer vk.Framebuffer
		if res := vk.CreateFramebuffer(s.ctx.Device, fbInfo, nil, &framebuffer); res != vk.Success {
			return fmt.Errorf("failed to create framebuffer %d: %d", i, res)
		}
		s.Framebuffers[i] = framebuffer
	}
	return nil
}

// cleanupSwapchain destroys swapchain-related resources.
func (s *Swapchain) cleanupSwapchain() {
	for _, fb := range s.Framebuffers {
		vk.DestroyFramebuffer(s.ctx.Device, fb, nil)
	}
	s.Framebuffers = nil

	if s.DepthImageView != nil {
		vk.DestroyImageView(s.ctx.Device, s.DepthImageView, nil)
	}
	if s.DepthImage != nil {
		vk.DestroyImage(s.ctx.Device, s.DepthImage, nil)
	}
	if s.DepthImageMemory != nil {
		vk.FreeMemory(s.ctx.Device, s.DepthImageMemory, nil)
	}

	for _, iv := range s.ImageViews {
		vk.DestroyImageView(s.ctx.Device, iv, nil)
	}
	s.ImageViews = nil

	if s.Handle != nil {
		vk.DestroySwapchain(s.ctx.Device, s.Handle, nil)
	}
}

// Cleanup destroys all swapchain resources.
func (s *Swapchain) Cleanup() {
	s.cleanupSwapchain()
}

// chooseSwapSurfaceFormat selects the preferred surface format.
// Prefers B8G8R8A8_SRGB with SRGB_NONLINEAR colour space.
func chooseSwapSurfaceFormat(formats []vk.SurfaceFormat) vk.SurfaceFormat {
	for _, f := range formats {
		if f.Format == vk.FormatB8g8r8a8Srgb &&
			f.ColorSpace == vk.ColorSpaceSrgbNonlinear {
			return f
		}
	}
	if len(formats) > 0 {
		return formats[0]
	}
	return vk.SurfaceFormat{
		Format:     vk.FormatB8g8r8a8Srgb,
		ColorSpace: vk.ColorSpaceSrgbNonlinear,
	}
}

// chooseSwapPresentMode selects the preferred present mode.
// Prefers Mailbox for triple buffering; falls back to FIFO.
func chooseSwapPresentMode(modes []vk.PresentMode) vk.PresentMode {
	for _, m := range modes {
		if m == vk.PresentModeMailbox {
			return vk.PresentModeMailbox
		}
	}
	return vk.PresentModeFifo
}

// chooseSwapExtent selects the swap extent matching the window dimensions.
func chooseSwapExtent(capabilities vk.SurfaceCapabilities, width, height uint32) vk.Extent2D {
	capabilities.CurrentExtent.Deref()
	if capabilities.CurrentExtent.Width != vk.MaxUint32 {
		return capabilities.CurrentExtent
	}

	capabilities.MinImageExtent.Deref()
	capabilities.MaxImageExtent.Deref()

	return vk.Extent2D{
		Width:  clampUint32(width, capabilities.MinImageExtent.Width, capabilities.MaxImageExtent.Width),
		Height: clampUint32(height, capabilities.MinImageExtent.Height, capabilities.MaxImageExtent.Height),
	}
}

// clampUint32 clamps val between lo and hi (inclusive).
func clampUint32(val, lo, hi uint32) uint32 {
	if val < lo {
		return lo
	}
	if val > hi {
		return hi
	}
	return val
}
