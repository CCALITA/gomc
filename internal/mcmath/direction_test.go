package mcmath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirection_Opposite(t *testing.T) {
	assert.Equal(t, South, North.Opposite())
	assert.Equal(t, North, South.Opposite())
	assert.Equal(t, West, East.Opposite())
	assert.Equal(t, East, West.Opposite())
	assert.Equal(t, Down, Up.Opposite())
	assert.Equal(t, Up, Down.Opposite())
}

func TestDirection_Opposite_Involution(t *testing.T) {
	for _, d := range All() {
		assert.Equal(t, d, d.Opposite().Opposite(), "%s.Opposite().Opposite() should be itself", d)
	}
}

func TestDirection_Normal(t *testing.T) {
	assert.Equal(t, Vec3{0, 0, -1}, North.Normal())
	assert.Equal(t, Vec3{0, 0, 1}, South.Normal())
	assert.Equal(t, Vec3{1, 0, 0}, East.Normal())
	assert.Equal(t, Vec3{-1, 0, 0}, West.Normal())
	assert.Equal(t, Vec3{0, 1, 0}, Up.Normal())
	assert.Equal(t, Vec3{0, -1, 0}, Down.Normal())
}

func TestDirection_Normal_UnitLength(t *testing.T) {
	for _, d := range All() {
		n := d.Normal()
		assert.InDelta(t, float32(1), n.Length(), 1e-5, "%s normal should have unit length", d)
	}
}

func TestDirection_String(t *testing.T) {
	assert.Equal(t, "North", North.String())
	assert.Equal(t, "South", South.String())
	assert.Equal(t, "East", East.String())
	assert.Equal(t, "West", West.String())
	assert.Equal(t, "Up", Up.String())
	assert.Equal(t, "Down", Down.String())
}

func TestDirection_String_Invalid(t *testing.T) {
	d := Direction(99)
	s := d.String()
	assert.Contains(t, s, "Direction(99)")
}

func TestAll(t *testing.T) {
	dirs := All()
	assert.Len(t, dirs, 6)
	assert.Contains(t, dirs, North)
	assert.Contains(t, dirs, South)
	assert.Contains(t, dirs, East)
	assert.Contains(t, dirs, West)
	assert.Contains(t, dirs, Up)
	assert.Contains(t, dirs, Down)
}

func TestDirection_Normal_Opposite_Negation(t *testing.T) {
	for _, d := range All() {
		n := d.Normal()
		opp := d.Opposite().Normal()
		sum := n.Add(opp)
		assert.InDelta(t, float32(0), sum.X, 1e-5)
		assert.InDelta(t, float32(0), sum.Y, 1e-5)
		assert.InDelta(t, float32(0), sum.Z, 1e-5)
	}
}
