package engine

type Direction int

const (
	North Direction = iota
	East
	South
	West
)

func (d Direction) TurnRight() Direction {
	return (d + 1) % 4
}

func (d Direction) TurnLeft() Direction {
	return (d + 3) % 4
}

func (d Direction) Opposite() Direction {
	return (d + 2) % 4
}

func (d Direction) Delta() (dx, dy int) {
	switch d {
	case North:
		return 0, -1
	case East:
		return 1, 0
	case South:
		return 0, 1
	case West:
		return -1, 0
	}
	return 0, 0
}

func (d Direction) String() string {
	switch d {
	case North:
		return "N"
	case East:
		return "E"
	case South:
		return "S"
	case West:
		return "W"
	}
	return "?"
}
