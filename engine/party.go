package engine

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Party struct {
	X      int
	Y      int
	Facing Direction
	World  *World
}

func NewParty(world *World, x, y int, facing Direction) *Party {
	world.PartyX = x
	world.PartyY = y
	return &Party{
		X:      x,
		Y:      y,
		Facing: facing,
		World:  world,
	}
}

func (p *Party) moveTo(nx, ny int) {
	p.X, p.Y = nx, ny
	p.World.PartyX = nx
	p.World.PartyY = ny
}

func (p *Party) MoveForward() bool {
	dx, dy := p.Facing.Delta()
	nx, ny := p.X+dx, p.Y+dy
	if p.World.HasWall(p.X, p.Y, p.Facing) {
		return false
	}
	if !p.World.IsPassable(nx, ny) || p.World.IsOccupied(nx, ny) {
		return false
	}
	p.moveTo(nx, ny)
	return true
}

func (p *Party) MoveBackward() bool {
	back := p.Facing.Opposite()
	dx, dy := back.Delta()
	nx, ny := p.X+dx, p.Y+dy
	if p.World.HasWall(p.X, p.Y, back) {
		return false
	}
	if !p.World.IsPassable(nx, ny) || p.World.IsOccupied(nx, ny) {
		return false
	}
	p.moveTo(nx, ny)
	return true
}

func (p *Party) StrafeLeft() bool {
	left := p.Facing.TurnLeft()
	dx, dy := left.Delta()
	nx, ny := p.X+dx, p.Y+dy
	if p.World.HasWall(p.X, p.Y, left) {
		return false
	}
	if !p.World.IsPassable(nx, ny) || p.World.IsOccupied(nx, ny) {
		return false
	}
	p.moveTo(nx, ny)
	return true
}

func (p *Party) StrafeRight() bool {
	right := p.Facing.TurnRight()
	dx, dy := right.Delta()
	nx, ny := p.X+dx, p.Y+dy
	if p.World.HasWall(p.X, p.Y, right) {
		return false
	}
	if !p.World.IsPassable(nx, ny) || p.World.IsOccupied(nx, ny) {
		return false
	}
	p.moveTo(nx, ny)
	return true
}

func (p *Party) TurnLeft() {
	p.Facing = p.Facing.TurnLeft()
}

func (p *Party) TurnRight() {
	p.Facing = p.Facing.TurnRight()
}

type InputState struct {
	cooldown int
}

const inputCooldownTicks = 10

func (s *InputState) HandleInput(party *Party) {
	// Interact fires once per keypress, independent of the movement cooldown.
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		party.Interact()
	}

	if s.cooldown > 0 {
		s.cooldown--
		return
	}

	moved := false

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		moved = party.MoveForward()
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		moved = party.MoveBackward()
	} else if ebiten.IsKeyPressed(ebiten.KeyQ) {
		moved = party.StrafeLeft()
	} else if ebiten.IsKeyPressed(ebiten.KeyE) {
		moved = party.StrafeRight()
	} else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		party.TurnLeft()
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		party.TurnRight()
		moved = true
	}

	if moved {
		s.cooldown = inputCooldownTicks
	}
}

func (p *Party) Interact() bool {
	dx, dy := p.Facing.Delta()
	tx, ty := p.X+dx, p.Y+dy
	if p.World.ToggleDoor(tx, ty) {
		return true
	}
	return p.Attack()
}

func (p *Party) Attack() bool {
	dx, dy := p.Facing.Delta()
	tx, ty := p.X+dx, p.Y+dy
	monsters := p.World.MonstersAt(tx, ty)
	if len(monsters) == 0 {
		return false
	}
	m := monsters[0]
	m.HP -= 2 + rand.Intn(4) // 2-5 damage placeholder
	if m.HP <= 0 {
		p.World.RemoveMonster(m)
	}
	return true
}
