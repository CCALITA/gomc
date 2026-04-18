#version 450

layout(set = 0, binding = 1) uniform TimeOfDay {
    vec3 sunDir;
    vec3 ambientColor;
    vec3 skyColor;
    float ambientLevel;
} tod;

layout(set = 1, binding = 0) uniform sampler2D texAtlas;

layout(location = 0) in vec3 fragNormal;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in float fragAO;
layout(location = 3) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

const vec3 sunColor = vec3(1.0, 0.98, 0.92);
const float fogStart = 128.0;
const float fogEnd = 256.0;

void main() {
    vec4 texColor = texture(texAtlas, fragTexCoord);
    if (texColor.a < 0.5) discard;

    vec3 normal = normalize(fragNormal);
    float diffuse = max(dot(normal, tod.sunDir), 0.0);
    vec3 lighting = tod.ambientColor + sunColor * diffuse;
    lighting *= fragAO * tod.ambientLevel;

    vec3 color = texColor.rgb * lighting;

    float dist = length(fragWorldPos);
    float fogFactor = clamp((dist - fogStart) / (fogEnd - fogStart), 0.0, 1.0);
    color = mix(color, tod.skyColor, fogFactor);

    outColor = vec4(color, 1.0);
}
