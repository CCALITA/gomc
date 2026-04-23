#version 450

layout(set = 0, binding = 1) uniform sampler2D texAtlas;

layout(location = 0) in vec3 fragNormal;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in float fragAO;
layout(location = 3) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

void main() {
    vec3 normal = normalize(fragNormal);

    // fragTexCoord.xy holds the tile origin in the atlas.
    // Derive per-block local UV from world position based on face direction.
    vec2 tileOrigin = fragTexCoord;
    float tileSize = 1.0 / 16.0;

    vec2 localUV;
    if (abs(normal.y) > 0.5) {
        // Top/bottom face: tile using XZ
        localUV = fract(fragWorldPos.xz);
    } else if (abs(normal.x) > 0.5) {
        // East/west face: tile using ZY
        localUV = fract(fragWorldPos.zy);
    } else {
        // North/south face: tile using XY
        localUV = fract(fragWorldPos.xy);
    }

    vec2 atlasUV = tileOrigin + localUV * tileSize;
    vec4 texColor = texture(texAtlas, atlasUV);

    // Sun lighting
    vec3 sunDir = normalize(vec3(0.5, 1.0, 0.3));
    float diffuse = max(dot(normal, sunDir), 0.0);

    // Ambient + diffuse with face shading
    vec3 ambient = vec3(0.35);
    vec3 lighting = ambient + vec3(0.75) * diffuse;

    // Face-dependent brightness (sides slightly darker, bottom darkest)
    if (abs(normal.y) < 0.5) {
        lighting *= 0.82;
    }
    if (normal.y < -0.5) {
        lighting *= 0.65;
    }

    lighting *= fragAO;
    vec3 color = texColor.rgb * lighting;

    // Distance fog
    vec3 skyColor = vec3(0.53, 0.81, 0.92);
    float dist = length(fragWorldPos);
    float fogFactor = clamp((dist - 120.0) / (200.0 - 120.0), 0.0, 1.0);
    color = mix(color, skyColor, fogFactor);

    outColor = vec4(color, 1.0);
}
