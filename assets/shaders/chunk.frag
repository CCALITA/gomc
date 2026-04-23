#version 450

layout(set = 0, binding = 1) uniform sampler2D texAtlas;

layout(location = 0) in vec3 fragNormal;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in float fragAO;
layout(location = 3) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

void main() {
    // Sample block texture from atlas
    vec4 texColor = texture(texAtlas, fragTexCoord);

    vec3 normal = normalize(fragNormal);

    // Sun lighting
    vec3 sunDir = normalize(vec3(0.5, 1.0, 0.3));
    float diffuse = max(dot(normal, sunDir), 0.0);

    // Ambient + diffuse with face shading
    vec3 ambient = vec3(0.4);
    vec3 lighting = ambient + vec3(0.7) * diffuse;

    // Face-dependent brightness
    if (abs(normal.y) < 0.5) {
        lighting *= 0.85;
    }
    if (normal.y < -0.5) {
        lighting *= 0.7;
    }

    lighting *= fragAO;
    vec3 color = texColor.rgb * lighting;

    // Distance fog
    vec3 skyColor = vec3(0.53, 0.81, 0.92);
    float dist = length(fragWorldPos);
    float fogFactor = clamp((dist - 120.0) / (200.0 - 120.0), 0.0, 1.0);
    color = mix(color, skyColor, fogFactor);

    outColor = vec4(color, texColor.a);
}
