#version 450

layout(set = 0, binding = 1) uniform sampler2D texAtlas;

layout(location = 0) in vec3 fragNormal;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in float fragAO;
layout(location = 3) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

void main() {
    vec3 normal = normalize(fragNormal);

    // Sample atlas tile — tileOrigin is in fragTexCoord
    vec2 tileOrigin = fragTexCoord;
    float tileSize = 1.0 / 16.0;

    vec2 localUV;
    if (abs(normal.y) > 0.5) {
        localUV = fract(fragWorldPos.xz);
    } else if (abs(normal.x) > 0.5) {
        localUV = fract(fragWorldPos.zy);
    } else {
        localUV = fract(fragWorldPos.xy);
    }

    vec2 atlasUV = tileOrigin + localUV * tileSize;
    vec4 texColor = texture(texAtlas, atlasUV);

    // If atlas tile is empty/transparent, use height-based fallback color
    if (texColor.a < 0.1) {
        float y = fragWorldPos.y;
        bool isTop = normal.y > 0.5;
        if (y < 55.0) {
            texColor = vec4(0.50, 0.50, 0.48, 1.0); // stone
        } else if (y < 64.0) {
            texColor = vec4(0.52, 0.38, 0.26, 1.0); // dirt
        } else {
            texColor = isTop ? vec4(0.33, 0.60, 0.18, 1.0) : vec4(0.52, 0.38, 0.26, 1.0);
        }
    }

    // Lighting
    vec3 sunDir = normalize(vec3(0.5, 1.0, 0.3));
    float diffuse = max(dot(normal, sunDir), 0.0);
    vec3 lighting = vec3(0.35) + vec3(0.75) * diffuse;

    if (abs(normal.y) < 0.5) lighting *= 0.82;
    if (normal.y < -0.5) lighting *= 0.65;
    lighting *= fragAO;

    vec3 color = texColor.rgb * lighting;

    vec3 skyColor = vec3(0.53, 0.81, 0.92);
    float dist = length(fragWorldPos);
    float fogFactor = clamp((dist - 120.0) / (200.0 - 120.0), 0.0, 1.0);
    color = mix(color, skyColor, fogFactor);

    outColor = vec4(color, 1.0);
}
