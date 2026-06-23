package engine

type TileType int

const (
	TileFloor TileType = iota
	TileWall
	TileDoor
	TileDoorOpen
	TileDoorFrame
	TileStairsUp
	TileStairsDown
	TilePit
)

type Tile struct {
	Type  TileType
	Walls [4]bool // N, E, S, W — true means wall present on that side
}

type DoorAnim struct {
	Opening  bool
	Progress float64 // 0.0 = closed, 1.0 = fully open
}

type World struct {
	Width    int
	Height   int
	Tiles    [][]Tile
	Wallset  string
	Monsters []*Monster
	Doors    map[[2]int]*DoorAnim
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
	return &World{Width: width, Height: height, Tiles: tiles, Wallset: "stone", Doors: make(map[[2]int]*DoorAnim)}
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

func (w *World) RemoveMonster(m *Monster) {
	for i, mon := range w.Monsters {
		if mon == m {
			w.Monsters = append(w.Monsters[:i], w.Monsters[i+1:]...)
			return
		}
	}
}

func (w *World) UpdateMonsters() {
	for _, m := range w.Monsters {
		m.Update(w)
	}
}

func (w *World) IsWall(x, y int) bool {
	if !w.InBounds(x, y) {
		return true
	}
	t := w.Tiles[y][x].Type
	return t == TileWall || t == TileDoorFrame
}

func (w *World) IsDoorFrame(x, y int) bool {
	if !w.InBounds(x, y) {
		return false
	}
	return w.Tiles[y][x].Type == TileDoorFrame
}

func (w *World) IsDoor(x, y int) bool {
	if !w.InBounds(x, y) {
		return false
	}
	t := w.Tiles[y][x].Type
	return t == TileDoor || t == TileDoorOpen
}

func (w *World) IsClosedDoor(x, y int) bool {
	if !w.InBounds(x, y) {
		return false
	}
	return w.Tiles[y][x].Type == TileDoor
}

func (w *World) ToggleDoor(x, y int) bool {
	if !w.InBounds(x, y) {
		return false
	}
	t := w.Tiles[y][x].Type
	if t != TileDoor && t != TileDoorOpen {
		return false
	}
	key := [2]int{x, y}
	da, exists := w.Doors[key]

	// Decide the new direction: reverse if animating, otherwise from settled state.
	willOpen := t == TileDoor
	if exists {
		willOpen = !da.Opening
	}

	// Can't close a door onto an occupant standing in the doorway.
	if !willOpen && w.IsOccupied(x, y) {
		return false
	}

	if !exists {
		w.Doors[key] = &DoorAnim{Opening: willOpen}
	} else {
		da.Opening = willOpen
	}
	return true
}

func (w *World) DoorProgress(x, y int) float64 {
	if da, ok := w.Doors[[2]int{x, y}]; ok {
		return da.Progress
	}
	// No active animation: progress is implied by the settled tile state.
	if w.InBounds(x, y) && w.Tiles[y][x].Type == TileDoorOpen {
		return 1.0
	}
	return 0
}

const doorSpeed = 0.04

func (w *World) UpdateDoors() {
	for pos, da := range w.Doors {
		if da.Opening {
			da.Progress += doorSpeed
			if da.Progress >= 1.0 {
				da.Progress = 1.0
				w.Tiles[pos[1]][pos[0]].Type = TileDoorOpen
				delete(w.Doors, pos)
			}
		} else {
			da.Progress -= doorSpeed
			if da.Progress <= 0 {
				da.Progress = 0
				w.Tiles[pos[1]][pos[0]].Type = TileDoor
				delete(w.Doors, pos)
			}
		}
	}
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

