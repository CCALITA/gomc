package render

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

// Renderer is the top-level rendering facade that ties together the Vulkan
// context, swapchain, pipeline, command pool, and chunk renderer.
type Renderer struct {
	Window *glfw.Window

	Context       VulkanContext
	Swapchain     *Swapchain
	Pipeline      *Pipeline
	CmdPool       *CommandPool
	ChunkRenderer *ChunkRenderer

	width  uint32
	height uint32

	framebufferResized bool
}

// Init creates a GLFW window and initialises the full Vulkan rendering stack.
func (r *Renderer) Init(windowWidth, windowHeight int, title string) error {
	if err := glfw.Init(); err != nil {
		return fmt.Errorf("failed to initialise GLFW: %w", err)
	}

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	glfw.WindowHint(glfw.Resizable, glfw.True)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, title, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create GLFW window: %w", err)
	}
	r.Window = window
	r.width = uint32(windowWidth)
	r.height = uint32(windowHeight)

	r.Window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		r.framebufferResized = true
		r.width = uint32(width)
		r.height = uint32(height)
	})

	// Vulkan context
	if err := r.Context.Init(r.Window); err != nil {
		return fmt.Errorf("failed to init vulkan context: %w", err)
	}

	// Pipeline and render pass
	r.Pipeline = NewPipeline(&r.Context)
	r.Swapchain = NewSwapchain(&r.Context)

	if err := r.Pipeline.CreateRenderPass(vk.FormatB8g8r8a8Srgb, vk.FormatD32Sfloat); err != nil {
		return fmt.Errorf("failed to create render pass: %w", err)
	}

	// Swapchain
	if err := r.Swapchain.Create(r.width, r.height, r.Pipeline.RenderPass); err != nil {
		return fmt.Errorf("failed to create swapchain: %w", err)
	}

	// Graphics pipeline
	if err := r.Pipeline.CreateGraphicsPipeline(
		"assets/shaders/chunk.vert.spv",
		"assets/shaders/chunk.frag.spv",
		r.Swapchain.Extent,
	); err != nil {
		return fmt.Errorf("failed to create graphics pipeline: %w", err)
	}

	// Command pool
	cmdPool, err := NewCommandPool(&r.Context)
	if err != nil {
		return fmt.Errorf("failed to create command pool: %w", err)
	}
	r.CmdPool = cmdPool

	// Chunk renderer
	chunkRenderer, err := NewChunkRenderer(&r.Context, r.CmdPool, r.Pipeline)
	if err != nil {
		return fmt.Errorf("failed to create chunk renderer: %w", err)
	}
	r.ChunkRenderer = chunkRenderer

	return nil
}

// BeginFrame acquires the next swapchain image, begins recording commands,
// and starts the render pass. Returns the image index and command buffer,
// or an error if the swapchain needs recreation.
func (r *Renderer) BeginFrame() (uint32, vk.CommandBuffer, error) {
	imageIndex, cmdBuf, err := r.CmdPool.BeginFrame(r.Swapchain)
	if err != nil {
		return 0, nil, err
	}

	// Set dynamic viewport and scissor.
	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(r.Swapchain.Extent.Width),
		Height:   float32(r.Swapchain.Extent.Height),
		MinDepth: 0,
		MaxDepth: 1,
	}
	vk.CmdSetViewport(cmdBuf, 0, 1, []vk.Viewport{viewport})

	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: r.Swapchain.Extent,
	}
	vk.CmdSetScissor(cmdBuf, 0, 1, []vk.Rect2D{scissor})

	BeginRenderPass(cmdBuf, r.Pipeline.RenderPass, r.Swapchain.Framebuffers[imageIndex], r.Swapchain.Extent)

	return imageIndex, cmdBuf, nil
}

// EndFrame finishes recording, submits the command buffer, and presents.
func (r *Renderer) EndFrame(imageIndex uint32) error {
	err := r.CmdPool.EndFrame(r.Swapchain, imageIndex)
	if err != nil && r.framebufferResized {
		r.framebufferResized = false
		return r.recreateSwapchain()
	}
	return err
}

// DrawChunks records draw commands for all uploaded chunk meshes.
func (r *Renderer) DrawChunks(cmdBuf vk.CommandBuffer, camera *Camera) {
	aspect := float32(r.Swapchain.Extent.Width) / float32(r.Swapchain.Extent.Height)
	r.ChunkRenderer.DrawAll(cmdBuf, camera, aspect, r.CmdPool.CurrentFrame)
}

// DrawUI is a placeholder for future UI rendering (HUD, inventory, etc.).
func (r *Renderer) DrawUI(_ vk.CommandBuffer) {
	// UI rendering will be implemented in a future pass.
}

// Resize handles a window resize event by flagging the swapchain for
// recreation.
func (r *Renderer) Resize(w, h int) {
	r.width = uint32(w)
	r.height = uint32(h)
	r.framebufferResized = true
}

// Aspect returns the current window aspect ratio (width / height).
func (r *Renderer) Aspect() float32 {
	if r.height == 0 {
		return 1.0
	}
	return float32(r.width) / float32(r.height)
}

// recreateSwapchain rebuilds the swapchain after a resize or suboptimal
// present.
func (r *Renderer) recreateSwapchain() error {
	// Wait for a non-zero framebuffer size (window minimised).
	w, h := r.Window.GetFramebufferSize()
	for w == 0 || h == 0 {
		glfw.WaitEvents()
		w, h = r.Window.GetFramebufferSize()
	}

	vk.DeviceWaitIdle(r.Context.Device)

	r.width = uint32(w)
	r.height = uint32(h)

	return r.Swapchain.Recreate(r.width, r.height, r.Pipeline.RenderPass)
}

// Cleanup destroys all rendering resources in reverse initialisation order.
func (r *Renderer) Cleanup() {
	if r.Context.Device != nil {
		vk.DeviceWaitIdle(r.Context.Device)
	}

	if r.ChunkRenderer != nil {
		r.ChunkRenderer.Cleanup()
	}
	if r.CmdPool != nil {
		r.CmdPool.Cleanup()
	}
	if r.Pipeline != nil {
		r.Pipeline.Cleanup()
	}
	if r.Swapchain != nil {
		r.Swapchain.Cleanup()
	}
	r.Context.Cleanup()

	if r.Window != nil {
		r.Window.Destroy()
	}
	glfw.Terminate()
}
