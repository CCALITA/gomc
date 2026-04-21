package render

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

// debugMode controls whether Vulkan validation layers are enabled.
// Set via build tags or linker flags in development builds.
var debugMode = false

// validationLayers lists the Vulkan validation layers to enable in debug mode.
var validationLayers = []string{
	"VK_LAYER_KHRONOS_validation\x00",
}

// deviceExtensions lists the required Vulkan device extensions.
var deviceExtensions = []string{
	"VK_KHR_swapchain\x00",
}

// QueueFamilyIndices holds the indices of the queue families needed by the renderer.
type QueueFamilyIndices struct {
	GraphicsFamily uint32
	PresentFamily  uint32
	HasGraphics    bool
	HasPresent     bool
}

// IsComplete returns true if both required queue families have been found.
func (q QueueFamilyIndices) IsComplete() bool {
	return q.HasGraphics && q.HasPresent
}

// VulkanContext holds the core Vulkan handles: instance, device, and queues.
type VulkanContext struct {
	Instance       vk.Instance
	Surface        vk.Surface
	PhysicalDevice vk.PhysicalDevice
	Device         vk.Device
	GraphicsQueue  vk.Queue
	PresentQueue   vk.Queue
	QueueIndices   QueueFamilyIndices
}

// Init initialises the Vulkan context using the given GLFW window.
// It creates an instance (with validation layers when debugMode is true),
// selects a physical device, and creates a logical device with graphics
// and present queues.
func (ctx *VulkanContext) Init(window *glfw.Window) error {
	// Initialise the Vulkan function loader via GLFW.
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())

	if err := vk.Init(); err != nil {
		return fmt.Errorf("initialise vulkan loader: %w", err)
	}

	// --- Create instance ---
	appInfo := &vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PApplicationName:   "GoMC\x00",
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "GoMC Engine\x00",
		EngineVersion:      vk.MakeVersion(1, 0, 0),
		ApiVersion:         vk.ApiVersion11,
	}

	extensions := window.GetRequiredInstanceExtensions()

	// macOS MoltenVK requires the portability enumeration extension.
	extensions = append(extensions, "VK_KHR_portability_enumeration\x00")

	createInfo := &vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        appInfo,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
		Flags:                   0x00000001, // VK_INSTANCE_CREATE_ENUMERATE_PORTABILITY_BIT_KHR
	}

	if debugMode {
		createInfo.EnabledLayerCount = uint32(len(validationLayers))
		createInfo.PpEnabledLayerNames = validationLayers
	}

	var instance vk.Instance
	if res := vk.CreateInstance(createInfo, nil, &instance); res != vk.Success {
		return fmt.Errorf("create vulkan instance: vulkan result %d", res)
	}
	ctx.Instance = instance
	vk.InitInstance(instance)

	// --- Create surface ---
	surfacePtr, err := window.CreateWindowSurface(instance, nil)
	if err != nil {
		return fmt.Errorf("create window surface: %w", err)
	}
	ctx.Surface = vk.SurfaceFromPointer(surfacePtr)

	// --- Pick physical device ---
	if err := ctx.pickPhysicalDevice(); err != nil {
		return fmt.Errorf("pick physical device: %w", err)
	}

	// --- Create logical device ---
	if err := ctx.createLogicalDevice(); err != nil {
		return fmt.Errorf("create logical device: %w", err)
	}

	return nil
}

// pickPhysicalDevice selects the first suitable GPU.
func (ctx *VulkanContext) pickPhysicalDevice() error {
	var deviceCount uint32
	if res := vk.EnumeratePhysicalDevices(ctx.Instance, &deviceCount, nil); res != vk.Success {
		return fmt.Errorf("enumerate physical devices count: vulkan result %d", res)
	}
	if deviceCount == 0 {
		return fmt.Errorf("no GPUs with Vulkan support found")
	}

	devices := make([]vk.PhysicalDevice, deviceCount)
	if res := vk.EnumeratePhysicalDevices(ctx.Instance, &deviceCount, devices); res != vk.Success {
		return fmt.Errorf("enumerate physical devices: vulkan result %d", res)
	}

	for _, device := range devices {
		if ctx.isDeviceSuitable(device) {
			ctx.PhysicalDevice = device
			return nil
		}
	}

	return fmt.Errorf("no suitable GPU found")
}

