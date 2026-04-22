package render

import (
	"fmt"
	"os"

	vk "github.com/vulkan-go/vulkan"
)

// vertexStride is the byte stride for a chunk vertex:
// position(vec3) + normal(vec3) + texcoord(vec2) + ao(float) = 9 floats.
const vertexStride = 9 * 4 // 36 bytes

// Pipeline holds the Vulkan render pass, graphics pipeline, and related
// layout objects.
type Pipeline struct {
	RenderPass       vk.RenderPass
	PipelineLayout   vk.PipelineLayout
	GraphicsPipeline vk.Pipeline
	DescriptorLayout vk.DescriptorSetLayout

	ctx *VulkanContext
}

// NewPipeline creates a Pipeline tied to the given VulkanContext.
func NewPipeline(ctx *VulkanContext) *Pipeline {
	return &Pipeline{ctx: ctx}
}

// CreateRenderPass creates a render pass with colour and depth attachments.
func (p *Pipeline) CreateRenderPass(colorFormat, depthFormat vk.Format) error {
	colorAttachment := vk.AttachmentDescription{
		Format:         colorFormat,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutPresentSrc,
	}

	depthAttachment := vk.AttachmentDescription{
		Format:         depthFormat,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpDontCare,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
	}

	colorRef := vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}

	depthRef := vk.AttachmentReference{
		Attachment: 1,
		Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
	}

	subpass := vk.SubpassDescription{
		PipelineBindPoint:       vk.PipelineBindPointGraphics,
		ColorAttachmentCount:    1,
		PColorAttachments:       []vk.AttachmentReference{colorRef},
		PDepthStencilAttachment: &depthRef,
	}

	dependency := vk.SubpassDependency{
		SrcSubpass:    vk.SubpassExternal,
		DstSubpass:    0,
		SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit | vk.PipelineStageEarlyFragmentTestsBit),
		DstStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit | vk.PipelineStageEarlyFragmentTestsBit),
		SrcAccessMask: 0,
		DstAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit | vk.AccessDepthStencilAttachmentWriteBit),
	}

	renderPassInfo := &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 2,
		PAttachments:    []vk.AttachmentDescription{colorAttachment, depthAttachment},
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
		DependencyCount: 1,
		PDependencies:   []vk.SubpassDependency{dependency},
	}

	var renderPass vk.RenderPass
	if res := vk.CreateRenderPass(p.ctx.Device, renderPassInfo, nil, &renderPass); res != vk.Success {
		return fmt.Errorf("failed to create render pass: %d", res)
	}
	p.RenderPass = renderPass
	return nil
}

