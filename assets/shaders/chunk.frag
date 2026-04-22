#version 450

layout(location = 0) in vec3 fragNormal;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in float fragAO;
layout(location = 3) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

void main() {
    vec3 normal = normalize(fragNormal);
    float y = fragWorldPos.y;

    // Sun lighting
    vec3 sunDir = normalize(vec3(0.5, 1.0, 0.3));
    float diffuse = max(dot(normal, sunDir), 0.0);

    // Block color based on world height (mimics Minecraft layers)
    vec3 baseColor;
    bool isTop = normal.y > 0.5;
    bool isBottom = normal.y < -0.5;

    if (y < 5.0) {
        // Bedrock layer — dark gray
        baseColor = vec3(0.2, 0.2, 0.2);
    } else if (y < 12.0) {
        // Deep stone with lava tint
        baseColor = vec3(0.45, 0.42, 0.40);
    } else if (y < 55.0) {
        // Stone layer — gray
        baseColor = isTop ? vec3(0.55, 0.55, 0.53) : vec3(0.50, 0.50, 0.48);
    } else if (y < 62.0) {
        // Near surface — dirt
        baseColor = isTop ? vec3(0.52, 0.38, 0.26) : vec3(0.48, 0.35, 0.24);
    } else if (y < 64.0) {
        // Beach/shore — sand
        baseColor = vec3(0.82, 0.77, 0.55);
    } else if (y < 66.0) {
        // Surface — grass top, dirt sides
        if (isTop) {
            baseColor = vec3(0.33, 0.60, 0.18);
        } else if (isBottom) {
            baseColor = vec3(0.48, 0.35, 0.24);
        } else {
            // Side: grass-dirt gradient (green strip on top edge of side)
            float sideFrac = fract(fragWorldPos.y);
            vec3 grassSide = vec3(0.30, 0.50, 0.15);
            vec3 dirtSide = vec3(0.52, 0.38, 0.26);
            baseColor = mix(dirtSide, grassSide, smoothstep(0.7, 1.0, sideFrac));
        }
    } else if (y < 80.0) {
        // Hills — grass gets lighter at altitude
        float t = (y - 66.0) / 14.0;
        vec3 lowGrass = vec3(0.33, 0.60, 0.18);
        vec3 highGrass = vec3(0.40, 0.55, 0.25);
        baseColor = isTop ? mix(lowGrass, highGrass, t) : vec3(0.52, 0.38, 0.26);
    } else if (y < 100.0) {
        // Mountains — stone with some dirt
        float t = (y - 80.0) / 20.0;
        baseColor = mix(vec3(0.50, 0.45, 0.35), vec3(0.60, 0.58, 0.55), t);
    } else {
        // High peaks — snow-capped
        baseColor = isTop ? vec3(0.92, 0.93, 0.95) : vec3(0.65, 0.63, 0.60);
    }

    // Add subtle variation using world position hash
    float hash = fract(sin(dot(floor(fragWorldPos), vec3(12.9898, 78.233, 45.164))) * 43758.5453);
    baseColor *= 0.95 + 0.10 * hash;

    // Lighting: ambient + diffuse + AO
    vec3 ambient = vec3(0.35);
    vec3 lighting = ambient + vec3(0.75) * diffuse;

    // Face-dependent shading (sides slightly darker, bottom darkest)
    if (abs(normal.y) < 0.5) {
        lighting *= 0.85; // sides
    }
    if (normal.y < -0.5) {
        lighting *= 0.7; // bottom
    }

    lighting *= fragAO;
    vec3 color = baseColor * lighting;

    // Distance fog
    vec3 skyColor = vec3(0.53, 0.81, 0.92);
    float dist = length(fragWorldPos);
    float fogFactor = clamp((dist - 120.0) / (200.0 - 120.0), 0.0, 1.0);
    color = mix(color, skyColor, fogFactor);

    outColor = vec4(color, 1.0);
}
