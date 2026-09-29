uniform sampler2D uNoiseMap;
uniform sampler2D uNoiseMapHigh;
uniform float uHexRadius;
uniform bool uCircularEdge;
uniform float uGrassOverflow;
uniform float uBandWidth;
uniform float uWarpAmount;
uniform float uNoiseScale;
uniform float uFadeProgress;
varying vec2 vLocalPos;
varying vec2 vTileOffset;

float hexSDF(vec2 p, float circumR) {
    float apothem = circumR * 0.866025404;
    float d1 = abs(p.x);
    float d2 = abs(0.5 * p.x + 0.866025404 * p.y);
    float d3 = abs(-0.5 * p.x + 0.866025404 * p.y);
    return max(max(d1, d2), d3) - apothem;
}
//#pragma body
#include <alphamap_fragment>

float d = uCircularEdge ? length(vLocalPos) - uHexRadius : hexSDF(vLocalPos, uHexRadius);
float edgeDist = -d;

vec2 noiseUV1 = vLocalPos * uNoiseScale + vTileOffset;
vec2 noiseUV2 = vLocalPos * uNoiseScale * 2.5 + vTileOffset * 1.7;
float n1 = texture2D(uNoiseMap, noiseUV1).r * 2.0 - 1.0;
float n2 = texture2D(uNoiseMapHigh, noiseUV2).r * 2.0 - 1.0;
float n = n1 * 0.7 + n2 * 0.3;

float edgeDistWarped = edgeDist + uGrassOverflow + n * uWarpAmount;
float t = smoothstep(0.0, uBandWidth, edgeDistWarped);
t = pow(t, 1.1);

diffuseColor.a *= t;
diffuseColor.a *= uFadeProgress;
