package entity

import (
	"container/heap"
	"slices"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

// PathFinder performs A* pathfinding on a block grid.
type PathFinder struct {
	// IsSolid reports whether the block at the given position is solid.
	IsSolid func(mcmath.BlockPos) bool
}

// FindPath returns a slice of block positions from start to goal using A* search.
// maxNodes limits the number of nodes explored (capped at 200). Returns nil if
// no route is found. Returns [start] when start == goal.
func (pf *PathFinder) FindPath(start, goal mcmath.BlockPos, maxNodes int) []mcmath.BlockPos {
	if start == goal {
		return []mcmath.BlockPos{start}
	}
	if maxNodes <= 0 || maxNodes > 200 {
		maxNodes = 200
	}

	open := &nodeHeap{}
	heap.Init(open)

	startNode := &pathNode{
		pos:   start,
		g:     0,
		f:     manhattan(start, goal),
		index: 0,
	}
	heap.Push(open, startNode)

	visited := map[mcmath.BlockPos]*pathNode{start: startNode}
	explored := 0

	for open.Len() > 0 && explored < maxNodes {
		current := heap.Pop(open).(*pathNode)
		explored++

		if current.pos == goal {
			return reconstructPath(current)
		}

		for _, neighbor := range pf.neighbors(current.pos) {
			tentativeG := current.g + 1

			if existing, ok := visited[neighbor]; ok {
				// Skip nodes already popped from the open set.
				if existing.index < 0 || tentativeG >= existing.g {
					continue
				}
				existing.g = tentativeG
				existing.f = tentativeG + manhattan(neighbor, goal)
				existing.parent = current
				heap.Fix(open, existing.index)
			} else {
				n := &pathNode{
					pos:    neighbor,
					g:      tentativeG,
					f:      tentativeG + manhattan(neighbor, goal),
					parent: current,
				}
				heap.Push(open, n)
				visited[neighbor] = n
			}
		}
	}

	return nil
}

// cardinals holds the four horizontal directions (N, S, E, W) as (dx, dz) pairs.
var cardinals = [4][2]int32{
	{0, -1}, // North
	{0, 1},  // South
	{1, 0},  // East
	{-1, 0}, // West
}

// neighbors returns walkable neighbor positions from pos. A position is walkable
// when it is non-solid and the block below it is solid.
func (pf *PathFinder) neighbors(pos mcmath.BlockPos) []mcmath.BlockPos {
	var result []mcmath.BlockPos
	for _, dir := range cardinals {
		ahead := mcmath.BlockPos{X: pos.X + dir[0], Y: pos.Y, Z: pos.Z + dir[1]}

		// Flat walk: ahead is air and below-ahead is solid.
		if !pf.IsSolid(ahead) && pf.IsSolid(ahead.Below()) {
			result = append(result, ahead)
			continue
		}

		// Step up: ahead is solid, but above-ahead and above-current are both air.
		aboveAhead := ahead.Above()
		if pf.IsSolid(ahead) && !pf.IsSolid(aboveAhead) && !pf.IsSolid(pos.Above()) {
			result = append(result, aboveAhead)
			continue
		}

		// Step down: ahead is air and below-ahead is also air, but two below pos in the direction is solid.
		belowAhead := ahead.Below()
		if !pf.IsSolid(ahead) && !pf.IsSolid(belowAhead) && pf.IsSolid(belowAhead.Below()) {
			result = append(result, belowAhead)
			continue
		}
	}
	return result
}

// manhattan returns the Manhattan distance between two block positions.
func manhattan(a, b mcmath.BlockPos) int {
	return abs(int(a.X)-int(b.X)) + abs(int(a.Y)-int(b.Y)) + abs(int(a.Z)-int(b.Z))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// reconstructPath walks back from goal to start via parent pointers and returns
// the path in start-to-goal order.
func reconstructPath(node *pathNode) []mcmath.BlockPos {
	var path []mcmath.BlockPos
	for n := node; n != nil; n = n.parent {
		path = append(path, n.pos)
	}
	slices.Reverse(path)
	return path
}

// pathNode represents a node in the A* open set.
type pathNode struct {
	pos    mcmath.BlockPos
	g      int // cost from start
	f      int // g + heuristic
	parent *pathNode
	index  int // index in the heap
}

// nodeHeap is a min-heap of pathNodes ordered by f score.
type nodeHeap []*pathNode

func (h nodeHeap) Len() int           { return len(h) }
func (h nodeHeap) Less(i, j int) bool { return h[i].f < h[j].f }
func (h nodeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *nodeHeap) Push(x any) {
	n := x.(*pathNode)
	n.index = len(*h)
	*h = append(*h, n)
}

func (h *nodeHeap) Pop() any {
	old := *h
	n := len(old)
	node := old[n-1]
	old[n-1] = nil // avoid memory leak
	node.index = -1
	*h = old[:n-1]
	return node
}
