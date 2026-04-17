package mcmath

import "fmt"

// Direction represents a cardinal direction in 3D space.
type Direction int

const (
	North Direction = iota // -Z
	South                  // +Z
	East                   // +X
	West                   // -X
	Up                     // +Y
	Down                   // -Y
)

var directionNames = [6]string{"North", "South", "East", "West", "Up", "Down"}

var directionNormals = [6]Vec3{
	{0, 0, -1}, // North
	{0, 0, 1},  // South
	{1, 0, 0},  // East
	{-1, 0, 0}, // West
	{0, 1, 0},  // Up
	{0, -1, 0}, // Down
}

var directionOpposites = [6]Direction{
	South, // North -> South
	North, // South -> North
	West,  // East  -> West
	East,  // West  -> East
	Down,  // Up    -> Down
	Up,    // Down  -> Up
}

// Opposite returns the opposite direction.
func (d Direction) Opposite() Direction {
	return directionOpposites[d]
}

// Normal returns the unit vector for this direction.
func (d Direction) Normal() Vec3 {
	return directionNormals[d]
}

// String returns the name of the direction.
func (d Direction) String() string {
	if d < North || d > Down {
		return fmt.Sprintf("Direction(%d)", int(d))
	}
	return directionNames[d]
}

// All returns a slice of all six directions.
func All() []Direction {
	return []Direction{North, South, East, West, Up, Down}
}