// isDeviceSuitable checks whether the physical device supports the
// required queue families and extensions.
func (ctx *VulkanContext) isDeviceSuitable(device vk.PhysicalDevice) bool {
	indices := ctx.findQueueFamilies(device)
	if !indices.IsComplete() {
		return false
	}

	if !checkDeviceExtensionSupport(device) {
		return false
	}

	details, err := querySwapchainSupport(device, ctx.Surface)
	if err != nil {
		return false
	}
	return len(details.Formats) > 0 && len(details.PresentModes) > 0
}

// findQueueFamilies locates the graphics and present queue families
// for the given physical device.
func (ctx *VulkanContext) findQueueFamilies(device vk.PhysicalDevice) QueueFamilyIndices {
	var indices QueueFamilyIndices

	var queueFamilyCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, nil)

	families := make([]vk.QueueFamilyProperties, queueFamilyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, families)

	for i, family := range families {
		family.Deref()

		if family.QueueFlags&vk.QueueFlags(vk.QueueGraphicsBit) != 0 {
			indices.GraphicsFamily = uint32(i)
			indices.HasGraphics = true
		}

		var presentSupport vk.Bool32
		if res := vk.GetPhysicalDeviceSurfaceSupport(device, uint32(i), ctx.Surface, &presentSupport); res != vk.Success {
			continue
		}
		if presentSupport != 0 {
			indices.PresentFamily = uint32(i)
			indices.HasPresent = true
		}

		if indices.IsComplete() {
			break
		}
	}

	return indices
}

// createLogicalDevice creates the logical device and retrieves queue handles.
func (ctx *VulkanContext) createLogicalDevice() error {
	ctx.QueueIndices = ctx.findQueueFamilies(ctx.PhysicalDevice)

	uniqueQueueFamilies := map[uint32]bool{
		ctx.QueueIndices.GraphicsFamily: true,
		ctx.QueueIndices.PresentFamily:  true,
	}

	queuePriority := []float32{1.0}
	var queueCreateInfos []vk.DeviceQueueCreateInfo
	for family := range uniqueQueueFamilies {
		queueCreateInfos = append(queueCreateInfos, vk.DeviceQueueCreateInfo{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: family,
			QueueCount:       1,
			PQueuePriorities: queuePriority,
		})
	}

	deviceFeatures := vk.PhysicalDeviceFeatures{}

	createInfo := &vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(queueCreateInfos)),
		PQueueCreateInfos:       queueCreateInfos,
		PEnabledFeatures:        []vk.PhysicalDeviceFeatures{deviceFeatures},
		EnabledExtensionCount:   uint32(len(deviceExtensions)),
		PpEnabledExtensionNames: deviceExtensions,
	}

	if debugMode {
		createInfo.EnabledLayerCount = uint32(len(validationLayers))
		createInfo.PpEnabledLayerNames = validationLayers
	}

	var device vk.Device
	if res := vk.CreateDevice(ctx.PhysicalDevice, createInfo, nil, &device); res != vk.Success {
		return fmt.Errorf("create logical device: vulkan result %d", res)
	}
	ctx.Device = device

	var graphicsQueue vk.Queue
	vk.GetDeviceQueue(ctx.Device, ctx.QueueIndices.GraphicsFamily, 0, &graphicsQueue)
	ctx.GraphicsQueue = graphicsQueue

	var presentQueue vk.Queue
	vk.GetDeviceQueue(ctx.Device, ctx.QueueIndices.PresentFamily, 0, &presentQueue)
	ctx.PresentQueue = presentQueue

	return nil
}

