package render

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/go-gl/mathgl/mgl32"
	vk "github.com/vulkan-go/vulkan"
)

// wireframeScale is slightly larger than 1.0 to prevent z-fighting with the
// block face beneath the outline.
const wireframeScale = 1.002

// wireframeOffset centres the scaled cube on the block so that the 0.002
// expansion is evenly distributed on each side.
const wireframeOffset = (1.0 - wireframeScale) / 2.0

// WireframePushConstants is the push constant block for the wireframe pipeline.
// It carries the MVP matrix (16 floats = 64 bytes).
type WireframePushConstants struct {
	MVP [16]float32
}

// WireframeRenderer draws a wireframe cube outline around a targeted block.
type WireframeRenderer struct {
	mu     sync.RWMutex
	target mcmath.BlockPos
	active bool

	ctx  *VulkanContext
	pipe *Pipeline
}

// NewWireframeRenderer creates a WireframeRenderer. The caller must supply a
// Pipeline that has already been initialised with a line-list topology and the
// wireframe shaders.
func NewWireframeRenderer(ctx *VulkanContext, pipe *Pipeline) *WireframeRenderer {
	return &WireframeRenderer{
		ctx:  ctx,
		pipe: pipe,
	}
}

// SetTarget sets the block to highlight with the wireframe outline.
func (wr *WireframeRenderer) SetTarget(pos mcmath.BlockPos) {
	wr.mu.Lock()
	wr.target = pos
	wr.active = true
	wr.mu.Unlock()
}

// ClearTarget removes the wireframe highlight.
func (wr *WireframeRenderer) ClearTarget() {
	wr.mu.Lock()
	wr.active = false
	wr.mu.Unlock()
}

// IsActive returns whether a target block is currently set.
func (wr *WireframeRenderer) IsActive() bool {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	return wr.active
}

// Target returns the current target block position and whether the wireframe
// is active.
func (wr *WireframeRenderer) Target() (mcmath.BlockPos, bool) {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	return wr.target, wr.active
}

// Draw records draw commands for the wireframe outline around the target block.
// It uses push constants to pass the MVP matrix and draws 12 line segments
// (24 vertices) as a VK_PRIMITIVE_TOPOLOGY_LINE_LIST.
func (wr *WireframeRenderer) Draw(cmdBuf vk.CommandBuffer, camera *Camera, aspect float32) {
	wr.mu.RLock()
	if !wr.active {
		wr.mu.RUnlock()
		return
	}
	pos := wr.target
	wr.mu.RUnlock()

	vp := camera.ViewProjectionMatrix(aspect)

	// Build a model matrix: translate to block position + offset, then scale
	// by wireframeScale. The cube vertices in the shader are in [0,1], so this
	// produces a 1.002-sized cube centred on the block.
	model := mgl32.Translate3D(
		float32(pos.X)+wireframeOffset,
		float32(pos.Y)+wireframeOffset,
		float32(pos.Z)+wireframeOffset,
	).Mul4(mgl32.Scale3D(wireframeScale, wireframeScale, wireframeScale))

	mvp := vp.Mul4(model)

	pc := WireframePushConstants{MVP: [16]float32(mvp)}

	vk.CmdBindPipeline(cmdBuf, vk.PipelineBindPointGraphics, wr.pipe.GraphicsPipeline)

	vk.CmdPushConstants(
		cmdBuf,
		wr.pipe.PipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		0,
		uint32(unsafe.Sizeof(pc)),
		unsafe.Pointer(&pc),
	)

	// Draw 24 vertices (12 lines) without a vertex buffer; the vertex shader
	// uses gl_VertexIndex to look up positions from a constant array.
	vk.CmdDraw(cmdBuf, 24, 1, 0, 0)
}

// Cleanup is a no-op since the WireframeRenderer does not own GPU buffers.
// The Pipeline must be cleaned up separately by the caller.
func (wr *WireframeRenderer) Cleanup() {
	// No GPU resources to free.
}

// WireframeCubeVertices returns the 24 vertices (as float32 triples) for the
// 12 line segments that make up a unit cube outline with corners at (0,0,0)
// and (1,1,1). This is exported for testing.
func WireframeCubeVertices() []float32 {
	corners := [8][3]float32{
		{0, 0, 0}, // 0
		{1, 0, 0}, // 1
		{1, 1, 0}, // 2
		{0, 1, 0}, // 3
		{0, 0, 1}, // 4
		{1, 0, 1}, // 5
		{1, 1, 1}, // 6
		{0, 1, 1}, // 7
	}

	// 12 edges of a cube, each as a pair of corner indices.
	edges := [12][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, // bottom face
		{4, 5}, {5, 6}, {6, 7}, {7, 4}, // top face
		{0, 4}, {1, 5}, {2, 6}, {3, 7}, // vertical edges
	}

	vertices := make([]float32, 0, 24*3)
	for _, edge := range edges {
		a := corners[edge[0]]
		b := corners[edge[1]]
		vertices = append(vertices, a[0], a[1], a[2])
		vertices = append(vertices, b[0], b[1], b[2])
	}
	return vertices
}

// WireframeBlockAABB returns the AABB for the wireframe outline at the given
// block position. The box is slightly larger than 1.0 due to wireframeScale.
func WireframeBlockAABB(pos mcmath.BlockPos) mcmath.AABB {
	return mcmath.AABB{
		Min: mcmath.Vec3{
			X: float32(pos.X) + wireframeOffset,
			Y: float32(pos.Y) + wireframeOffset,
			Z: float32(pos.Z) + wireframeOffset,
		},
		Max: mcmath.Vec3{
			X: float32(pos.X) + wireframeOffset + wireframeScale,
			Y: float32(pos.Y) + wireframeOffset + wireframeScale,
			Z: float32(pos.Z) + wireframeOffset + wireframeScale,
		},
	}
}

