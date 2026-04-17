package inventory

import (
	"github.com/fanxiyao/gomc/internal/item"
)

// CraftingGrid represents a 3x3 crafting grid.
// A 2x2 crafting table uses only the top-left portion.
type CraftingGrid struct {
	Grid [3][3]item.ItemStack
}

// SetSlot places the given stack at the specified grid position.
// Does nothing if row or col is out of bounds.
func (g *CraftingGrid) SetSlot(row, col int, stack item.ItemStack) {
	if row < 0 || row >= 3 || col < 0 || col >= 3 {
		return
	}
	g.Grid[row][col] = stack
}

// GetSlot returns the stack at the specified grid position.
// Returns an empty ItemStack if out of bounds.
func (g *CraftingGrid) GetSlot(row, col int) item.ItemStack {
	if row < 0 || row >= 3 || col < 0 || col >= 3 {
		return item.ItemStack{}
	}
	return g.Grid[row][col]
}

// Clear empties all grid slots.
func (g *CraftingGrid) Clear() {
	g.Grid = [3][3]item.ItemStack{}
}

// GetResult checks the grid against all registered recipes and returns the
// result stack if a match is found. Returns an empty stack if no recipe matches.
func (g *CraftingGrid) GetResult() item.ItemStack {
	gridIDs := g.extractIDs()

	for _, recipe := range recipes {
		if recipe.Shapeless {
			if matchShapeless(gridIDs, recipe.Pattern) {
				return recipe.Result
			}
		} else {
			if matchShaped(gridIDs, recipe.Pattern) {
				return recipe.Result
			}
		}
	}
	return item.ItemStack{}
}

// Craft attempts to craft the current recipe. If a recipe matches, it
// decrements each non-empty grid slot by 1 and returns the result.
// Returns (result, true) on success or (empty, false) on failure.
func (g *CraftingGrid) Craft() (item.ItemStack, bool) {
	result := g.GetResult()
	if result.IsEmpty() {
		return item.ItemStack{}, false
	}

	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if !g.Grid[r][c].IsEmpty() {
				g.Grid[r][c].Count--
				if g.Grid[r][c].Count <= 0 {
					g.Grid[r][c] = item.ItemStack{}
				}
			}
		}
	}

	return result, true
}

// extractIDs returns a 3x3 grid of item IDs (0 for empty slots).
func (g *CraftingGrid) extractIDs() [3][3]uint16 {
	var ids [3][3]uint16
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if !g.Grid[r][c].IsEmpty() {
				ids[r][c] = g.Grid[r][c].ItemID
			}
		}
	}
	return ids
}

// matchShaped checks whether the grid matches a shaped recipe pattern
// by trying all possible offsets.
func matchShaped(gridIDs, pattern [3][3]uint16) bool {
	// Determine the bounding box of the pattern.
	pMinR, pMinC, pMaxR, pMaxC := patternBounds(pattern)
	if pMinR > pMaxR {
		// Empty pattern matches only an empty grid.
		gMinR, _, _, _ := patternBounds(gridIDs)
		return gMinR > 2
	}

	pH := pMaxR - pMinR + 1
	pW := pMaxC - pMinC + 1

	// Determine the bounding box of the grid.
	gMinR, gMinC, gMaxR, gMaxC := patternBounds(gridIDs)
	if gMinR > gMaxR {
		return false // grid is empty, pattern is not
	}

	gH := gMaxR - gMinR + 1
	gW := gMaxC - gMinC + 1

	if gH != pH || gW != pW {
		return false
	}

	// Compare pattern contents offset to grid contents.
	for dr := 0; dr < pH; dr++ {
		for dc := 0; dc < pW; dc++ {
			if pattern[pMinR+dr][pMinC+dc] != gridIDs[gMinR+dr][gMinC+dc] {
				return false
			}
		}
	}
	return true
}

// matchShapeless checks whether the grid contains exactly the same item IDs
// as the pattern, regardless of positions.
func matchShapeless(gridIDs, pattern [3][3]uint16) bool {
	gridCounts := make(map[uint16]int)
	patCounts := make(map[uint16]int)

	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if gridIDs[r][c] != 0 {
				gridCounts[gridIDs[r][c]]++
			}
			if pattern[r][c] != 0 {
				patCounts[pattern[r][c]]++
			}
		}
	}

	if len(gridCounts) != len(patCounts) {
		return false
	}
	for id, count := range patCounts {
		if gridCounts[id] != count {
			return false
		}
	}
	return true
}

// patternBounds returns the bounding box (minRow, minCol, maxRow, maxCol)
// of non-zero entries. If all entries are zero, returns (3, 3, -1, -1).
func patternBounds(p [3][3]uint16) (int, int, int, int) {
	minR, minC := 3, 3
	maxR, maxC := -1, -1
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if p[r][c] != 0 {
				if r < minR {
					minR = r
				}
				if r > maxR {
					maxR = r
				}
				if c < minC {
					minC = c
				}
				if c > maxC {
					maxC = c
				}
			}
		}
	}
	return minR, minC, maxR, maxC
}