// checkDeviceExtensionSupport verifies that the device provides all
// required extensions.
func checkDeviceExtensionSupport(device vk.PhysicalDevice) bool {
	var extensionCount uint32
	if res := vk.EnumerateDeviceExtensionProperties(device, "", &extensionCount, nil); res != vk.Success {
		return false
	}

	available := make([]vk.ExtensionProperties, extensionCount)
	if res := vk.EnumerateDeviceExtensionProperties(device, "", &extensionCount, available); res != vk.Success {
		return false
	}

	requiredSet := make(map[string]bool)
	for _, ext := range deviceExtensions {
		requiredSet[vk.ToString([]byte(ext))] = true
	}

	for _, ext := range available {
		ext.Deref()
		name := vk.ToString(ext.ExtensionName[:])
		delete(requiredSet, name)
	}

	return len(requiredSet) == 0
}

// SwapchainSupportDetails holds the surface capabilities, formats, and
// present modes for a physical device.
type SwapchainSupportDetails struct {
	Capabilities vk.SurfaceCapabilities
	Formats      []vk.SurfaceFormat
	PresentModes []vk.PresentMode
}

// querySwapchainSupport queries the swapchain support details for the
// given physical device and surface.
func querySwapchainSupport(device vk.PhysicalDevice, surface vk.Surface) (SwapchainSupportDetails, error) {
	var details SwapchainSupportDetails

	if res := vk.GetPhysicalDeviceSurfaceCapabilities(device, surface, &details.Capabilities); res != vk.Success {
		return details, fmt.Errorf("get surface capabilities: vulkan result %d", res)
	}
	details.Capabilities.Deref()

	var formatCount uint32
	if res := vk.GetPhysicalDeviceSurfaceFormats(device, surface, &formatCount, nil); res != vk.Success {
		return details, fmt.Errorf("get surface format count: vulkan result %d", res)
	}
	if formatCount > 0 {
		details.Formats = make([]vk.SurfaceFormat, formatCount)
		if res := vk.GetPhysicalDeviceSurfaceFormats(device, surface, &formatCount, details.Formats); res != vk.Success {
			return details, fmt.Errorf("get surface formats: vulkan result %d", res)
		}
		for i := range details.Formats {
			details.Formats[i].Deref()
		}
	}

	var presentModeCount uint32
	if res := vk.GetPhysicalDeviceSurfacePresentModes(device, surface, &presentModeCount, nil); res != vk.Success {
		return details, fmt.Errorf("get present mode count: vulkan result %d", res)
	}
	if presentModeCount > 0 {
		details.PresentModes = make([]vk.PresentMode, presentModeCount)
		if res := vk.GetPhysicalDeviceSurfacePresentModes(device, surface, &presentModeCount, details.PresentModes); res != vk.Success {
			return details, fmt.Errorf("get present modes: vulkan result %d", res)
		}
	}

	return details, nil
}

// findMemoryType finds a memory type index on the physical device that
// matches the given type filter and property flags.
func findMemoryType(physDevice vk.PhysicalDevice, typeFilter uint32, properties vk.MemoryPropertyFlags) (uint32, error) {
	var memProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(physDevice, &memProperties)
	memProperties.Deref()

	for i := uint32(0); i < memProperties.MemoryTypeCount; i++ {
		memProperties.MemoryTypes[i].Deref()
		if typeFilter&(1<<i) != 0 &&
			vk.MemoryPropertyFlags(memProperties.MemoryTypes[i].PropertyFlags)&properties == properties {
			return i, nil
		}
	}

	return 0, fmt.Errorf("find suitable memory type: no match for filter %d with properties %d", typeFilter, properties)
}

// Cleanup destroys the Vulkan context resources in reverse order.
func (ctx *VulkanContext) Cleanup() {
	if ctx.Device != nil {
		vk.DeviceWaitIdle(ctx.Device)
		vk.DestroyDevice(ctx.Device, nil)
	}
	if ctx.Surface != vk.NullSurface {
		vk.DestroySurface(ctx.Instance, ctx.Surface, nil)
	}
	if ctx.Instance != nil {
		vk.DestroyInstance(ctx.Instance, nil)
	}
}

// sliceUint32 reinterprets a []byte as []uint32 for SPIR-V loading.
// The byte slice length must be a multiple of 4.
func sliceUint32(data []byte) []uint32 {
	if len(data) == 0 {
		return nil
	}
	return unsafe.Slice((*uint32)(unsafe.Pointer(&data[0])), len(data)/4)
}
