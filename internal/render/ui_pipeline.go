// Package render — UI pipeline for 2D overlay rendering.
package render

import (
	"fmt"
	"os"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// UIVertex is a 2D vertex with color (no texture for now).
// Layout: x, y, r, g, b, a (6 floats per vertex, 24 bytes).
type UIVertex struct {
	X, Y       float32
	R, G, B, A float32
}

const uiVertexStride = 6 * 4 // 24 bytes

// UIPipeline is a separate Vulkan pipeline for UI overlay rendering.
type UIPipeline struct {
	GraphicsPipeline vk.Pipeline
	PipelineLayout   vk.PipelineLayout
	ctx              *VulkanContext
}

// uiPushConstants — screen size for NDC conversion.
type uiPushConstants struct {
	ScreenW, ScreenH float32
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// NewUIPipeline creates a UI pipeline that draws solid-colored 2D quads.
func NewUIPipeline(ctx *VulkanContext, renderPass vk.RenderPass) (*UIPipeline, error) {
	p := &UIPipeline{ctx: ctx}

	vertCode, err := readFile("assets/shaders/ui_solid.vert.spv")
	if err != nil {
		return nil, fmt.Errorf("read ui vertex shader: %w", err)
	}
	fragCode, err := readFile("assets/shaders/ui_solid.frag.spv")
	if err != nil {
		return nil, fmt.Errorf("read ui fragment shader: %w", err)
	}

	vertModule, err := createShaderModule(ctx.Device, vertCode)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(ctx.Device, vertModule, nil)
	fragModule, err := createShaderModule(ctx.Device, fragCode)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(ctx.Device, fragModule, nil)

	stages := []vk.PipelineShaderStageCreateInfo{
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageVertexBit,
			Module: vertModule,
			PName:  "main\x00",
		},
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageFragmentBit,
			Module: fragModule,
			PName:  "main\x00",
		},
	}

	bindingDesc := vk.VertexInputBindingDescription{
		Binding:   0,
		Stride:    uiVertexStride,
		InputRate: vk.VertexInputRateVertex,
	}
	attrDescs := []vk.VertexInputAttributeDescription{
		{Binding: 0, Location: 0, Format: vk.FormatR32g32Sfloat, Offset: 0},      // pos
		{Binding: 0, Location: 1, Format: vk.FormatR32g32b32a32Sfloat, Offset: 8}, // color
	}
	vertexInput := &vk.PipelineVertexInputStateCreateInfo{
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

	viewportState := &vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		ScissorCount:  1,
	}

	rasterizer := &vk.PipelineRasterizationStateCreateInfo{
		SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
		PolygonMode: vk.PolygonModeFill,
		LineWidth:   1,
		CullMode:    vk.CullModeFlags(vk.CullModeNone),
		FrontFace:   vk.FrontFaceClockwise,
	}

	multisampling := &vk.PipelineMultisampleStateCreateInfo{
		SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples: vk.SampleCount1Bit,
	}

	depthStencil := &vk.PipelineDepthStencilStateCreateInfo{
		SType:            vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:  vk.False,
		DepthWriteEnable: vk.False,
	}

	colorBlendAttachment := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(
			vk.ColorComponentRBit | vk.ColorComponentGBit |
				vk.ColorComponentBBit | vk.ColorComponentABit,
		),
		BlendEnable:         vk.True,
		SrcColorBlendFactor: vk.BlendFactorSrcAlpha,
		DstColorBlendFactor: vk.BlendFactorOneMinusSrcAlpha,
		ColorBlendOp:        vk.BlendOpAdd,
		SrcAlphaBlendFactor: vk.BlendFactorOne,
		DstAlphaBlendFactor: vk.BlendFactorZero,
		AlphaBlendOp:        vk.BlendOpAdd,
	}

	colorBlending := &vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{colorBlendAttachment},
	}

	dynamicStates := []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateScissor}
	dynamicState := &vk.PipelineDynamicStateCreateInfo{
		SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynamicStates)),
		PDynamicStates:    dynamicStates,
	}

	pushConstantRange := vk.PushConstantRange{
		StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		Offset:     0,
		Size:       uint32(unsafe.Sizeof(uiPushConstants{})),
	}

	layoutInfo := &vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		PushConstantRangeCount: 1,
		PPushConstantRanges:    []vk.PushConstantRange{pushConstantRange},
	}
	if res := vk.CreatePipelineLayout(ctx.Device, layoutInfo, nil, &p.PipelineLayout); res != vk.Success {
		return nil, fmt.Errorf("create ui pipeline layout: vulkan result %d", res)
	}

	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          uint32(len(stages)),
		PStages:             stages,
		PVertexInputState:   vertexInput,
		PInputAssemblyState: inputAssembly,
		PViewportState:      viewportState,
		PRasterizationState: rasterizer,
		PMultisampleState:   multisampling,
		PDepthStencilState:  depthStencil,
		PColorBlendState:    colorBlending,
		PDynamicState:       dynamicState,
		Layout:              p.PipelineLayout,
		RenderPass:          renderPass,
		Subpass:             0,
	}
	pipelines := make([]vk.Pipeline, 1)
	if res := vk.CreateGraphicsPipelines(ctx.Device, nil, 1, []vk.GraphicsPipelineCreateInfo{pipelineInfo}, nil, pipelines); res != vk.Success {
		vk.DestroyPipelineLayout(ctx.Device, p.PipelineLayout, nil)
		return nil, fmt.Errorf("create ui graphics pipeline: vulkan result %d", res)
	}
	p.GraphicsPipeline = pipelines[0]
	return p, nil
}

// Cleanup destroys the pipeline.
func (p *UIPipeline) Cleanup() {
	if p.GraphicsPipeline != nil {
		vk.DestroyPipeline(p.ctx.Device, p.GraphicsPipeline, nil)
	}
	if p.PipelineLayout != nil {
		vk.DestroyPipelineLayout(p.ctx.Device, p.PipelineLayout, nil)
	}
}
