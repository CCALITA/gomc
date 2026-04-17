#version 450

layout(set = 1, binding = 0) uniform sampler2D texAtlas;

layout(location = 0) in vec3 fragNormal;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in float fragAO;
layout(location = 3) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

const vec3 sunDir = normalize(vec3(0.4, 1.0, 0.3));
const vec3 sunColor = vec3(1.0, 0.98, 0.92);
const vec3 ambientColor = vec3(0.3, 0.35, 0.45);
const float fogStart = 128.0;
const float fogEnd = 256.0;
const vec3 fogColor = vec3(0.6, 0.75, 0.95);

void main() {
    vec4 texColor = texture(texAtlas, fragTexCoord);
    if (texColor.a < 0.5) discard;

    vec3 normal = normalize(fragNormal);
    float diffuse = max(dot(normal, sunDir), 0.0);
    vec3 lighting = ambientColor + sunColor * diffuse;
    lighting *= fragAO;

    vec3 color = texColor.rgb * lighting;

    float dist = length(fragWorldPos);
    float fogFactor = clamp((dist - fogStart) / (fogEnd - fogStart), 0.0, 1.0);
    color = mix(color, fogColor, fogFactor);

    outColor = vec4(color, 1.0);
}
