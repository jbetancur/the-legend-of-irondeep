package engine

type TileType int

const (
	TileFloor TileType = iota
	TileWall
	TileDoor
	TileDoorOpen
	TileStairsUp
	TileStairsDown
	TilePit
)

type Tile struct {
	Type  TileType
	Walls [4]bool // N, E, S, W — true means wall present on that side
}

type World struct {
	Width    int
	Height   int
	Tiles    [][]Tile
	Wallset  string
	Monsters []*Monster
	PartyX   int
	PartyY   int
}

func NewWorld(width, height int) *World {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x] = Tile{Type: TileWall, Walls: [4]bool{true, true, true, true}}
		}
	}
	return &World{Width: width, Height: height, Tiles: tiles, Wallset: "stone"}
}

func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.Width && y >= 0 && y < w.Height
}

func (w *World) IsPassable(x, y int) bool {
	if !w.InBounds(x, y) {
		return false
	}
	t := w.Tiles[y][x].Type
	return t == TileFloor || t == TileDoorOpen || t == TileStairsUp || t == TileStairsDown
}

func (w *World) IsOccupied(x, y int) bool {
	if w.PartyX == x && w.PartyY == y {
		return true
	}
	for _, m := range w.Monsters {
		if m.X == x && m.Y == y {
			return true
		}
	}
	return false
}

func (w *World) IsOccupiedExcept(x, y int, exclude *Monster) bool {
	if w.PartyX == x && w.PartyY == y {
		return true
	}
	for _, m := range w.Monsters {
		if m != exclude && m.X == x && m.Y == y {
			return true
		}
	}
	return false
}

func (w *World) MonstersAt(x, y int) []*Monster {
	var result []*Monster
	for _, m := range w.Monsters {
		if m.X == x && m.Y == y {
			result = append(result, m)
		}
	}
	return result
}

func (w *World) UpdateMonsters() {
	for _, m := range w.Monsters {
		m.Update(w)
	}
}

// IsWall reports whether the cell is a solid wall block (or out of bounds).
func (w *World) IsWall(x, y int) bool {
	if !w.InBounds(x, y) {
		return true
	}
	return w.Tiles[y][x].Type == TileWall
}

// IsDoor reports whether the cell holds a door (open or closed).
func (w *World) IsDoor(x, y int) bool {
	if !w.InBounds(x, y) {
		return false
	}
	t := w.Tiles[y][x].Type
	return t == TileDoor || t == TileDoorOpen
}

func (w *World) HasWall(x, y int, dir Direction) bool {
	if !w.InBounds(x, y) {
		return true
	}
	return w.Tiles[y][x].Walls[dir]
}

func (w *World) SetFloor(x, y int) {
	if !w.InBounds(x, y) {
		return
	}
	w.Tiles[y][x].Type = TileFloor
	w.Tiles[y][x].Walls = [4]bool{false, false, false, false}
}

func (w *World) SetWallBetween(x1, y1, x2, y2 int) {
	dx := x2 - x1
	dy := y2 - y1

	var dir Direction
	switch {
	case dy == -1:
		dir = North
	case dx == 1:
		dir = East
	case dy == 1:
		dir = South
	case dx == -1:
		dir = West
	default:
		return
	}

	if w.InBounds(x1, y1) {
		w.Tiles[y1][x1].Walls[dir] = true
	}
	if w.InBounds(x2, y2) {
		w.Tiles[y2][x2].Walls[dir.Opposite()] = true
	}
}

func BuildTestLevel() *World {
	w := NewWorld(16, 16)

	// Carve out a small dungeon
	// Entry corridor going north
	for y := 12; y >= 4; y-- {
		w.SetFloor(4, y)
	}

	// East-west corridor
	for x := 2; x <= 10; x++ {
		w.SetFloor(x, 8)
	}

	// Side room (east)
	for y := 6; y <= 10; y++ {
		for x := 8; x <= 10; x++ {
			w.SetFloor(x, y)
		}
	}

	// Side room (west)
	for y := 6; y <= 8; y++ {
		for x := 2; x <= 3; x++ {
			w.SetFloor(x, y)
		}
	}

	// North chamber
	for y := 3; y <= 5; y++ {
		for x := 3; x <= 6; x++ {
			w.SetFloor(x, y)
		}
	}

	// Dead-end alcove
	w.SetFloor(4, 2)

	w.Monsters = []*Monster{
		NewMonster(Skeleton, 4, 9),
		NewMonster(Skeleton, 4, 6),
		NewMonster(Skeleton, 9, 8),
	}

	return w
}
