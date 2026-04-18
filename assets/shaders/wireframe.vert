#version 450

layout(push_constant) uniform PushConstants {
    mat4 mvp;
} push;

// 24 vertices defining the 12 edges of a unit cube [0,1]^3.
const vec3 positions[24] = vec3[24](
    // Bottom face edges
    vec3(0, 0, 0), vec3(1, 0, 0),
    vec3(1, 0, 0), vec3(1, 1, 0),
    vec3(1, 1, 0), vec3(0, 1, 0),
    vec3(0, 1, 0), vec3(0, 0, 0),
    // Top face edges
    vec3(0, 0, 1), vec3(1, 0, 1),
    vec3(1, 0, 1), vec3(1, 1, 1),
    vec3(1, 1, 1), vec3(0, 1, 1),
    vec3(0, 1, 1), vec3(0, 0, 1),
    // Vertical edges
    vec3(0, 0, 0), vec3(0, 0, 1),
    vec3(1, 0, 0), vec3(1, 0, 1),
    vec3(1, 1, 0), vec3(1, 1, 1),
    vec3(0, 1, 0), vec3(0, 1, 1)
);

void main() {
    gl_Position = push.mvp * vec4(positions[gl_VertexIndex], 1.0);
}
