package block

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
)

func TestJukeboxInsertDisc(t *testing.T) {
	mgr := NewJukeboxManager()
	pos := mcmath.BlockPos{X: 1, Y: 2, Z: 3}

	ok := mgr.InsertDisc(pos, 501)
	assert.True(t, ok)
	assert.True(t, mgr.IsPlaying(pos))
}

func TestJukeboxEjectDisc(t *testing.T) {
	mgr := NewJukeboxManager()
	pos := mcmath.BlockPos{X: 1, Y: 2, Z: 3}

	mgr.InsertDisc(pos, 502)
	discID := mgr.EjectDisc(pos)

	assert.Equal(t, uint16(502), discID)
	assert.False(t, mgr.IsPlaying(pos))
}

func TestJukeboxEjectEmpty(t *testing.T) {
	mgr := NewJukeboxManager()
	pos := mcmath.BlockPos{X: 1, Y: 2, Z: 3}

	discID := mgr.EjectDisc(pos)
	assert.Equal(t, uint16(0), discID)
}

func TestJukeboxNoDoubleInsert(t *testing.T) {
	mgr := NewJukeboxManager()
	pos := mcmath.BlockPos{X: 1, Y: 2, Z: 3}

	ok1 := mgr.InsertDisc(pos, 500)
	ok2 := mgr.InsertDisc(pos, 501)

	assert.True(t, ok1)
	assert.False(t, ok2, "should not allow double insert")

	// The original disc should still be there.
	discID := mgr.EjectDisc(pos)
	assert.Equal(t, uint16(500), discID)
}

func TestJukeboxIsPlayingDefault(t *testing.T) {
	mgr := NewJukeboxManager()
	pos := mcmath.BlockPos{X: 10, Y: 20, Z: 30}

	assert.False(t, mgr.IsPlaying(pos))
}

func TestGetDiscTrackName(t *testing.T) {
	tests := []struct {
		discID uint16
		want   string
	}{
		{500, "music/13"},
		{501, "music/cat"},
		{502, "music/blocks"},
		{503, "music/chirp"},
		{504, "music/far"},
		{999, ""},
	}
	for _, tc := range tests {
		got := GetDiscTrackName(tc.discID)
		assert.Equal(t, tc.want, got, "GetDiscTrackName(%d)", tc.discID)
	}
}

func TestIsDisc(t *testing.T) {
	assert.True(t, IsDisc(500))
	assert.True(t, IsDisc(504))
	assert.False(t, IsDisc(100))
	assert.False(t, IsDisc(0))
}

func TestJukeboxProperties(t *testing.T) {
	p := GetProperties(Jukebox)
	assert.Equal(t, "jukebox", p.Name)
	assert.True(t, p.Solid)
	assert.Equal(t, float32(2.0), p.Hardness)
	assert.Equal(t, float32(6), p.BlastResistance)
}

func TestIsJukebox(t *testing.T) {
	assert.True(t, IsJukebox(Jukebox))
	assert.False(t, IsJukebox(Stone))
	assert.False(t, IsJukebox(Air))
}
