package engine

import (
	"encoding/json"
	"fmt"
	"os"
)

type LevelData struct {
	Name       string          `json:"name"`
	Width      int             `json:"width"`
	Height     int             `json:"height"`
	Wallset    string          `json:"wallset"`
	PartyStart partyStartData  `json:"partyStart"`
	Tiles      []string        `json:"tiles"`
	Monsters   []monsterPlace  `json:"monsters"`
}

type partyStartData struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Facing string `json:"facing"`
}

type monsterPlace struct {
	Type string `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

func LoadLevel(path string) (*World, int, int, Direction) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("load level %s: %v", path, err))
	}
	var ld LevelData
	if err := json.Unmarshal(data, &ld); err != nil {
		panic(fmt.Sprintf("parse level %s: %v", path, err))
	}

	w := NewWorld(ld.Width, ld.Height)
	w.Wallset = ld.Wallset

	for y, row := range ld.Tiles {
		for x, ch := range row {
			if x >= ld.Width || y >= ld.Height {
				continue
			}
			switch ch {
			case '.':
				w.SetFloor(x, y)
			case 'D':
				w.Tiles[y][x] = Tile{Type: TileDoor, Walls: [4]bool{false, false, false, false}}
			case '|':
				w.Tiles[y][x] = Tile{Type: TileDoorFrame, Walls: [4]bool{true, true, true, true}}
			case '<':
				w.Tiles[y][x] = Tile{Type: TileStairsUp, Walls: [4]bool{false, false, false, false}}
			case '>':
				w.Tiles[y][x] = Tile{Type: TileStairsDown, Walls: [4]bool{false, false, false, false}}
			}
		}
	}

	for _, mp := range ld.Monsters {
		mt, ok := MonsterTypes[mp.Type]
		if !ok {
			panic(fmt.Sprintf("unknown monster type %q in level %s", mp.Type, path))
		}
		w.Monsters = append(w.Monsters, NewMonster(mt, mp.X, mp.Y))
	}

	facing := parseFacing(ld.PartyStart.Facing)
	return w, ld.PartyStart.X, ld.PartyStart.Y, facing
}

func parseFacing(s string) Direction {
	switch s {
	case "N":
		return North
	case "E":
		return East
	case "S":
		return South
	case "W":
		return West
	default:
		return North
	}
}
