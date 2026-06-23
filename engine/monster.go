package engine

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
)

type MonsterState int

const (
	MonsterIdle MonsterState = iota
	MonsterWander
	MonsterPursue
	MonsterAttack
	MonsterFlee
)

type MonsterType struct {
	Name       string `json:"name"`
	HP         int    `json:"hp"`
	AC         int    `json:"ac"`
	Sprite     string `json:"sprite"`
	DetectDist int    `json:"detectDist"`
	Speed      int    `json:"speed"`
}

var MonsterTypes map[string]MonsterType

func LoadMonsterTypes(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("load monster types: %v", err)
	}
	var types []MonsterType
	if err := json.Unmarshal(data, &types); err != nil {
		log.Fatalf("parse monster types: %v", err)
	}
	MonsterTypes = make(map[string]MonsterType, len(types))
	for _, t := range types {
		MonsterTypes[t.Name] = t
	}
}

type Monster struct {
	Type      MonsterType
	X, Y      int
	HP        int
	State     MonsterState
	Facing    Direction
	MoveTicks int
	MoveTimer int
	AlertedBy *Monster
}

func NewMonster(mt MonsterType, x, y int) *Monster {
	return &Monster{
		Type:      mt,
		X:         x,
		Y:         y,
		HP:        mt.HP,
		State:     MonsterWander,
		Facing:    Direction(rand.Intn(4)),
		MoveTicks: mt.Speed + rand.Intn(30),
	}
}

func (m *Monster) distTo(x, y int) int {
	return abs(m.X-x) + abs(m.Y-y)
}

func (m *Monster) adjacentTo(x, y int) bool {
	return m.distTo(x, y) == 1
}

func (m *Monster) canDetect(w *World) bool {
	dist := m.distTo(w.PartyX, w.PartyY)
	if dist > m.Type.DetectDist {
		return false
	}
	return hasLineOfSight(w, m.X, m.Y, w.PartyX, w.PartyY)
}

func hasLineOfSight(w *World, x0, y0, x1, y1 int) bool {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	cx, cy := x0, y0
	for {
		if cx == x1 && cy == y1 {
			return true
		}
		if (cx != x0 || cy != y0) && !w.IsPassable(cx, cy) {
			return false
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			cx += sx
		}
		if e2 < dx {
			err += dx
			cy += sy
		}
	}
}

func (m *Monster) Update(w *World) {
	m.MoveTimer++
	speed := m.moveTicks()
	if m.MoveTimer < speed {
		return
	}
	m.MoveTimer = 0

	switch m.State {
	case MonsterIdle:
		m.idle(w)
	case MonsterWander:
		m.wander(w)
	case MonsterPursue:
		m.pursue(w)
	case MonsterAttack:
		m.attack(w)
	}
}

func (m *Monster) moveTicks() int {
	switch m.State {
	case MonsterPursue:
		return m.Type.Speed
	case MonsterAttack:
		return m.Type.Speed + 20
	default:
		return m.Type.Speed + rand.Intn(40)
	}
}

func (m *Monster) idle(w *World) {
	if m.canDetect(w) {
		m.alertNearby(w)
		m.State = MonsterPursue
		return
	}
	if rand.Intn(4) == 0 {
		m.State = MonsterWander
	}
}

func (m *Monster) wander(w *World) {
	if m.canDetect(w) {
		m.alertNearby(w)
		m.State = MonsterPursue
		return
	}

	if rand.Intn(3) == 0 {
		m.Facing = Direction(rand.Intn(4))
	}

	dx, dy := m.Facing.Delta()
	nx, ny := m.X+dx, m.Y+dy
	if w.IsPassable(nx, ny) && !w.IsOccupiedExcept(nx, ny, m) {
		m.X = nx
		m.Y = ny
	} else {
		m.Facing = Direction(rand.Intn(4))
	}
}

func (m *Monster) pursue(w *World) {
	if m.adjacentTo(w.PartyX, w.PartyY) {
		m.State = MonsterAttack
		return
	}

	if !m.canDetect(w) && m.distTo(w.PartyX, w.PartyY) > m.Type.DetectDist+2 {
		m.State = MonsterWander
		return
	}

	nx, ny, found := FindPath(w, m.X, m.Y, w.PartyX, w.PartyY, m)
	if found && (nx != m.X || ny != m.Y) {
		if !w.IsOccupiedExcept(nx, ny, m) {
			m.X = nx
			m.Y = ny
			return
		}
		// Path blocked by another monster — try flanking
		m.flank(w)
		return
	}

	// No path found, try flanking
	m.flank(w)
}

func (m *Monster) flank(w *World) {
	// Try approaching from a perpendicular direction
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	// Shuffle to avoid all monsters picking the same flank
	for i := len(dirs) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	best := -1
	bestDist := m.distTo(w.PartyX, w.PartyY)
	for i, d := range dirs {
		nx, ny := m.X+d[0], m.Y+d[1]
		if !w.IsPassable(nx, ny) || w.IsOccupiedExcept(nx, ny, m) {
			continue
		}
		dist := abs(nx-w.PartyX) + abs(ny-w.PartyY)
		if dist < bestDist {
			bestDist = dist
			best = i
		}
	}

	if best >= 0 {
		m.X += dirs[best][0]
		m.Y += dirs[best][1]
	}
}

func (m *Monster) attack(w *World) {
	if !m.adjacentTo(w.PartyX, w.PartyY) {
		m.State = MonsterPursue
		return
	}
	// TODO: deal damage to party when combat is wired up
}

func (m *Monster) alertNearby(w *World) {
	for _, other := range w.Monsters {
		if other == m || other.State == MonsterPursue || other.State == MonsterAttack {
			continue
		}
		if other.distTo(m.X, m.Y) <= 4 {
			other.State = MonsterPursue
			other.AlertedBy = m
		}
	}
}
