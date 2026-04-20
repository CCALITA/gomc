package block

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Door / Trapdoor IDs and properties
// ---------------------------------------------------------------------------

func TestDoorBlockIDs(t *testing.T) {
	assert.Equal(t, BlockID(81), OakDoor)
	assert.Equal(t, BlockID(82), IronDoor)
	assert.Equal(t, BlockID(83), OakTrapdoor)
}

func TestDoorProperties(t *testing.T) {
	tests := []struct {
		id       BlockID
		name     string
		hardness float32
		isDoor   bool
		isTrap   bool
	}{
		{OakDoor, "oak_door", 3.0, true, false},
		{IronDoor, "iron_door", 5.0, true, false},
		{OakTrapdoor, "oak_trapdoor", 3.0, false, true},
	}
	for _, tc := range tests {
		p := GetProperties(tc.id)
		assert.Equal(t, tc.name, p.Name)
		assert.Equal(t, tc.hardness, p.Hardness, "hardness for %s", tc.name)
		assert.True(t, p.Transparent, "door/trapdoor should be transparent: %s", tc.name)
		assert.Equal(t, tc.isDoor, p.IsDoor, "IsDoor for %s", tc.name)
		assert.Equal(t, tc.isTrap, p.IsTrapdoor, "IsTrapdoor for %s", tc.name)
	}
}

// ---------------------------------------------------------------------------
// Open-bit encoding
// ---------------------------------------------------------------------------

func TestDoorOpenBit_Toggle(t *testing.T) {
	closed := OakDoor
	assert.False(t, IsDoorOpen(closed), "new door should be closed")

	opened := ToggleDoorOpen(closed)
	assert.True(t, IsDoorOpen(opened), "toggled door should be open")
	assert.Equal(t, OakDoor, BaseID(opened), "base ID preserved after toggle")

	closedAgain := ToggleDoorOpen(opened)
	assert.False(t, IsDoorOpen(closedAgain), "double toggle returns to closed")
	assert.Equal(t, closed, closedAgain)
}

func TestDoorOpenBit_WithDoorOpen(t *testing.T) {
	id := IronDoor
	openID := WithDoorOpen(id, true)
	assert.True(t, IsDoorOpen(openID))
	assert.Equal(t, IronDoor, BaseID(openID))

	closedID := WithDoorOpen(openID, false)
	assert.False(t, IsDoorOpen(closedID))
	assert.Equal(t, IronDoor, closedID)
}

func TestTrapdoorOpenBit(t *testing.T) {
	closed := OakTrapdoor
	assert.False(t, IsDoorOpen(closed))

	opened := WithDoorOpen(closed, true)
	assert.True(t, IsDoorOpen(opened))
	assert.True(t, IsTrapdoorBlock(opened))
	assert.Equal(t, OakTrapdoor, BaseID(opened))
}

// ---------------------------------------------------------------------------
// IsSolid depends on open/closed state
// ---------------------------------------------------------------------------

func TestDoor_IsSolid_ClosedVsOpen(t *testing.T) {
	tests := []struct {
		name   string
		id     BlockID
	}{
		{"OakDoor", OakDoor},
		{"IronDoor", IronDoor},
		{"OakTrapdoor", OakTrapdoor},
	}
	for _, tc := range tests {
		t.Run(tc.name+"_closed", func(t *testing.T) {
			assert.True(t, IsSolid(tc.id), "closed %s should be solid", tc.name)
		})
		t.Run(tc.name+"_open", func(t *testing.T) {
			openID := WithDoorOpen(tc.id, true)
			assert.False(t, IsSolid(openID), "open %s should not be solid", tc.name)
		})
	}
}

// ---------------------------------------------------------------------------
// Classification helpers
// ---------------------------------------------------------------------------

func TestIsDoorBlock(t *testing.T) {
	assert.True(t, IsDoorBlock(OakDoor))
	assert.True(t, IsDoorBlock(IronDoor))
	assert.True(t, IsDoorBlock(WithDoorOpen(OakDoor, true)))
	assert.False(t, IsDoorBlock(OakTrapdoor))
	assert.False(t, IsDoorBlock(Stone))
}

func TestIsTrapdoorBlock(t *testing.T) {
	assert.True(t, IsTrapdoorBlock(OakTrapdoor))
	assert.True(t, IsTrapdoorBlock(WithDoorOpen(OakTrapdoor, true)))
	assert.False(t, IsTrapdoorBlock(OakDoor))
	assert.False(t, IsTrapdoorBlock(Stone))
}

func TestIsDoorOrTrapdoor(t *testing.T) {
	assert.True(t, IsDoorOrTrapdoor(OakDoor))
	assert.True(t, IsDoorOrTrapdoor(IronDoor))
	assert.True(t, IsDoorOrTrapdoor(OakTrapdoor))
	assert.True(t, IsDoorOrTrapdoor(WithDoorOpen(OakDoor, true)))
	assert.False(t, IsDoorOrTrapdoor(Stone))
	assert.False(t, IsDoorOrTrapdoor(Air))
}

// ---------------------------------------------------------------------------
// GetProperties falls back to base ID for open doors
// ---------------------------------------------------------------------------

func TestGetProperties_OpenDoor(t *testing.T) {
	openDoor := WithDoorOpen(OakDoor, true)
	p := GetProperties(openDoor)
	assert.Equal(t, "oak_door", p.Name, "open door should resolve to base properties")
	assert.Equal(t, float32(3.0), p.Hardness)
	assert.True(t, p.IsDoor)
}

func TestGetProperties_OpenTrapdoor(t *testing.T) {
	openTrap := WithDoorOpen(OakTrapdoor, true)
	p := GetProperties(openTrap)
	assert.Equal(t, "oak_trapdoor", p.Name)
	assert.True(t, p.IsTrapdoor)
}
