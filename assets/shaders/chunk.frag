#version 450

layout(location = 0) in vec3 fragNormal;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in float fragAO;
layout(location = 3) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

void main() {
    vec3 normal = normalize(fragNormal);

    // Simple directional light (sun from upper-right)
    vec3 sunDir = normalize(vec3(0.5, 1.0, 0.3));
    float diffuse = max(dot(normal, sunDir), 0.0);

    // Base color from normal direction (gives distinct colors per face)
    vec3 baseColor;
    if (abs(normal.y) > 0.5) {
        // Top/bottom faces — green for grass
        baseColor = normal.y > 0 ? vec3(0.35, 0.6, 0.2) : vec3(0.5, 0.4, 0.3);
    } else {
        // Side faces — brown for dirt/stone
        baseColor = vec3(0.55, 0.45, 0.35);
    }

    // Ambient + diffuse lighting with AO
    vec3 ambient = vec3(0.4);
    vec3 lighting = ambient + vec3(0.8) * diffuse;
    lighting *= fragAO;

    vec3 color = baseColor * lighting;

    // Distance fog toward sky blue
    vec3 skyColor = vec3(0.53, 0.81, 0.92);
    float dist = length(fragWorldPos);
    float fogFactor = clamp((dist - 128.0) / (256.0 - 128.0), 0.0, 1.0);
    color = mix(color, skyColor, fogFactor);

    outColor = vec4(color, 1.0);
}
