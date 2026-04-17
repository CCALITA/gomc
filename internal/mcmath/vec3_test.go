package mcmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVec3_Add(t *testing.T) {
	v := Vec3{1, 2, 3}
	other := Vec3{4, 5, 6}
	result := v.Add(other)
	assert.Equal(t, Vec3{5, 7, 9}, result)
}

func TestVec3_Sub(t *testing.T) {
	v := Vec3{5, 7, 9}
	other := Vec3{1, 2, 3}
	result := v.Sub(other)
	assert.Equal(t, Vec3{4, 5, 6}, result)
}

func TestVec3_Mul(t *testing.T) {
	v := Vec3{2, 3, 4}
	other := Vec3{5, 6, 7}
	result := v.Mul(other)
	assert.Equal(t, Vec3{10, 18, 28}, result)
}

func TestVec3_Scale(t *testing.T) {
	v := Vec3{1, 2, 3}
	result := v.Scale(2)
	assert.Equal(t, Vec3{2, 4, 6}, result)
}

func TestVec3_Dot(t *testing.T) {
	v := Vec3{1, 2, 3}
	other := Vec3{4, 5, 6}
	assert.InDelta(t, float32(32), v.Dot(other), 1e-5)
}

func TestVec3_Cross(t *testing.T) {
	v := Vec3{1, 0, 0}
	other := Vec3{0, 1, 0}
	result := v.Cross(other)
	assert.Equal(t, Vec3{0, 0, 1}, result)

	// Test anti-commutativity
	reverse := other.Cross(v)
	assert.Equal(t, Vec3{0, 0, -1}, reverse)
}

func TestVec3_LengthSq(t *testing.T) {
	v := Vec3{3, 4, 0}
	assert.InDelta(t, float32(25), v.LengthSq(), 1e-5)
}

func TestVec3_Length(t *testing.T) {
	v := Vec3{3, 4, 0}
	assert.InDelta(t, float32(5), v.Length(), 1e-5)
}

func TestVec3_Normalize(t *testing.T) {
	v := Vec3{3, 0, 0}
	n := v.Normalize()
	assert.InDelta(t, float32(1), n.X, 1e-5)
	assert.InDelta(t, float32(0), n.Y, 1e-5)
	assert.InDelta(t, float32(0), n.Z, 1e-5)
}

func TestVec3_Normalize_Zero(t *testing.T) {
	v := Vec3{0, 0, 0}
	n := v.Normalize()
	assert.Equal(t, Vec3{}, n)
}

func TestVec3_Normalize_Unit(t *testing.T) {
	v := Vec3{1, 2, 3}
	n := v.Normalize()
	assert.InDelta(t, float32(1), n.Length(), 1e-5)
}

func TestVec3_Lerp(t *testing.T) {
	a := Vec3{0, 0, 0}
	b := Vec3{10, 20, 30}

	mid := a.Lerp(b, 0.5)
	assert.InDelta(t, float32(5), mid.X, 1e-5)
	assert.InDelta(t, float32(10), mid.Y, 1e-5)
	assert.InDelta(t, float32(15), mid.Z, 1e-5)

	start := a.Lerp(b, 0)
	assert.Equal(t, a, start)

	end := a.Lerp(b, 1)
	assert.Equal(t, b, end)
}

func TestVec3_Distance(t *testing.T) {
	a := Vec3{0, 0, 0}
	b := Vec3{3, 4, 0}
	assert.InDelta(t, float32(5), a.Distance(b), 1e-5)
}

func TestVec3_DistanceSq(t *testing.T) {
	a := Vec3{0, 0, 0}
	b := Vec3{3, 4, 0}
	assert.InDelta(t, float32(25), a.DistanceSq(b), 1e-5)
}

func TestVec3_Floor(t *testing.T) {
	tests := []struct {
		name string
		v    Vec3
		want BlockPos
	}{
		{"positive", Vec3{1.5, 2.7, 3.1}, BlockPos{1, 2, 3}},
		{"negative", Vec3{-0.5, -1.5, -2.5}, BlockPos{-1, -2, -3}},
		{"integer", Vec3{1, 2, 3}, BlockPos{1, 2, 3}},
		{"zero", Vec3{0, 0, 0}, BlockPos{0, 0, 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.v.Floor())
		})
	}
}

func TestVec3_String(t *testing.T) {
	v := Vec3{1.5, 2.5, 3.5}
	s := v.String()
	assert.Contains(t, s, "Vec3")
	assert.Contains(t, s, "1.5000")
}

func TestVec3_Operations_Immutable(t *testing.T) {
	original := Vec3{1, 2, 3}
	other := Vec3{4, 5, 6}

	original.Add(other)
	assert.Equal(t, Vec3{1, 2, 3}, original, "Add should not mutate the original")

	original.Sub(other)
	assert.Equal(t, Vec3{1, 2, 3}, original, "Sub should not mutate the original")

	original.Scale(2)
	assert.Equal(t, Vec3{1, 2, 3}, original, "Scale should not mutate the original")
}
