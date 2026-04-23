#version 450

layout(push_constant) uniform PushConstants {
    vec2 screenSize;
} push;

layout(location = 0) in vec2 inPosition;
layout(location = 1) in vec4 inColor;

layout(location = 0) out vec4 fragColor;

void main() {
    vec2 ndc = (inPosition / push.screenSize) * 2.0 - 1.0;
    gl_Position = vec4(ndc.x, ndc.y, 0.0, 1.0);
    fragColor = inColor;
}