// CreateGraphicsPipeline loads the given SPIR-V shader files and creates
// the graphics pipeline with appropriate vertex input, push constants,
// and descriptor set layout.
func (p *Pipeline) CreateGraphicsPipeline(vertShaderPath, fragShaderPath string, extent vk.Extent2D) error {
	vertCode, err := os.ReadFile(vertShaderPath)
	if err != nil {
		return fmt.Errorf("failed to read vertex shader: %w", err)
	}
	fragCode, err := os.ReadFile(fragShaderPath)
	if err != nil {
		return fmt.Errorf("failed to read fragment shader: %w", err)
	}

	vertModule, err := createShaderModule(p.ctx.Device, vertCode)
	if err != nil {
		return fmt.Errorf("failed to create vertex shader module: %w", err)
	}
	defer vk.DestroyShaderModule(p.ctx.Device, vertModule, nil)

	fragModule, err := createShaderModule(p.ctx.Device, fragCode)
	if err != nil {
		return fmt.Errorf("failed to create fragment shader module: %w", err)
	}
	defer vk.DestroyShaderModule(p.ctx.Device, fragModule, nil)

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

	// Vertex input: position(3) + normal(3) + texcoord(2) + ao(1) = 9 floats
	bindingDesc := vk.VertexInputBindingDescription{
		Binding:   0,
		Stride:    vertexStride,
		InputRate: vk.VertexInputRateVertex,
	}

	attrDescs := []vk.VertexInputAttributeDescription{
		{Location: 0, Binding: 0, Format: vk.FormatR32g32b32Sfloat, Offset: 0},     // position
		{Location: 1, Binding: 0, Format: vk.FormatR32g32b32Sfloat, Offset: 3 * 4}, // normal
		{Location: 2, Binding: 0, Format: vk.FormatR32g32Sfloat, Offset: 6 * 4},     // texcoord
		{Location: 3, Binding: 0, Format: vk.FormatR32Sfloat, Offset: 8 * 4},         // ao
	}

	vertexInputInfo := &vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   1,
		PVertexBindingDescriptions:      []vk.VertexInputBindingDescription{bindingDesc},
		VertexAttributeDescriptionCount: uint32(len(attrDescs)),
		PVertexAttributeDescriptions:    attrDescs,
	}

	inputAssembly := &vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               vk.PrimitiveTopologyTriangleList,
		PrimitiveRestartEnable: vk.False,
	}

	viewport := vk.Viewport{
		X:        0,
		Y:        0,
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
		LineWidth:               1,
		CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceClockwise,
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
		DepthWriteEnable:      vk.True,
		DepthCompareOp:        vk.CompareOpLess,
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

	// Descriptor set layout for uniform buffer (set 0, binding 0)
	uboLayoutBinding := vk.DescriptorSetLayoutBinding{
		Binding:         0,
		DescriptorType:  vk.DescriptorTypeUniformBuffer,
		DescriptorCount: 1,
		StageFlags:      vk.ShaderStageFlags(vk.ShaderStageVertexBit),
	}

	layoutBindings := []vk.DescriptorSetLayoutBinding{uboLayoutBinding}

	descriptorLayoutInfo := &vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: uint32(len(layoutBindings)),
		PBindings:    layoutBindings,
	}

	var descriptorLayout vk.DescriptorSetLayout
	if res := vk.CreateDescriptorSetLayout(p.ctx.Device, descriptorLayoutInfo, nil, &descriptorLayout); res != vk.Success {
		return fmt.Errorf("failed to create descriptor set layout: %d", res)
	}
	p.DescriptorLayout = descriptorLayout

	// Push constants for chunk world position (vec3 = 12 bytes, padded to 16)
	pushConstantRange := vk.PushConstantRange{
		StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		Offset:     0,
		Size:       64, // mat4 model matrix
	}

	pipelineLayoutInfo := &vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descriptorLayout},
		PushConstantRangeCount: 1,
		PPushConstantRanges:    []vk.PushConstantRange{pushConstantRange},
	}

	var pipelineLayout vk.PipelineLayout
	if res := vk.CreatePipelineLayout(p.ctx.Device, pipelineLayoutInfo, nil, &pipelineLayout); res != vk.Success {
		return fmt.Errorf("failed to create pipeline layout: %d", res)
	}
	p.PipelineLayout = pipelineLayout

	// Dynamic states for viewport and scissor
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
		RenderPass:          p.RenderPass,
		Subpass:             0,
	}

	pipelines := make([]vk.Pipeline, 1)
	if res := vk.CreateGraphicsPipelines(p.ctx.Device, nil, 1, []vk.GraphicsPipelineCreateInfo{pipelineInfo}, nil, pipelines); res != vk.Success {
		return fmt.Errorf("failed to create graphics pipeline: %d", res)
	}
	p.GraphicsPipeline = pipelines[0]

	return nil
}

// createShaderModule creates a Vulkan shader module from SPIR-V bytecode.
func createShaderModule(device vk.Device, code []byte) (vk.ShaderModule, error) {
	createInfo := &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(code)),
		PCode:    sliceUint32(code),
	}

	var module vk.ShaderModule
	if res := vk.CreateShaderModule(device, createInfo, nil, &module); res != vk.Success {
		return nil, fmt.Errorf("failed to create shader module: %d", res)
	}
	return module, nil
}

// Cleanup destroys the pipeline, layout, descriptor set layout, and render pass.
func (p *Pipeline) Cleanup() {
	if p.GraphicsPipeline != nil {
		vk.DestroyPipeline(p.ctx.Device, p.GraphicsPipeline, nil)
	}
	if p.PipelineLayout != nil {
		vk.DestroyPipelineLayout(p.ctx.Device, p.PipelineLayout, nil)
	}
	if p.DescriptorLayout != nil {
		vk.DestroyDescriptorSetLayout(p.ctx.Device, p.DescriptorLayout, nil)
	}
	if p.RenderPass != nil {
		vk.DestroyRenderPass(p.ctx.Device, p.RenderPass, nil)
	}
}