// CreateWireframePipeline creates a Vulkan graphics pipeline configured for
// wireframe line rendering. It uses VK_PRIMITIVE_TOPOLOGY_LINE_LIST, no
// vertex input (vertices are generated in the shader), and push constants
// for the MVP matrix.
func CreateWireframePipeline(ctx *VulkanContext, renderPass vk.RenderPass, vertShaderPath, fragShaderPath string, extent vk.Extent2D) (*Pipeline, error) {
	p := NewPipeline(ctx)
	p.RenderPass = renderPass

	vertCode, err := os.ReadFile(vertShaderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wireframe vertex shader: %w", err)
	}
	fragCode, err := os.ReadFile(fragShaderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wireframe fragment shader: %w", err)
	}

	vertModule, err := createShaderModule(ctx.Device, vertCode)
	if err != nil {
		return nil, fmt.Errorf("failed to create wireframe vertex shader module: %w", err)
	}
	defer vk.DestroyShaderModule(ctx.Device, vertModule, nil)

	fragModule, err := createShaderModule(ctx.Device, fragCode)
	if err != nil {
		return nil, fmt.Errorf("failed to create wireframe fragment shader module: %w", err)
	}
	defer vk.DestroyShaderModule(ctx.Device, fragModule, nil)

	vertStage := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageVertexBit,
		Module: vertModule,
		PName:  "main\x00",
	}

	fragStage := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageFragmentBit,
		Module: fragModule,
		PName:  "main\x00",
	}

	shaderStages := []vk.PipelineShaderStageCreateInfo{vertStage, fragStage}

	// No vertex input -- vertices are generated in the vertex shader.
	vertexInputInfo := &vk.PipelineVertexInputStateCreateInfo{
		SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
	}

	inputAssembly := &vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               vk.PrimitiveTopologyLineList,
		PrimitiveRestartEnable: vk.False,
	}

	viewport := vk.Viewport{
		X: 0, Y: 0,
		Width:    float32(extent.Width),
		Height:   float32(extent.Height),
		MinDepth: 0,
		MaxDepth: 1,
	}

	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: extent,
	}

	viewportState := &vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    []vk.Viewport{viewport},
		ScissorCount:  1,
		PScissors:     []vk.Rect2D{scissor},
	}

	rasterizer := &vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        vk.False,
		RasterizerDiscardEnable: vk.False,
		PolygonMode:             vk.PolygonModeFill,
		LineWidth:               2.0,
		CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceCounterClockwise,
		DepthBiasEnable:         vk.False,
	}

	multisampling := &vk.PipelineMultisampleStateCreateInfo{
		SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
		SampleShadingEnable:  vk.False,
		RasterizationSamples: vk.SampleCount1Bit,
	}

	depthStencil := &vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       vk.True,
		DepthWriteEnable:      vk.False,
		DepthCompareOp:        vk.CompareOpLessOrEqual,
		DepthBoundsTestEnable: vk.False,
		StencilTestEnable:     vk.False,
	}

	colorBlendAttachment := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(
			vk.ColorComponentRBit | vk.ColorComponentGBit |
				vk.ColorComponentBBit | vk.ColorComponentABit,
		),
		BlendEnable: vk.False,
	}

	colorBlending := &vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{colorBlendAttachment},
	}

	// Push constants: 64 bytes for the MVP matrix (mat4).
	pushConstantRange := vk.PushConstantRange{
		StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		Offset:     0,
		Size:       64,
	}

	pipelineLayoutInfo := &vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		PushConstantRangeCount: 1,
		PPushConstantRanges:    []vk.PushConstantRange{pushConstantRange},
	}

	var pipelineLayout vk.PipelineLayout
	if res := vk.CreatePipelineLayout(ctx.Device, pipelineLayoutInfo, nil, &pipelineLayout); res != vk.Success {
		return nil, fmt.Errorf("failed to create wireframe pipeline layout: %d", res)
	}
	p.PipelineLayout = pipelineLayout

	dynamicStates := []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateScissor}
	dynamicState := &vk.PipelineDynamicStateCreateInfo{
		SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynamicStates)),
		PDynamicStates:    dynamicStates,
	}

	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          uint32(len(shaderStages)),
		PStages:             shaderStages,
		PVertexInputState:   vertexInputInfo,
		PInputAssemblyState: inputAssembly,
		PViewportState:      viewportState,
		PRasterizationState: rasterizer,
		PMultisampleState:   multisampling,
		PDepthStencilState:  depthStencil,
		PColorBlendState:    colorBlending,
		PDynamicState:       dynamicState,
		Layout:              pipelineLayout,
		RenderPass:          renderPass,
		Subpass:             0,
	}

	pipelines := make([]vk.Pipeline, 1)
	if res := vk.CreateGraphicsPipelines(ctx.Device, nil, 1, []vk.GraphicsPipelineCreateInfo{pipelineInfo}, nil, pipelines); res != vk.Success {
		return nil, fmt.Errorf("failed to create wireframe graphics pipeline: %d", res)
	}
	p.GraphicsPipeline = pipelines[0]

	return p, nil
}
