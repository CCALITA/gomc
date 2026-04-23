#!/bin/bash
# GoMC launcher — compiles shaders, builds, and runs with correct Vulkan/MoltenVK settings.

set -e

MOLTEN_VK_PREFIX="$(brew --prefix molten-vk 2>/dev/null || echo /opt/homebrew/opt/molten-vk)"
VULKAN_ICD="$MOLTEN_VK_PREFIX/../../etc/vulkan/icd.d/MoltenVK_icd.json"
if [ ! -f "$VULKAN_ICD" ]; then
    VULKAN_ICD="/opt/homebrew/etc/vulkan/icd.d/MoltenVK_icd.json"
fi

# Compile shaders
echo "Compiling shaders..."
for shader in assets/shaders/*.vert assets/shaders/*.frag; do
    [ -f "$shader" ] || continue
    spv="${shader}.spv"
    if [ ! -f "$spv" ] || [ "$shader" -nt "$spv" ]; then
        glslc "$shader" -o "$spv" && echo "  $shader → $spv"
    fi
done

echo "Building..."
CGO_LDFLAGS="-L$MOLTEN_VK_PREFIX/lib" GODEBUG=cgocheck=0 go build -o gomc ./cmd/client

echo "Running GoMC..."
exec env \
    VK_ICD_FILENAMES="$VULKAN_ICD" \
    DYLD_FALLBACK_LIBRARY_PATH="$MOLTEN_VK_PREFIX/lib:/opt/homebrew/lib" \
    MVK_CONFIG_USE_METAL_ARGUMENT_BUFFERS=0 \
    GODEBUG=cgocheck=0 \
    ./gomc "$@"
