package model

// Direction is a cardinal direction with respect to the map.
type Direction int

const (
	DirectionNone Direction = iota
	DirectionNorthEast
	DirectionNorth
	DirectionNorthWest
	DirectionWest
	DirectionSouthWest
	DirectionSouth
	DirectionSouthEast
	DirectionEast
)

// DirectionFromDelta returns a direction represented by a difference in x and/or y coordinates. The vector v does not
// need to be a unit vector.
func DirectionFromDelta(v Vector2D) Direction {
	if v.X == 0 && v.Y == 0 {
		return DirectionNone
	} else if v.X < 0 && v.Y == 0 {
		return DirectionWest
	} else if v.X > 0 && v.Y == 0 {
		return DirectionEast
	} else if v.X == 0 && v.Y < 0 {
		return DirectionSouth
	} else if v.X == 0 && v.Y > 0 {
		return DirectionNorth
	} else if v.X < 0 && v.Y < 0 {
		return DirectionSouthWest
	} else if v.X < 0 && v.Y > 0 {
		return DirectionNorthWest
	} else if v.X > 0 && v.Y < 0 {
		return DirectionSouthEast
	} else {
		return DirectionNorthEast
	}
}
