package entity

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeGrid builds an IsSolid function from a set of solid positions.
func makeGrid(solid map[mcmath.BlockPos]bool) func(mcmath.BlockPos) bool {
	return func(pos mcmath.BlockPos) bool {
		return solid[pos]
	}
}

// flatFloor returns a solid function where every block at y == floorY is solid
// and everything else is air, except for additional walls.
func flatFloor(floorY int32, walls map[mcmath.BlockPos]bool) func(mcmath.BlockPos) bool {
	return func(pos mcmath.BlockPos) bool {
		if walls[pos] {
			return true
		}
		return pos.Y == floorY
	}
}

func TestFindPath(t *testing.T) {
	tests := []struct {
		name     string
		isSolid  func(mcmath.BlockPos) bool
		start    mcmath.BlockPos
		goal     mcmath.BlockPos
		maxNodes int
		wantNil  bool
		wantLen  int // expected path length; 0 means don't check length
		check    func(t *testing.T, path []mcmath.BlockPos)
	}{
		{
			name:    "start equals goal returns single element",
			isSolid: flatFloor(0, nil),
			start:   mcmath.BlockPos{X: 5, Y: 1, Z: 5},
			goal:    mcmath.BlockPos{X: 5, Y: 1, Z: 5},
			wantLen: 1,
			check: func(t *testing.T, path []mcmath.BlockPos) {
				assert.Equal(t, mcmath.BlockPos{X: 5, Y: 1, Z: 5}, path[0])
			},
		},
		{
			name:    "straight line path",
			isSolid: flatFloor(0, nil),
			start:   mcmath.BlockPos{X: 0, Y: 1, Z: 0},
			goal:    mcmath.BlockPos{X: 5, Y: 1, Z: 0},
			wantLen: 6, // 0,1,2,3,4,5 = 6 positions
			check: func(t *testing.T, path []mcmath.BlockPos) {
				assert.Equal(t, mcmath.BlockPos{X: 0, Y: 1, Z: 0}, path[0])
				assert.Equal(t, mcmath.BlockPos{X: 5, Y: 1, Z: 0}, path[len(path)-1])
			},
		},
		{
			name: "L-shape around wall",
			isSolid: flatFloor(0, map[mcmath.BlockPos]bool{
				// Wall from (2,1,0) to (2,1,3) blocking direct east path
				{X: 2, Y: 1, Z: 0}: true,
				{X: 2, Y: 1, Z: 1}: true,
				{X: 2, Y: 1, Z: 2}: true,
				{X: 2, Y: 1, Z: 3}: true,
			}),
			start:   mcmath.BlockPos{X: 0, Y: 1, Z: 0},
			goal:    mcmath.BlockPos{X: 4, Y: 1, Z: 0},
			wantNil: false,
			check: func(t *testing.T, path []mcmath.BlockPos) {
				assert.Equal(t, mcmath.BlockPos{X: 0, Y: 1, Z: 0}, path[0])
				assert.Equal(t, mcmath.BlockPos{X: 4, Y: 1, Z: 0}, path[len(path)-1])
				// Path must avoid wall at x=2, y=1
				for _, p := range path {
					if p.X == 2 && p.Y == 1 {
						t.Errorf("path goes through wall at %v", p)
					}
				}
			},
		},
		{
			name: "unreachable returns nil",
			isSolid: flatFloor(0, map[mcmath.BlockPos]bool{
				// Complete 2-high wall ring around the goal to prevent step-ups.
				{X: 4, Y: 1, Z: 0}: true, {X: 4, Y: 2, Z: 0}: true,
				{X: 4, Y: 1, Z: 1}: true, {X: 4, Y: 2, Z: 1}: true,
				{X: 4, Y: 1, Z: 2}: true, {X: 4, Y: 2, Z: 2}: true,
				{X: 5, Y: 1, Z: 0}: true, {X: 5, Y: 2, Z: 0}: true,
				{X: 5, Y: 1, Z: 2}: true, {X: 5, Y: 2, Z: 2}: true,
				{X: 6, Y: 1, Z: 0}: true, {X: 6, Y: 2, Z: 0}: true,
				{X: 6, Y: 1, Z: 1}: true, {X: 6, Y: 2, Z: 1}: true,
				{X: 6, Y: 1, Z: 2}: true, {X: 6, Y: 2, Z: 2}: true,
			}),
			start:   mcmath.BlockPos{X: 0, Y: 1, Z: 1},
			goal:    mcmath.BlockPos{X: 5, Y: 1, Z: 1},
			wantNil: true,
		},
		{
			name: "step up one block",
			isSolid: func() func(mcmath.BlockPos) bool {
				solid := map[mcmath.BlockPos]bool{
					// Floor at y=0
					{X: 0, Y: 0, Z: 0}: true,
					{X: 1, Y: 0, Z: 0}: true,
					{X: 2, Y: 0, Z: 0}: true,
					// Step: block at (2,1,0) making it a step up
					{X: 2, Y: 1, Z: 0}: true,
					{X: 3, Y: 1, Z: 0}: true,
				}
				return makeGrid(solid)
			}(),
			start: mcmath.BlockPos{X: 0, Y: 1, Z: 0},
			goal:  mcmath.BlockPos{X: 3, Y: 2, Z: 0},
			check: func(t *testing.T, path []mcmath.BlockPos) {
				assert.Equal(t, mcmath.BlockPos{X: 0, Y: 1, Z: 0}, path[0])
				assert.Equal(t, mcmath.BlockPos{X: 3, Y: 2, Z: 0}, path[len(path)-1])
				// Should step up at x=2
				foundStepUp := false
				for i := 1; i < len(path); i++ {
					if path[i].Y-path[i-1].Y == 1 {
						foundStepUp = true
					}
				}
				assert.True(t, foundStepUp, "path should include a step up")
			},
		},
		{
			name: "step down one block",
			isSolid: func() func(mcmath.BlockPos) bool {
				solid := map[mcmath.BlockPos]bool{
					// Higher floor on left
					{X: 0, Y: 1, Z: 0}: true,
					{X: 1, Y: 1, Z: 0}: true,
					// Lower floor on right
					{X: 2, Y: 0, Z: 0}: true,
					{X: 3, Y: 0, Z: 0}: true,
				}
				return makeGrid(solid)
			}(),
			start: mcmath.BlockPos{X: 0, Y: 2, Z: 0},
			goal:  mcmath.BlockPos{X: 3, Y: 1, Z: 0},
			check: func(t *testing.T, path []mcmath.BlockPos) {
				assert.Equal(t, mcmath.BlockPos{X: 0, Y: 2, Z: 0}, path[0])
				assert.Equal(t, mcmath.BlockPos{X: 3, Y: 1, Z: 0}, path[len(path)-1])
				// Should step down somewhere
				foundStepDown := false
				for i := 1; i < len(path); i++ {
					if path[i-1].Y-path[i].Y == 1 {
						foundStepDown = true
					}
				}
				assert.True(t, foundStepDown, "path should include a step down")
			},
		},
		{
			name:     "max nodes limit prevents finding distant path",
			isSolid:  flatFloor(0, nil),
			start:    mcmath.BlockPos{X: 0, Y: 1, Z: 0},
			goal:     mcmath.BlockPos{X: 100, Y: 1, Z: 100},
			maxNodes: 10,
			wantNil:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pf := &PathFinder{IsSolid: tc.isSolid}
			maxNodes := tc.maxNodes
			if maxNodes == 0 {
				maxNodes = 200
			}
			path := pf.FindPath(tc.start, tc.goal, maxNodes)

			if tc.wantNil {
				assert.Nil(t, path)
				return
			}

			require.NotNil(t, path, "expected a path but got nil")

			if tc.wantLen > 0 {
				assert.Len(t, path, tc.wantLen)
			}

			if tc.check != nil {
				tc.check(t, path)
			}

			// Verify path continuity: each step should be adjacent.
			for i := 1; i < len(path); i++ {
				dx := mcmath.Abs(int(path[i].X) - int(path[i-1].X))
				dy := mcmath.Abs(int(path[i].Y) - int(path[i-1].Y))
				dz := mcmath.Abs(int(path[i].Z) - int(path[i-1].Z))
				dist := dx + dy + dz
				assert.LessOrEqual(t, dist, 2, "path step %d->%d too far: %v -> %v", i-1, i, path[i-1], path[i])
			}
		})
	}
}
