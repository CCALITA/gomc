package render

import (
	"fmt"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// Texture wraps a Vulkan image, image view, sampler, and associated memory
// for use as a GPU texture.
type Texture struct {
	Image     vk.Image
	ImageView vk.ImageView
	Sampler   vk.Sampler
	Memory    vk.DeviceMemory

	Width  uint32
	Height uint32

	ctx *VulkanContext
}

// CreateTexture creates a Vulkan texture from raw RGBA image data.
// The imageData must be width * height * 4 bytes (RGBA8).
func CreateTexture(ctx *VulkanContext, cmdPool *CommandPool, imageData []byte, width, height int) (*Texture, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	imageSize := vk.DeviceSize(width * height * 4)

	// Create staging buffer
	staging, err := CreateBuffer(
		ctx,
		imageSize,
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create texture staging buffer: %w", err)
	}
	defer staging.Cleanup()

	if err := mapAndCopy(ctx.Device, staging.Memory, imageSize, unsafe.Pointer(&imageData[0])); err != nil {
		return nil, fmt.Errorf("failed to copy texture data to staging: %w", err)
	}

	tex := &Texture{
		Width:  uint32(width),
		Height: uint32(height),
		ctx:    ctx,
	}

	// Create the VkImage
	imageInfo := &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    vk.FormatR8g8b8a8Srgb,
		Extent: vk.Extent3D{
			Width:  tex.Width,
			Height: tex.Height,
			Depth:  1,
		},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage:       vk.ImageUsageFlags(vk.ImageUsageTransferDstBit | vk.ImageUsageSampledBit),
	}

	var image vk.Image
	if res := vk.CreateImage(ctx.Device, imageInfo, nil, &image); res != vk.Success {
		return nil, fmt.Errorf("failed to create texture image: %d", res)
	}
	tex.Image = image

	// Allocate and bind memory
	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(ctx.Device, tex.Image, &memReqs)
	memReqs.Deref()

	memTypeIndex, err := findMemoryType(
		ctx.PhysicalDevice,
		memReqs.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find texture memory type: %w", err)
	}

	allocInfo := &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var memory vk.DeviceMemory
	if res := vk.AllocateMemory(ctx.Device, allocInfo, nil, &memory); res != vk.Success {
		return nil, fmt.Errorf("failed to allocate texture memory: %d", res)
	}
	tex.Memory = memory
	vk.BindImageMemory(ctx.Device, tex.Image, tex.Memory, 0)

	// Transition layout and copy
	if err := cmdPool.TransitionImageLayout(tex.Image, vk.ImageLayoutUndefined, vk.ImageLayoutTransferDstOptimal); err != nil {
		return nil, fmt.Errorf("failed to transition image for copy: %w", err)
	}

	if err := cmdPool.CopyBufferToImage(staging.Handle, tex.Image, tex.Width, tex.Height); err != nil {
		return nil, fmt.Errorf("failed to copy buffer to image: %w", err)
	}

	if err := cmdPool.TransitionImageLayout(tex.Image, vk.ImageLayoutTransferDstOptimal, vk.ImageLayoutShaderReadOnlyOptimal); err != nil {
		return nil, fmt.Errorf("failed to transition image for shader read: %w", err)
	}

	// Create image view
	viewInfo := &vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    tex.Image,
		ViewType: vk.ImageViewType2d,
		Format:   vk.FormatR8g8b8a8Srgb,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}

	var imageView vk.ImageView
	if res := vk.CreateImageView(ctx.Device, viewInfo, nil, &imageView); res != vk.Success {
		return nil, fmt.Errorf("failed to create texture image view: %d", res)
	}
	tex.ImageView = imageView

	// Create sampler
	samplerInfo := &vk.SamplerCreateInfo{
		SType:                   vk.StructureTypeSamplerCreateInfo,
		MagFilter:               vk.FilterNearest,
		MinFilter:               vk.FilterNearest,
		AddressModeU:            vk.SamplerAddressModeRepeat,
		AddressModeV:            vk.SamplerAddressModeRepeat,
		AddressModeW:            vk.SamplerAddressModeRepeat,
		AnisotropyEnable:        vk.False,
		BorderColor:             vk.BorderColorIntOpaqueBlack,
		UnnormalizedCoordinates: vk.False,
		CompareEnable:           vk.False,
		MipmapMode:              vk.SamplerMipmapModeNearest,
		MipLodBias:              0,
		MinLod:                  0,
		MaxLod:                  0,
	}

	var sampler vk.Sampler
	if res := vk.CreateSampler(ctx.Device, samplerInfo, nil, &sampler); res != vk.Success {
		return nil, fmt.Errorf("failed to create texture sampler: %d", res)
	}
	tex.Sampler = sampler

	return tex, nil
}

// Cleanup destroys all texture resources.
func (t *Texture) Cleanup() {
	if t.Sampler != nil {
		vk.DestroySampler(t.ctx.Device, t.Sampler, nil)
	}
	if t.ImageView != nil {
		vk.DestroyImageView(t.ctx.Device, t.ImageView, nil)
	}
	if t.Image != nil {
		vk.DestroyImage(t.ctx.Device, t.Image, nil)
	}
	if t.Memory != nil {
		vk.FreeMemory(t.ctx.Device, t.Memory, nil)
	}
}

// TextureAtlas holds a single texture atlas image for block textures.
// Block UVs index into sub-regions of this atlas.
type TextureAtlas struct {
	Texture    *Texture
	TileWidth  int
	TileHeight int
	Columns    int
	Rows       int
}

// NewTextureAtlas creates a texture atlas from raw RGBA data.
// tileWidth and tileHeight specify the size of each tile in pixels.
func NewTextureAtlas(
	ctx *VulkanContext,
	cmdPool *CommandPool,
	imageData []byte,
	atlasWidth, atlasHeight int,
	tileWidth, tileHeight int,
) (*TextureAtlas, error) {
	tex, err := CreateTexture(ctx, cmdPool, imageData, atlasWidth, atlasHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to create texture atlas: %w", err)
	}

	cols := atlasWidth / tileWidth
	rows := atlasHeight / tileHeight
	if cols == 0 || rows == 0 {
		tex.Cleanup()
		return nil, fmt.Errorf("atlas dimensions (%dx%d) too small for tile size (%dx%d)",
			atlasWidth, atlasHeight, tileWidth, tileHeight)
	}

	return &TextureAtlas{
		Texture:    tex,
		TileWidth:  tileWidth,
		TileHeight: tileHeight,
		Columns:    cols,
		Rows:       rows,
	}, nil
}

// TileUV returns the normalised UV coordinates (u0, v0, u1, v1) for the
// tile at the given column and row.
func (a *TextureAtlas) TileUV(col, row int) (u0, v0, u1, v1 float32) {
	atlasW := float32(a.Columns * a.TileWidth)
	atlasH := float32(a.Rows * a.TileHeight)
	u0 = float32(col*a.TileWidth) / atlasW
	v0 = float32(row*a.TileHeight) / atlasH
	u1 = float32((col+1)*a.TileWidth) / atlasW
	v1 = float32((row+1)*a.TileHeight) / atlasH
	return
}

// Cleanup destroys the underlying texture.
func (a *TextureAtlas) Cleanup() {
	if a.Texture != nil {
		a.Texture.Cleanup()
	}
}
