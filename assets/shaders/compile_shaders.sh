#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Compiling shaders to SPIR-V..."

for shader in "$SCRIPT_DIR"/*.vert "$SCRIPT_DIR"/*.frag; do
    [ -f "$shader" ] || continue
    out="${shader}.spv"
    echo "  $(basename "$shader") -> $(basename "$out")"
    glslc "$shader" -o "$out"
done

echo "Done."
