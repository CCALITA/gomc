#version 450

layout(set=0, binding=0) uniform UBO {
    mat4 viewProj;
} ubo;

layout(push_constant) uniform Push {
    vec4 offset;
} push;

layout(location=0) in vec3 inPosition;
layout(location=1) in vec3 inColor;

layout(location=0) out vec3 fragColor;

void main() {
    gl_Position = ubo.viewProj * vec4(inPosition + push.offset.xyz, 1.0);
    fragColor = inColor;
}
