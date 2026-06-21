package engine

import (
	"math/rand"
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
	Name   string
	HP     int
	AC     int
	Sprite string
}

var Skeleton = MonsterType{
	Name:   "Skeleton",
	HP:     8,
	AC:     7,
	Sprite: "skeleton",
}

type Monster struct {
	Type      MonsterType
	X, Y      int
	HP        int
	State     MonsterState
	Facing    Direction
	MoveTicks int
	MoveTimer int
}

func NewMonster(mt MonsterType, x, y int) *Monster {
	return &Monster{
		Type:      mt,
		X:         x,
		Y:         y,
		HP:        mt.HP,
		State:     MonsterWander,
		Facing:    Direction(rand.Intn(4)),
		MoveTicks: 60 + rand.Intn(60),
	}
}

func (m *Monster) Update(w *World) {
	m.MoveTimer++
	if m.MoveTimer < m.MoveTicks {
		return
	}
	m.MoveTimer = 0
	m.MoveTicks = 60 + rand.Intn(60)

	switch m.State {
	case MonsterWander:
		m.wander(w)
	}
}

func (m *Monster) wander(w *World) {
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
